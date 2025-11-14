# Waverless Database Deployment

This document describes the MySQL database deployment included with Waverless.

## Overview

Waverless includes a MySQL 8.0 deployment for production use, automatically configured and initialized with all required tables.

## Architecture

```
┌─────────────────┐
│  Waverless App  │
└────────┬────────┘
         │
    ┌────┴────┬──────────────┐
    │         │              │
┌───▼────┐ ┌──▼──────┐ ┌────▼────┐
│ Redis  │ │  MySQL  │ │   K8s   │
│ Cache  │ │Database │ │ Cluster │
└────────┘ └─────────┘ └─────────┘
```

## Default Configuration

**Service**: `waverless-mysql-svc` (ClusterIP)
**Port**: 3306
**Database**: `waverless`

**Credentials**:
- Root Password: `waverless_root_pass`
- User: `waverless`
- Password: `waverless_pass`

## Database Schema

The MySQL deployment includes automatic initialization with the following tables:

### Core Tables

1. **tasks** - Task lifecycle and execution tracking
2. **endpoints** - Endpoint configuration and metadata
3. **autoscaler_config** - Autoscaler settings per endpoint

### Event Tracking

4. **task_events** - Detailed task event audit log
5. **scaling_events** - Autoscaling event history

### Statistics & Monitoring

6. **task_statistics** - Pre-aggregated task statistics (global/per-endpoint)
7. **gpu_usage_records** - Task-level GPU usage records
8. **gpu_usage_statistics_minute** - Minute-level GPU aggregation
9. **gpu_usage_statistics_hourly** - Hourly GPU aggregation
10. **gpu_usage_statistics_daily** - Daily GPU aggregation

## Deployment

### Automatic Deployment

The MySQL database is automatically deployed when using the deployment script:

```bash
./deploy.sh install
```

Deployment sequence:
1. Namespace creation
2. RBAC setup
3. ConfigMap creation
4. **Redis deployment** ← Cache layer
5. **MySQL deployment** ← Database layer
6. Waverless API server
7. Web UI

### Manual Deployment

If deploying manually:

```bash
# Apply MySQL deployment
kubectl apply -f k8s/mysql-deployment.yaml

# Wait for MySQL to be ready
kubectl wait --for=condition=ready pod -l app=waverless-mysql -n wavespeed --timeout=180s

# Verify MySQL is running
kubectl get pods -n wavespeed -l app=waverless-mysql
```

## Database Initialization

The database is automatically initialized on first startup via ConfigMap containing `init.sql`:

- Creates all required tables
- Sets up indexes for optimal performance
- Initializes global statistics record

## Resource Configuration

**Requests**:
- Memory: 512Mi
- CPU: 200m

**Limits**:
- Memory: 1Gi
- CPU: 500m

## Storage

**Current**: EmptyDir (ephemeral)
**Production Recommendation**: Use PersistentVolumeClaim

### Upgrading to Persistent Storage

1. Create a PersistentVolumeClaim:

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: mysql-pvc
  namespace: wavespeed
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 20Gi
  storageClassName: your-storage-class
```

2. Update `k8s/mysql-deployment.yaml`:

```yaml
volumes:
- name: mysql-data
  persistentVolumeClaim:
    claimName: mysql-pvc
```

## Health Checks

**Liveness Probe**: `mysqladmin ping` every 10 seconds (starts after 30s)
**Readiness Probe**: `mysqladmin ping` every 5 seconds (starts after 10s)

## Accessing the Database

### From within the cluster

```bash
mysql -h waverless-mysql-svc -u waverless -pwaverless_pass waverless
```

### Using kubectl port-forward

```bash
# Port forward to localhost
kubectl port-forward -n wavespeed svc/waverless-mysql-svc 3306:3306

