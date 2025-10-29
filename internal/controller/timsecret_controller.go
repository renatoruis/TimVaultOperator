package controller

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sort"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"

	secretsv1alpha1 "github.com/renatoruis/timvault-operator/api/v1alpha1"
	"github.com/renatoruis/timvault-operator/internal/vault"
)

// TimSecretReconciler reconciles a TimSecret object
type TimSecretReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=secrets.tim.operator,resources=timsecrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=secrets.tim.operator,resources=timsecrets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=secrets.tim.operator,resources=timsecrets/finalizers,verbs=update
// +kubebuilder:rbac:groups=secrets.tim.operator,resources=timsecretconfigs,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop
func (r *TimSecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the TimSecret instance
	timSecret := &secretsv1alpha1.TimSecret{}
	err := r.Get(ctx, req.NamespacedName, timSecret)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("TimSecret resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get TimSecret")
		return ctrl.Result{}, err
	}

	// Parse sync interval (default 5 minutes)
	syncInterval := r.parseSyncInterval(timSecret.Spec.SyncInterval)

	// Resolve Vault configuration
	vaultURL, vaultToken, err := r.resolveVaultConfig(ctx, timSecret)
	if err != nil {
		logger.Error(err, "Failed to resolve Vault configuration")
		return r.handleError(ctx, timSecret, syncInterval, err, "VaultConfigResolutionFailed")
	}

	// Create Vault client
	vaultClient, err := vault.NewClient(vaultURL, vaultToken)
	if err != nil {
		logger.Error(err, "Failed to create Vault client")
		return r.handleError(ctx, timSecret, syncInterval, err, "VaultClientCreationFailed")
	}

	// Get secrets from Vault
	secretData, err := vaultClient.GetSecrets(ctx, timSecret.Spec.VaultPath)
	if err != nil {
		logger.Error(err, "Failed to get secrets from Vault")
		return r.handleError(ctx, timSecret, syncInterval, err, "VaultSecretFetchFailed")
	}

	// Calculate hash of secret data
	newHash := calculateHash(secretData)
	oldHash := timSecret.Status.SecretHash

	// Log hash comparison for debugging
	logger.V(1).Info("Hash comparison",
		"oldHash", oldHash,
		"newHash", newHash,
		"changed", oldHash != newHash)

	// Determine namespace
	namespace := timSecret.Spec.Namespace
	if namespace == "" {
		namespace = timSecret.Namespace
	}

	// Create or update Kubernetes Secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      timSecret.Spec.SecretName,
			Namespace: namespace,
		},
	}

	secretExists := true
	err = r.Get(ctx, types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, secret)
	if err != nil {
		if errors.IsNotFound(err) {
			secretExists = false
		} else {
			logger.Error(err, "Failed to get Secret")
			return ctrl.Result{}, err
		}
	}

	// Check if secret data has changed
	secretChanged := timSecret.Status.SecretHash != newHash

	// Convert map[string]string to map[string][]byte
	secretDataBytes := make(map[string][]byte)
	for k, v := range secretData {
		secretDataBytes[k] = []byte(v)
	}

	// Create or update Secret only if it doesn't exist or data changed
	if !secretExists {
		// Create new secret
		secret.Data = secretDataBytes
		secret.Type = corev1.SecretTypeOpaque

		if err := ctrl.SetControllerReference(timSecret, secret, r.Scheme); err != nil {
			logger.Error(err, "Failed to set controller reference")
			return ctrl.Result{}, err
		}
		if err := r.Create(ctx, secret); err != nil {
			logger.Error(err, "Failed to create Secret")
			return ctrl.Result{}, err
		}
		logger.Info("Created Secret", "name", secret.Name, "namespace", secret.Namespace)
	} else if secretChanged {
		// Update existing secret ONLY if data changed
		secret.Data = secretDataBytes
		secret.Type = corev1.SecretTypeOpaque

		if err := r.Update(ctx, secret); err != nil {
			logger.Error(err, "Failed to update Secret")
			return ctrl.Result{}, err
		}
		logger.Info("Updated Secret", "name", secret.Name, "namespace", secret.Namespace)
	} else {
		logger.Info("Secret data unchanged, skipping update", "name", secret.Name, "namespace", secret.Namespace)
	}

	// Restart deployment if secret changed and deployment name is specified
	if secretChanged && timSecret.Spec.DeploymentName != "" {
		if err := r.restartDeployment(ctx, timSecret.Spec.DeploymentName, namespace); err != nil {
			logger.Error(err, "Failed to restart Deployment")
			return ctrl.Result{}, err
		}
		logger.Info("Restarted Deployment", "name", timSecret.Spec.DeploymentName, "namespace", namespace)
	}

	// Update status - success, reset retry count
	now := metav1.Now()
	timSecret.Status.LastSyncTime = &now
	timSecret.Status.SecretHash = newHash
	timSecret.Status.RetryCount = 0 // Reset on success
	timSecret.Status.LastError = "" // Clear error
	timSecret.Status.Conditions = []metav1.Condition{
		{
			Type:               "Ready",
			Status:             metav1.ConditionTrue,
			LastTransitionTime: now,
			Reason:             "SecretSynced",
			Message:            "Secret successfully synced from Vault",
		},
	}

	if err := r.Status().Update(ctx, timSecret); err != nil {
		logger.Error(err, "Failed to update TimSecret status")
		return ctrl.Result{}, err
	}

	// Requeue after configured sync interval
	logger.Info("Secret synced successfully, requeueing", "syncInterval", syncInterval)
	return ctrl.Result{RequeueAfter: syncInterval}, nil
}

