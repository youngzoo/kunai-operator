setup-kind:
	kind create cluster

build-images:
	docker build -t yeongjukang/kunai:latest -f dockerfiles/Dockerfile.daemonset .
	docker build -t yeongjukang/kunai-operator:latest -f dockerfiles/Dockerfile.operator .

push-images: build-images
	docker push yeongjukang/kunai:latest
	docker push yeongjukang/kunai-operator:latest

kind-load-images: build-images
	kind load docker-image yeongjukang/kunai:latest
	kind load docker-image yeongjukang/kunai-operator:latest

docker-compose-up:
	docker compose -f deploy/docker/docker-compose.yaml up -d

test-helm:
	helm template kunai -nkunai deploy/helm > test-helm.yaml
	cat test-helm.yaml
	helm lint deploy/helm

deploy-on-kind:
	helm upgrade --install kunai -nkunai deploy/helm