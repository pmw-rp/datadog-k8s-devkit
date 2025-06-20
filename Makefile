EXECUTABLES := ddev kubectl helm kustomize
K := $(foreach exec,$(EXECUTABLES),\
	$(if $(shell which $(exec)),some string,$(error "Error: command $(exec) not found in PATH - this is required for makefile to proceed")))

ROOT := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))

.PHONY: config clean build test yaml deploy undeploy

config:
	ddev config set repos.extras ${ROOT}/integrations-extras
	ddev config set repo extras

clean: config
	rm -rf target || true
	ddev clean

build: config
	ddev release build redpanda

test: config
	ddev test redpanda

yaml: build
	mkdir ${ROOT}/target || true
	kubectl create secret generic datadog-secret --from-literal api-key=$(API_KEY) --dry-run=client -o yaml > ${ROOT}/target/dd-secret.yaml
	kubectl create configmap redpanda-dd-config --from-file ${ROOT}/conf/redpanda.yaml --dry-run=client -o yaml > ${ROOT}/target/redpanda-datadog-config-configmap.yaml
	helm template datadog-agent datadog/datadog -f ${ROOT}/conf/dd-values.yaml > ${ROOT}/target/pre-deployment.yaml
	kubectl create configmap redpanda-dd-wheel \
		--from-file=${ROOT}/integrations-extras/redpanda/dist/datadog_redpanda-2.1.0-py2.py3-none-any.whl \
		--dry-run=client -o yaml > ${ROOT}/target/redpanda-datadog-wheel-configmap.yaml
	cp ${ROOT}/conf/patch.yaml ${ROOT}/target
	cp ${ROOT}/conf/kustomization.yaml ${ROOT}/target
	kustomize build target > target/deployment.yaml

deploy: yaml
	kubectl apply -f ${ROOT}/target/deployment.yaml

undeploy: yaml
	kubectl delete -f ${ROOT}/target/deployment.yaml