// restartDeployment restarts a deployment by updating its annotation
func (r *TimSecretReconciler) restartDeployment(ctx context.Context, name, namespace string) error {
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, deployment)
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}

	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = make(map[string]string)
	}

	// Update annotation to trigger restart
	deployment.Spec.Template.Annotations["secrets.tim.operator/restartedAt"] = time.Now().Format(time.RFC3339)

	if err := r.Update(ctx, deployment); err != nil {
		return fmt.Errorf("failed to update deployment: %w", err)
	}

	return nil
}

// resolveVaultConfig resolves Vault configuration from TimSecretConfig or direct values
func (r *TimSecretReconciler) resolveVaultConfig(ctx context.Context, ts *secretsv1alpha1.TimSecret) (string, string, error) {
	// Priority: direct values > TimSecretConfig
	if ts.Spec.VaultURL != "" && ts.Spec.VaultToken != "" {
		return ts.Spec.VaultURL, ts.Spec.VaultToken, nil
	}

	// Try to get from TimSecretConfig
	if ts.Spec.VaultConfig != "" {
		namespace := ts.Spec.VaultConfigNamespace
		if namespace == "" {
			namespace = ts.Namespace
		}

		config := &secretsv1alpha1.TimSecretConfig{}
		err := r.Get(ctx, types.NamespacedName{Name: ts.Spec.VaultConfig, Namespace: namespace}, config)
		if err != nil {
			return "", "", fmt.Errorf("failed to get TimSecretConfig %s/%s: %w", namespace, ts.Spec.VaultConfig, err)
		}

		return config.Spec.VaultURL, config.Spec.VaultToken, nil
	}

	return "", "", fmt.Errorf("either vaultConfig or both vaultURL and vaultToken must be specified")
}

// updateCondition updates a single condition in the TimSecret status
func (r *TimSecretReconciler) updateCondition(ctx context.Context, ts *secretsv1alpha1.TimSecret, condition metav1.Condition) {
	ts.Status.Conditions = []metav1.Condition{condition}
	_ = r.Status().Update(ctx, ts)
}

// calculateHash calculates SHA256 hash of secret data with deterministic ordering
func calculateHash(data map[string]string) string {
	// Sort keys to ensure deterministic hash
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Calculate hash with sorted keys
	h := sha256.New()
	for _, k := range keys {
		h.Write([]byte(k))
		h.Write([]byte(data[k]))
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// parseSyncInterval parses the sync interval string, defaults to 5 minutes
func (r *TimSecretReconciler) parseSyncInterval(interval string) time.Duration {
	if interval == "" {
		return 5 * time.Minute // Default
	}

	duration, err := time.ParseDuration(interval)
	if err != nil {
		// Invalid format, use default
		return 5 * time.Minute
	}

	// Minimum 30 seconds to avoid excessive load
	if duration < 30*time.Second {
		return 30 * time.Second
	}

	// Maximum 1 hour to ensure timely updates
	if duration > 1*time.Hour {
		return 1 * time.Hour
	}

	return duration
}

// handleError handles errors with automatic retry and exponential backoff
func (r *TimSecretReconciler) handleError(ctx context.Context, ts *secretsv1alpha1.TimSecret, syncInterval time.Duration, err error, reason string) (ctrl.Result, error) {
	// Increment retry count
	ts.Status.RetryCount++
	ts.Status.LastError = err.Error()

	// Cap retry count to prevent overflow (max 20 retries before resetting)
	const maxRetries = 20
	if ts.Status.RetryCount > maxRetries {
		ts.Status.RetryCount = maxRetries
	}

	// Calculate backoff with exponential strategy
	// Base: 10s, Max: syncInterval or 5 minutes
	// Formula: 10s * 2^(min(retryCount-1, 8))
	// Capped at 8 to prevent overflow: 10s * 256 = 2560s = ~42 minutes
	var backoff time.Duration
	exponent := ts.Status.RetryCount - 1
	if exponent > 8 {
		exponent = 8 // Cap exponent to prevent overflow
	}

	if exponent <= 0 {
		backoff = 10 * time.Second
	} else {
		backoff = time.Duration(10*int64(time.Second)) * (1 << uint(exponent))
	}

	// Ensure minimum backoff of 10 seconds
	if backoff < 10*time.Second {
		backoff = 10 * time.Second
	}

	// Cap at sync interval
	if backoff > syncInterval {
		backoff = syncInterval
	}

	// Max 5 minutes for any retry
	if backoff > 5*time.Minute {
		backoff = 5 * time.Minute
	}

	// Update status with failure condition
	ts.Status.Conditions = []metav1.Condition{
		{
			Type:               "Ready",
			Status:             metav1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             reason,
			Message:            fmt.Sprintf("Retry %d (max %d): %v", ts.Status.RetryCount, maxRetries, err),
		},
	}

	if updateErr := r.Status().Update(ctx, ts); updateErr != nil {
		// If we can't update status, return both errors
		return ctrl.Result{RequeueAfter: backoff}, fmt.Errorf("failed to update status: %w (original error: %v)", updateErr, err)
	}

	// Log retry information
	log.FromContext(ctx).Info("Retrying after error",
		"retryCount", ts.Status.RetryCount,
		"backoff", backoff,
		"error", err.Error())

	return ctrl.Result{RequeueAfter: backoff}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TimSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&secretsv1alpha1.TimSecret{}).
		Owns(&corev1.Secret{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 10, // Process 10 TimSecrets in parallel
		}).
		Complete(r)
}
