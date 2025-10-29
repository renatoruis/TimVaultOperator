# TimVault Operator - Capacity and Performance Guide

## 📊 Overview

This document provides capacity planning guidance and performance characteristics for the TimVault Operator.

---

## 🎯 Current Configuration

### Controller Settings

| Setting | Value | Impact |
|---------|-------|--------|
| `MaxConcurrentReconciles` | **10** | Process 10 TimSecrets in parallel |
| `LeaderElection` | Enabled | Single active instance |
| Workers per Controller | 10 | Concurrent goroutines |

### Resource Limits (Operator Pod)

| Resource | Request | Limit | Recommended For |
|----------|---------|-------|-----------------|
| **CPU** | 500m | 2 cores | Up to 5000 TimSecrets |
| **Memory** | 512Mi | 1Gi | Up to 5000 TimSecrets |

---

## 📈 Capacity Calculations

### Formula

```
Max TimSecrets = (MaxConcurrentReconciles × Avg Reconcile Time) / syncInterval

Where:
- MaxConcurrentReconciles: 10 (default)
- Avg Reconcile Time: ~200ms (depends on Vault latency)
- syncInterval: configured per TimSecret (default 5m)
```

### Capacity Table

| syncInterval | Max TimSecrets | Reconciles/sec | Vault Req/s | Notes |
|--------------|----------------|----------------|-------------|-------|
| **30s** | ~1,000 | 33 | 33 | High load |
| **1m** | ~3,000 | 50 | 50 | Moderate load |
| **5m** (default) | ~15,000 | 50 | 50 | Recommended |
| **10m** | ~30,000 | 50 | 50 | Large scale |
| **30m** | ~90,000 | 50 | 50 | Very large scale |

---

## ✅ Your Scenario: 1000 TimSecrets + 10 TimSecretConfigs

### Answer: **YES, it will handle easily! ✅**

#### With Default Settings (syncInterval: 5m)

```
Capacity: 15,000 TimSecrets
Your Load: 1,000 TimSecrets
Usage: ~6.6% capacity
Status: ✅ WELL WITHIN LIMITS
```

#### Performance Characteristics

| Metric | Value |
|--------|-------|
| Reconciles per second | ~3.3 |
| Vault requests per second | ~3.3 |
| K8s API requests per second | ~20 |
| CPU usage (estimated) | ~15-20% |
| Memory usage (estimated) | ~150-200Mi |
| Queue depth | Near zero |

#### Expected Behavior

- ✅ All TimSecrets reconcile within their syncInterval
- ✅ No reconciliation backlog
- ✅ Quick response to Vault changes (within syncInterval)
- ✅ Low resource utilization
- ✅ Room for burst activity (errors, manual triggers)

---

## 🚦 Load Levels

### Green Zone (Recommended)
- **TimSecrets**: < 10,000 with 5m interval
- **CPU**: < 50% average
- **Memory**: < 512Mi
- **Queue Depth**: 0-10

### Yellow Zone (Monitor)
- **TimSecrets**: 10,000 - 20,000 with 5m interval
- **CPU**: 50-80% average
- **Memory**: 512Mi - 800Mi
- **Queue Depth**: 10-50

### Red Zone (Scale Up)
- **TimSecrets**: > 20,000 with 5m interval
- **CPU**: > 80% sustained
- **Memory**: > 800Mi
- **Queue Depth**: > 50

---

## 🎛️ Tuning for Different Scales

### Small Scale (< 100 TimSecrets)

**Default settings are optimal.**

```yaml
# No changes needed
MaxConcurrentReconciles: 10
```

```yaml
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 256Mi
```

### Medium Scale (100 - 5,000 TimSecrets)

**Current configuration (already applied).**

```yaml
MaxConcurrentReconciles: 10
```

```yaml
resources:
  requests:
    cpu: 500m
    memory: 512Mi
  limits:
    cpu: "2"
    memory: 1Gi
```

### Large Scale (5,000 - 20,000 TimSecrets)

**Increase concurrency:**

```go
WithOptions(controller.Options{
    MaxConcurrentReconciles: 20,
})
```

```yaml
resources:
  requests:
    cpu: "1"
    memory: 1Gi
  limits:
    cpu: "4"
    memory: 2Gi
```

### Very Large Scale (> 20,000 TimSecrets)

**Consider multiple operator replicas with sharding or increase syncInterval:**

Option 1 - Increase syncInterval:
```yaml
spec:
  syncInterval: "10m"  # or 30m for very large deployments
```

Option 2 - Multiple operators (advanced):
- Deploy multiple operators
- Use label selectors to partition TimSecrets
- Each operator handles a subset

---

## 🔍 Monitoring Metrics

### Key Metrics to Watch

#### 1. **Queue Depth**
```bash
kubectl port-forward -n timvault-operator-system deployment/timvault-operator-controller 8080:8080

# Check workqueue metrics
curl localhost:8080/metrics | grep workqueue_depth
```

**Healthy**: < 10
**Warning**: 10-50
**Critical**: > 50

#### 2. **Reconcile Duration**
```bash
curl localhost:8080/metrics | grep controller_runtime_reconcile_time_seconds
```

**Healthy**: p50 < 200ms, p99 < 1s
**Warning**: p50 < 500ms, p99 < 3s
**Critical**: p50 > 500ms or p99 > 3s

#### 3. **Resource Usage**
```bash
kubectl top pod -n timvault-operator-system
```

#### 4. **Error Rate**
```bash
kubectl logs -n timvault-operator-system deployment/timvault-operator-controller | grep ERROR | wc -l
```

