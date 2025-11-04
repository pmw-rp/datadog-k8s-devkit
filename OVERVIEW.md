# Overview of the Integration

The agent integration performs three main functions:

1. It groups metrics into functional areas
2. It renames metrics into a simpler, friendlier naming structure
3. It filters out metrics

## 1. Metrics Groups

Redpanda's Datadog integration groups metrics into functional groups. For example:

- `REDPANDA_APPLICATION`
- `REDPANDA_CLUSTER`
- ...
- `REDPANDA_STORAGE`

Not all the metrics groups are enabled by default. The groups listed in `INSTANCE_DEFAULT_METRICS` (in [`metrics.py`](integrations-extras/redpanda/datadog_checks/redpanda/metrics.py))
define what metrics groups are sent by default, using code similar to this:

```python
INSTANCE_DEFAULT_METRICS = [
    REDPANDA_APPLICATION,
    REDPANDA_CLUSTER,
    ...
    REDPANDA_STORAGE,
]
```

### Optional Metrics Groups

As the name suggests, optional (additional) metrics groups are only published by the agent if the agent
configuration lists the group (e.g. `redpanda.controller`) in the agent configuration:

```yaml
# redpanda.yaml
instances:
- openmetrics_endpoint: http://redpanda.redpanda.svc.cluster.local:9644/public_metrics
  metric_groups:
  - redpanda.controller        # <---- additional metrics group
logs:
- type: journald
  source: redpanda
```

The optional groups are defined in `ADDITIONAL_METRICS_MAP` (in [`metrics.py`](integrations-extras/redpanda/datadog_checks/redpanda/metrics.py)),
using code similar to this:

```python
ADDITIONAL_METRICS_MAP = {
    'redpanda.cloud': REDPANDA_CLOUD,
    'redpanda.controller': REDPANDA_CONTROLLER,
    ...
    'redpanda.schemaregistry': REDPANDA_SCHEMA_REGISTRY,
}
```

## 2. Metrics Renaming

Metric renaming is configured on a metric-by-metric basis, with the mapping defined in
[`metrics.py`](integrations-extras/redpanda/datadog_checks/redpanda/metrics.py).

Metrics are renamed within each metrics group as follows:

```Python
REDPANDA_APPLICATION = {
    'redpanda_application_uptime_seconds_total': 'application.uptime',
    'redpanda_application_build': 'application.build',
}
```
In this example, we see:
- There is a single metrics group `REDPANDA_APPLICATION`, which contains two metrics:
  - The first metric `redpanda_application_uptime_seconds_total` is renamed to `redpanda.application.uptime`
  - The second metric `redpanda_application_build` is renamed to `redpanda.application.build`

> Note:
> 
> The renamed metrics (`application.uptime`, etc.) are all put under the `redpanda` namespace, therefore becoming `redpanda.application.uptime` in the Datadog UI. (The namespace is defined in [`redpanda.py`](integrations-extras/redpanda/datadog_checks/redpanda/redpanda.py), `__NAMESPACE__ = 'redpanda'`).

## 3. Metrics Filtering

There are two type of metrics filtering performed by the agent:

1. Any metrics within optional metrics groups that have not been enabled will not be sent to Datadog.
2. Any metrics that are unknown to the integration (see [`metrics.py`](integrations-extras/redpanda/datadog_checks/redpanda/metrics.py)) will be ignored and not sent to Datadog.

# Metrics Naming

For more details on how metrics are named throughout the integration, see [METRICS.md](./METRICS.md).