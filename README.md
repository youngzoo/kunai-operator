# kunai-operator


## Overview 
This is the very early stage operator release designed to deploy Kunai as a Kubernetes DaemonSet. At present, modifications to the Kunai configuration ConfigMap result in a basic service restart. Key features include:

- Kunai Docker Image Packaging and Docker Compose: We provide a comprehensive Docker image packaging solution for Kunai, along with a [docker-compose.yaml](deploy/docker/docker-compose.yaml) for streamlined local development and testing.

- Kunai Operator Binary (Go) and Image Packaging: The Kunai Operator, developed in Go, is packaged as a Docker image, enabling efficient Kubernetes deployment and management of Kunai DaemonSets.

- Helm Chart for Kunai Operator: A dedicated Helm chart, found in [deploy/helm](deploy/helm), facilitates easy installation and configuration of the Kunai Operator within your Kubernetes cluster.

## Development

Please refer to the [Justfile](Justfile) for development instructions.