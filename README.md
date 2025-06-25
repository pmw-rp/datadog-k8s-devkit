# Redpanda Datadog Integration Development Kit



> Note: This project is a development kit for the Redpanda Datadog integration, NOT the integration itself.
> 
> This is useful for making changes to the integration, deploying for testing in local Kubernetes, and
> building container images for beta testing in external deployments.
> 
> NOT FOR CUSTOMER OR PRODUCTION USE!

# Prerequisites

## Docker

Docker is used for multiple purposes:
- To create multi-platform images of the integration for beta testing
- For integration testing within the Datadog tooling (which spins up a local, one-node Redpanda cluster)

In my testing, I use Orbstack.

## ddev

Datadog publish a command line tool `ddev`, that performs the necessary compilation and testing. Pre-requisites for that
tool include:

- Python 3.12
- pipx
- ddev command line tool

The process I used on Mac was as follows:

```shell
brew update
brew upgrade
brew install python@3.12

curl -L -o ddev-11.4.0.pkg https://github.com/DataDog/integrations-core/releases/download/ddev-v11.4.0/ddev-11.4.0.pkg
sudo installer -pkg ./ddev-11.4.0.pkg -target /

# now restart your shell to take path change into effect

ddev --version
```

For more detailed instructions, see the Datadog docs ([here](https://datadoghq.dev/integrations-core/setup/) and
[here](https://docs.datadoghq.com/developers/integrations/python/?tab=macos#install-from-the-command-line)).

## Redpanda Integration

The Redpanda integration can be found in the [`integrations-extras/redpanda`](integrations-extras/redpanda) folder.

With `integrations-extras` available, we first need to configure `ddev` and test the current integration:

```shell
# Configure ddev to point at the code

ddev config set repos.extras integrations-extras
ddev config set repo extras

# Test the current integration

ddev test redpanda
```

(Configuring the repo location configuration and testing can also be performed using the [`Makefile`](Makefile)).

## Redpanda Test Cluster

The last pre-requisite is a Redpanda cluster to test with. I built my test cluster as follows, running via Orbstack, but any Redpanda cluster deployed on K8s will do.

```shell
# Install Cert Manager

helm repo add jetstack https://charts.jetstack.io
helm repo update
helm install cert-manager jetstack/cert-manager --set crds.enabled=true --namespace cert-manager --create-namespace

# Install Redpanda via Helm

cat << EOF | helm install redpanda redpanda/redpanda \
  --version 5.9.24 \
  --namespace redpanda \
  --create-namespace \
  -f -
image:
  tag: v25.1.5
external:
  service:
    enabled: false
statefulset:
  replicas: 1
config:
  cluster:
    default_topic_replications: 1
tls:
  enabled: false
EOF
```

(Keep an eye on the Helm chart and Redpanda versions, depending on what you need.)

# Configuration

The [`conf`](conf) folder includes the following configuration files:

- [`.env`](conf/.env): this is used to hold the Datadog API key
- [`dd-values.yaml`](conf/dd-values.yaml): this is the `values.yaml` file used to install the Datadog agent in our local K8s test cluster
- [`redpanda.yaml`](conf/redpanda.yaml): this is the Datadog configuration file
- [`kustomization.yaml`](conf/kustomization.yaml): this defines how to produce the finalised deployment yaml
- [`patch.yaml`](conf/patch.yaml): this defines how we need to patch the Datadog daemonset to include our development artifact

# Testing on Local Kubernetes

The project allows for testing your integration changes by adding the generated `.whl` Python wheel to the standard Datadog image - no custom image is required for testing.

## API Key

Firstly, if you will be deploying the integration, it is necessary to obtain a [Datadog API key](https://app.datadoghq.com/organization-settings/api-keys) by creating it in the UI. Store the resulting key in the [`.env`](conf/.env) file.

Once the [`.env`](conf/.env) file has been updated, run the following to make the key available as a local environment variable:

```shell
source conf/.env
```

## Make

The project provides a simple [`Makefile`](Makefile), in order to simplify development. The following targets are specified:

- **clean**: removes the build folder that contains the generated yaml, also cleans the `ddev` project
- **build**: compiles the integration and builds a Python wheel
- **test**: runs the unit and integration tests for the Redpanda integration
- **yaml**: generates a single yaml output file (`target/deployment.yaml`) using a combination of `kubectl`, `helm` and `kustomize` commands
- **deploy**: installs the Datadog agent by applying the generated yaml to a K8s cluster via `kubectl`
- **undeploy**: uninstalls the agent via `kubectl`

# Building (and pushing) the Container

While not required for testing locally on Kubernetes, the project includes a simple Dockerfile for building a Datadog agent image that includes the Redpanda integration for testing in external environments.

## Make

The project provides the following additional [`Makefile`](Makefile) targets for building and pushing the container:

- **docker**: performs a multi-platform build, tagging the result with the details (image name and version) found in the [`conf/.env`](conf/.env) file
- **push**: pushes the resulting container image to a container registry of your choosing

# Overview of the Integration

The agent performs two main functions:

1. It renames metrics into a simpler, friendlier naming structure
2. It filters metrics scraped from Redpanda

Both of these functions are handled by [`metrics.py`](integrations-extras/redpanda/datadog_checks/redpanda/metrics.py).

## Renaming

Metrics are renamed in the Python dicts as follows (original metric name -> new metric name):

```Python
REDPANDA_APPLICATION = {
    'redpanda_application_uptime_seconds_total': 'application.uptime',
    'redpanda_application_build': 'application.build',
}
```

The renamed metrics (`application.uptime`, etc.) are all put under the `redpanda` namespace, therefore becoming `redpanda.application.uptime` in the Datadog UI. (The namespace is defined in [`redpanda.py`](integrations-extras/redpanda/datadog_checks/redpanda/redpanda.py), `__NAMESPACE__ = 'redpanda'`).

Putting this together, we see that `redpanda_application_uptime_seconds_total` is renamed to `redpanda.application.uptime`.

## Filtering

Metrics that are known by the integration (see [`metrics.py`](integrations-extras/redpanda/datadog_checks/redpanda/metrics.py) and [`metadata.csv`](integrations-extras/redpanda/metadata.csv)) are transformed and sent to Datadog. Any metrics scraped that are unknown are ignored by the Datadog Agent and not forwarded.

## Metrics Groups

Datadog groups metrics into logical groups, that typically relate to an area of functionality - examples of groups include:

- `REDPANDA_APPLICATION` (as seen above)
- `REDPANDA_CLUSTER`
- ...
- `REDPANDA_STORAGE`

### Default vs Optional Metrics

One of the uses of metrics groups is to help define them as either default or optional:

```python
INSTANCE_DEFAULT_METRICS = [
    REDPANDA_APPLICATION,
    REDPANDA_CLUSTER,
    ...
    REDPANDA_STORAGE,
]

ADDITIONAL_METRICS_MAP = {
    'redpanda.cloud': REDPANDA_CLOUD,
    'redpanda.controller': REDPANDA_CONTROLLER,
    ...
    'redpanda.schemaregistry': REDPANDA_SCHEMA_REGISTRY,
}
```

As the name suggests, default metrics are always published, whereas optional (additional) metrics are only published by
the agent if the user-supplied configuration includes the key (e.g. `redpanda.controller`) in the list of metrics groups:

```yaml
instances:
- openmetrics_endpoint: http://redpanda.redpanda.svc.cluster.local:9644/public_metrics
  metric_groups:
  - redpanda.controller # <---- additional metrics group requested here
logs:
- type: journald
  source: redpanda
```