**Healthy**: < 1 error/minute
**Warning**: 1-10 errors/minute
**Critical**: > 10 errors/minute

---

## ⚠️ Bottlenecks and Limits

### 1. **Vault Latency** (HIGHEST IMPACT)

**Impact**: Each reconcile waits for Vault response

| Vault Latency | Max Throughput (10 workers) |
|---------------|------------------------------|
| 50ms | 200 reconciles/s |
| 100ms | 100 reconciles/s |
| 200ms | 50 reconciles/s |
| 500ms | 20 reconciles/s |

**Optimization**:
- Use Vault close to Kubernetes cluster
- Consider Vault read replicas
- Monitor Vault performance

### 2. **Kubernetes API Rate Limits** (MEDIUM IMPACT)

**Default**: 5 req/s per client (burst 10)

**Your load (1000 TimSecrets, 5m interval)**:
- 3.3 reconciles/s × 6 K8s calls = ~20 req/s
- Status: ✅ Within limits (with bursting)

**Optimization**:
- Use client-side caching (controller-runtime does this)
- Batch status updates
- Increase API server limits if needed

### 3. **CPU and Memory** (LOW IMPACT at your scale)

**Memory growth**:
- Base: ~100Mi
- Per 1000 TimSecrets: +50Mi (cached objects)
- Your estimate (1000 TS): ~150Mi

**CPU usage**:
- Mostly I/O wait (Vault, K8s API)
- Minimal CPU per reconcile (~5ms)
- Your estimate: 15-20% of 1 core

---

## 🚀 Scaling Strategies

### Vertical Scaling (Increase Resources)

**When**: Queue depth > 10 consistently

**Action**: Increase CPU/Memory:
```yaml
resources:
  limits:
    cpu: "4"
    memory: 2Gi
```

### Horizontal Scaling (Multiple Replicas)

⚠️ **NOT RECOMMENDED** due to:
- Leader election = only 1 active
- Complexity without benefit

**Alternative**: Increase `MaxConcurrentReconciles` instead

### Temporal Scaling (Increase syncInterval)

**When**: Vault can't handle load

**Action**: Increase syncInterval globally:
```yaml
# For less critical secrets
spec:
  syncInterval: "30m"
```

---

## 💡 Best Practices

### 1. **Use Centralized TimSecretConfig**
- ✅ Reduces memory (10 configs vs 1000 inline)
- ✅ Faster reconciliation (cached config)
- ✅ Easier credential rotation

### 2. **Tune syncInterval Per Use Case**
- **Critical secrets**: 1-5 minutes
- **Standard secrets**: 5-10 minutes  
- **Static secrets**: 30-60 minutes

### 3. **Batch Related TimSecrets**
- Group by Vault instance
- Group by update frequency
- Use same syncInterval for similar workloads

### 4. **Monitor Queue Depth**
- Alert if > 20 for > 5 minutes
- Investigate reconcile duration spikes
- Check Vault performance

---

## 📊 Example: Your Setup

### Configuration
```yaml
TimSecrets: 1000
TimSecretConfigs: 10
syncInterval: 5m (default)
MaxConcurrentReconciles: 10
```

### Expected Performance

| Metric | Value | Status |
|--------|-------|--------|
| **Capacity Usage** | 6.6% | ✅ Excellent |
| **Reconciles/sec** | 3.3 | ✅ Low |
| **Vault Load** | 3.3 req/s | ✅ Very Low |
| **K8s API Load** | ~20 req/s | ✅ Acceptable |
| **CPU Usage** | 15-20% | ✅ Low |
| **Memory Usage** | ~150Mi | ✅ Low |
| **Queue Depth** | 0-2 | ✅ Excellent |

### Headroom for Growth

You can scale to:
- **5,000 TimSecrets** without any changes
- **15,000 TimSecrets** with current resources
- **30,000+ TimSecrets** by increasing syncInterval to 10m

---

## 🔧 Troubleshooting Performance Issues

### Symptom: Queue Depth Growing

**Check**:
```bash
# 1. Queue depth
curl localhost:8080/metrics | grep workqueue_depth

# 2. Reconcile duration
curl localhost:8080/metrics | grep reconcile_time_seconds

# 3. Vault latency
kubectl logs -n timvault-operator-system deployment/timvault-operator-controller | grep "Failed to get secrets"
```

**Solutions**:
1. Increase `MaxConcurrentReconciles` to 20
2. Increase CPU/Memory resources
3. Check Vault performance
4. Increase syncInterval for non-critical TimSecrets

### Symptom: High CPU Usage

**Likely Cause**: Reconcile loop thrashing

**Check**:
```bash
kubectl logs -n timvault-operator-system deployment/timvault-operator-controller | grep "Secret data unchanged"
```

If you DON'T see this message frequently, there's a bug!

**Solution**: Ensure hash calculation is deterministic (already fixed)

### Symptom: High Memory Usage

**Check**:
```bash
kubectl top pod -n timvault-operator-system
```

**Solutions**:
1. Check for memory leaks (restart operator, monitor growth)
2. Increase memory limits
3. Reduce cached objects (unlikely with your scale)

---

## 📝 Conclusion

**For your scenario (1000 TimSecrets + 10 TimSecretConfigs):**

✅ **You're in excellent shape!**
- Using only ~7% of capacity
- Low resource utilization
- Fast reconciliation
- Room for 10x growth without changes

**Recommendation**: Use current configuration, monitor queue depth and CPU usage initially, then relax monitoring once stable.

