# Redpanda Datadog Integration Development Kit

This development kit wraps the Redpanda Datadog Integration. The purposes are as follows:

* Simplify development and testing of the integration
* Building custom container images for beta-testing in external deployments
* Deployment of test artifacts in local Kubernetes

> **Note:**
> 
> This project is NOT the Datadog Integration.
> That integration is hosted at https://github.com/DataDog/integrations-extras

### See Also:

- For more details on how metrics are named throughout the integration, see [METRICS.md](./METRICS.md).
- For more details on the function of the integration, see [OVERVIEW.md](./OVERVIEW.md).

# Prerequisites

## Docker

Docker is used for multiple purposes:

- To create multi-platform images of the integration for beta testing
- For integration testing within the Datadog tooling (which spins up a local, one-node Redpanda cluster)

## ddev

Datadog publish a command line tool `ddev`, that performs the necessary compilation and testing. Pre-requisites for that
tool include:

- Python 3.13
- ddev command line tool

> **Important:**
> 
> Be sure to use the latest version of ddev by checking https://github.com/DataDog/integrations-core/releases


The installation instructions can be found at https://docs.datadoghq.com/developers/integrations/python/.

For reference, the installation process I used (on Mac) was as follows:

```shell
brew update
brew upgrade
brew install python@3.13 pipx

wget https://github.com/DataDog/integrations-core/releases/download/ddev-v13.0.0/ddev-13.0.0.pkg
sudo installer -pkg ./ddev-13.0.0.pkg -target /
```

For more detailed instructions, see the Datadog docs ([here](https://datadoghq.dev/integrations-core/setup/) and
[here](https://docs.datadoghq.com/developers/integrations/python/?tab=macos#install-from-the-command-line)).

# Building and Testing

### Where is the integration?

The Redpanda integration can be found in the [`integrations-extras/redpanda`](integrations-extras/redpanda) folder,
which is a git submodule. Once the submodule is available, a symlink to the Redpanda integration can also
be found at [`PROJECT_ROOT/redpanda`](./redpanda).

### Makefile-based Testing (Recommended)

The project provides a [`Makefile`](Makefile), in order to simplify development. The following targets are specified:

- **clean**: removes the build folder that contains the generated yaml, also cleans the `ddev` project
- **build**: compiles the integration and builds a Python wheel
- **test**: runs the unit and integration tests for the Redpanda integration

### Manual Testing (Optional)

It is also possible to test by hand rather than using the [`Makefile`](Makefile), using the following:

```shell
# Configure ddev to point at the code

ddev config set repos.extras integrations-extras
ddev config set repo extras

# Test the current integration

ddev test redpanda
```

### Local Integration Testing

##### Prerequisite: Install Redpanda

In order to test the integration against a local cluster, you can deploy Redpanda as follows:

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

##### Configure Datadog API Key

The [`conf`](conf) folder includes the following configuration file, which must be customised if the metrics are to be sent
to Datadog Cloud:

- [`.env`](conf/.env): this is used to hold the Datadog API key

##### Configure Local Agent (Optional)

While these are unlikely to need to be changed, For reference, the [`conf`](conf) folder includes the following configuration files:

- [`dd-values.yaml`](conf/dd-values.yaml): this is the `values.yaml` file used to install the Datadog agent in our local K8s test cluster
- [`redpanda.yaml`](conf/redpanda.yaml): this is the Datadog agent configuration file
- [`kustomization.yaml`](conf/kustomization.yaml): this defines how to produce the finalised deployment yaml
- [`patch.yaml`](conf/patch.yaml): this defines how we need to patch the Datadog daemonset to include our development artifact

##### Deploy Local Agent

The [`Makefile`](Makefile) provides the following targets for local agent deployment:

- **yaml**: generates a single yaml output file (`target/deployment.yaml`) using a combination of `kubectl`, `helm` and `kustomize` commands
- **deploy**: installs the Datadog agent by applying the generated yaml to a K8s cluster via `kubectl`
- **undeploy**: uninstalls the agent via `kubectl`

### External Integration Testing

##### Build Docker Image

While not required for testing locally on Kubernetes, the project includes a simple Dockerfile for building a Datadog
agent image that includes the Redpanda integration for testing in external environments.

The [`Makefile`](Makefile) provides the following targets for building a custom image:

- **docker**: performs a multi-platform build, tagging the result with the details (image name and version) found in the [`conf/.env`](conf/.env) file
- **push**: pushes the resulting container image to a container registry of your choosing.