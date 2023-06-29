[![Go Report Card](https://goreportcard.com/badge/github.com/danzim/prometheus-provider)](https://goreportcard.com/report/github.com/danzim/prometheus-provider)

# prometheus-provider
Integrate OPA Gatekeeper's ExternalData feature with OpenShift Prometheus to determine if a deployment meets the ratio between resource requests and actual consumption

> This repo is meant for testing Gatekeeper external data feature. Do not use for production.

## Installation

- Deploy Gatekeeper with external data enabled (`--enable-external-data`)
```sh
helm repo add gatekeeper https://open-policy-agent.github.io/gatekeeper/charts
helm install gatekeeper/gatekeeper  \
    --name-template=gatekeeper \
    --namespace gatekeeper-system --create-namespace \
    --set enableExternalData=true \
    --set controllerManager.dnsPolicy=ClusterFirst,audit.dnsPolicy=ClusterFirst \
    --version 3.12.0
```
_Note: This repository is currently only working with Gatekeeper 3.12 and the `externalData` feature in `beta`.