# Connect from local machine
mysql -h 127.0.0.1 -u waverless -pwaverless_pass waverless
```

### From a debug pod

```bash
kubectl run -it --rm debug --image=mysql:8.0 --restart=Never -n wavespeed -- \
  mysql -h waverless-mysql-svc -u waverless -pwaverless_pass waverless
```

## Backup & Restore

### Backup

```bash
# Export database to file
kubectl exec -n wavespeed deployment/waverless-mysql -- \
  mysqldump -u waverless -pwaverless_pass waverless > backup.sql
```

### Restore

```bash
# Import from backup file
kubectl exec -i -n wavespeed deployment/waverless-mysql -- \
  mysql -u waverless -pwaverless_pass waverless < backup.sql
```

## Monitoring

### Check MySQL status

```bash
# View logs
kubectl logs -n wavespeed -l app=waverless-mysql -f

# Check resource usage
kubectl top pod -n wavespeed -l app=waverless-mysql

# View pod details
kubectl describe pod -n wavespeed -l app=waverless-mysql
```

### Query statistics

```sql
-- Check table sizes
SELECT
    table_name,
    ROUND(((data_length + index_length) / 1024 / 1024), 2) AS size_mb
FROM information_schema.tables
WHERE table_schema = 'waverless'
ORDER BY (data_length + index_length) DESC;

-- View recent tasks
SELECT task_id, endpoint, status, created_at, completed_at
FROM tasks
ORDER BY created_at DESC
LIMIT 10;

-- Check global statistics
SELECT * FROM task_statistics WHERE scope_type = 'global';

-- View GPU usage summary
SELECT
    DATE(completed_at) as date,
    endpoint,
    COUNT(*) as task_count,
    SUM(gpu_hours) as total_gpu_hours
FROM gpu_usage_records
GROUP BY DATE(completed_at), endpoint
ORDER BY date DESC, total_gpu_hours DESC;
```

## Troubleshooting

### MySQL pod not starting

```bash
# Check pod status
kubectl get pods -n wavespeed -l app=waverless-mysql

# View events
kubectl describe pod -n wavespeed -l app=waverless-mysql

# Check logs
kubectl logs -n wavespeed -l app=waverless-mysql
```

### Connection issues

```bash
# Test connectivity from Waverless pod
kubectl exec -it -n wavespeed deployment/waverless -- \
  nc -zv waverless-mysql-svc 3306

# Verify service
kubectl get svc -n wavespeed waverless-mysql-svc
```

### Reset database

```bash
# Delete MySQL pod (will recreate with fresh database)
kubectl delete pod -n wavespeed -l app=waverless-mysql

# Or delete entire deployment
kubectl delete -f k8s/mysql-deployment.yaml
kubectl apply -f k8s/mysql-deployment.yaml
```

## Security Considerations

⚠️ **Important**: The default credentials are for development/testing only!

### For Production:

1. **Use Kubernetes Secrets** for database credentials:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: mysql-secret
  namespace: wavespeed
type: Opaque
stringData:
  root-password: <strong-random-password>
  user: waverless
  password: <strong-random-password>
```

2. **Update deployment** to use secrets:

```yaml
env:
- name: MYSQL_ROOT_PASSWORD
  valueFrom:
    secretKeyRef:
      name: mysql-secret
      key: root-password
- name: MYSQL_PASSWORD
  valueFrom:
    secretKeyRef:
      name: mysql-secret
      key: password
```

3. **Enable TLS/SSL** for encrypted connections
4. **Use network policies** to restrict database access
5. **Enable audit logging** for compliance
6. **Regular backups** to persistent storage

## Migration from External Database

If you have an existing external MySQL database:

1. Export data from old database
2. Update ConfigMap to point to external database
3. Remove or skip MySQL deployment
4. Import data to external database

```yaml
# config.yaml in ConfigMap
mysql:
  host: external-mysql.example.com
  port: 3306
  user: waverless
  password: <external-password>
  database: waverless
```

---

**Last Updated**: 2025-11-14
**Version**: 1.0
