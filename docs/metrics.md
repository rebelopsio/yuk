# Metrics Reference

Yuk exposes comprehensive Prometheus metrics to provide observability into the controller's operation. These metrics can be used for monitoring, alerting, and debugging.

## Endpoint

Metrics are exposed on the `/metrics` endpoint of the controller's metrics port (default: 8080).

## Available Metrics

### Controller Metrics

#### `yuk_controller_reconciliation_duration_seconds`
**Type:** Histogram  
**Description:** Time taken for YukConfig reconciliation  
**Labels:**
- `namespace` - Namespace of the YukConfig resource
- `name` - Name of the YukConfig resource  
- `result` - Result of reconciliation (`success`, `error`, `skipped`)

#### `yuk_controller_reconciliation_total`
**Type:** Counter  
**Description:** Total number of reconciliations performed  
**Labels:**
- `namespace` - Namespace of the YukConfig resource
- `name` - Name of the YukConfig resource
- `result` - Result of reconciliation (`success`, `error`, `skipped`)

#### `yuk_controller_queue_depth`
**Type:** Gauge  
**Description:** Current depth of the controller work queue  
**Labels:**
- `controller` - Controller name

### Repository Metrics

#### `yuk_repository_checks_total`
**Type:** Counter  
**Description:** Total number of repository checks performed  
**Labels:**
- `repository_type` - Type of repository (`ecr`)
- `repository_name` - Name of the repository
- `result` - Result of the check (`success`, `error`)

#### `yuk_repository_check_duration_seconds`
**Type:** Histogram  
**Description:** Time taken for repository checks  
**Labels:**
- `repository_type` - Type of repository (`ecr`)
- `repository_name` - Name of the repository

### Git Operation Metrics

#### `yuk_git_operations_total`
**Type:** Counter  
**Description:** Total number of Git operations performed  
**Labels:**
- `operation` - Type of operation (`clone`, `commit`, `push`)
- `repository` - Git repository URL
- `result` - Result of the operation (`success`, `error`)

#### `yuk_git_operation_duration_seconds`
**Type:** Histogram  
**Description:** Time taken for Git operations  
**Labels:**
- `operation` - Type of operation (`clone`, `commit`, `push`)
- `repository` - Git repository URL

### Update Metrics

#### `yuk_updates_performed_total`
**Type:** Counter  
**Description:** Total number of successful updates performed  
**Labels:**
- `namespace` - Namespace of the YukConfig resource
- `name` - Name of the YukConfig resource
- `repository_type` - Type of repository (`ecr`)
- `repository_name` - Name of the repository

#### `yuk_files_updated_total`
**Type:** Counter  
**Description:** Total number of files updated  
**Labels:**
- `namespace` - Namespace of the YukConfig resource
- `name` - Name of the YukConfig resource
- `file_path` - Path to the updated file

### Status Metrics

#### `yuk_current_version_info`
**Type:** Gauge  
**Description:** Information about the current version being monitored (value is always 1)  
**Labels:**
- `namespace` - Namespace of the YukConfig resource
- `name` - Name of the YukConfig resource
- `repository_name` - Name of the repository
- `current_tag` - Current tag being used
- `latest_tag` - Latest tag available

#### `yuk_config_status`
**Type:** Gauge  
**Description:** Status of YukConfig resources (1=ready, 0=not ready)  
**Labels:**
- `namespace` - Namespace of the YukConfig resource
- `name` - Name of the YukConfig resource
- `condition_type` - Type of condition (`Ready`, etc.)

### Timestamp Metrics

#### `yuk_last_check_timestamp_seconds`
**Type:** Gauge  
**Description:** Timestamp of the last repository check  
**Labels:**
- `namespace` - Namespace of the YukConfig resource
- `name` - Name of the YukConfig resource
- `repository_name` - Name of the repository

#### `yuk_last_update_timestamp_seconds`
**Type:** Gauge  
**Description:** Timestamp of the last successful update  
**Labels:**
- `namespace` - Namespace of the YukConfig resource
- `name` - Name of the YukConfig resource
- `repository_name` - Name of the repository

### Error Metrics

#### `yuk_errors_total`
**Type:** Counter  
**Description:** Total number of errors encountered  
**Labels:**
- `error_type` - Type of error (`repository`, `git`, `yaml`, `auth`, `validation`, `network`)
- `namespace` - Namespace of the YukConfig resource
- `name` - Name of the YukConfig resource

## Monitoring Setup

### Prometheus Configuration

Add the following to your Prometheus configuration to scrape Yuk metrics:

```yaml
scrape_configs:
- job_name: 'yuk-controller'
  kubernetes_sd_configs:
  - role: endpoints
    namespaces:
      names:
      - yuk-system
  relabel_configs:
  - source_labels: [__meta_kubernetes_service_name]
    action: keep
    regex: yuk-metrics
  - source_labels: [__meta_kubernetes_endpoint_port_name]
    action: keep
    regex: metrics
```

### ServiceMonitor

If using the Prometheus Operator, enable the ServiceMonitor in your Helm values:

```yaml
monitoring:
  enabled: true
  serviceMonitor:
    enabled: true
    namespace: monitoring  # Optional: specify monitoring namespace
    labels:
      app: prometheus
    annotations:
      prometheus.io/scrape: "true"
```

## Example Queries

### Reconciliation Rate
```promql
rate(yuk_controller_reconciliation_total[5m])
```

### Error Rate
```promql
rate(yuk_errors_total[5m])
```

### Repository Check Success Rate
```promql
rate(yuk_repository_checks_total{result="success"}[5m]) / 
rate(yuk_repository_checks_total[5m])
```

### Git Operation Duration (95th percentile)
```promql
histogram_quantile(0.95, rate(yuk_git_operation_duration_seconds_bucket[5m]))
```

### Number of Configs Out of Sync
```promql
count(yuk_current_version_info{current_tag!="",latest_tag!="",current_tag!=latest_tag})
```

### Time Since Last Update
```promql
time() - yuk_last_update_timestamp_seconds
```

## Alerting Rules

### High Error Rate
```yaml
- alert: YukHighErrorRate
  expr: rate(yuk_errors_total[5m]) > 0.1
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Yuk controller has high error rate"
    description: "Yuk controller {{ $labels.namespace }}/{{ $labels.name }} has error rate of {{ $value }} errors/second"
```

### Repository Check Failures
```yaml
- alert: YukRepositoryCheckFailure
  expr: rate(yuk_repository_checks_total{result="error"}[5m]) > 0
  for: 2m
  labels:
    severity: warning
  annotations:
    summary: "Yuk repository checks failing"
    description: "Repository {{ $labels.repository_name }} checks are failing"
```

### Config Not Updated
```yaml
- alert: YukConfigNotUpdated
  expr: (time() - yuk_last_update_timestamp_seconds) > 86400  # 24 hours
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Yuk config not updated in 24 hours"
    description: "YukConfig {{ $labels.namespace }}/{{ $labels.name }} has not been updated in over 24 hours"
```

## Dashboard

For a comprehensive dashboard, consider creating Grafana panels for:

1. **Overview Panel**: Total configs, success rate, error rate
2. **Reconciliation Metrics**: Duration, frequency, success/failure rates
3. **Repository Metrics**: Check frequency, success rates, latest versions
4. **Git Operations**: Clone/push durations, success rates
5. **Error Analysis**: Error breakdown by type and resource
6. **Version Tracking**: Current vs latest versions across all configs