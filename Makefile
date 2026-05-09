-include .env
export

IMAGE     ?= registry.gitlab.com/liuzuocai/video-saver
TAG       ?= latest
NAMESPACE ?= video-saver

build:
	docker build --platform linux/amd64 -t $(IMAGE):$(TAG) .

push:
	docker push $(IMAGE):$(TAG)

run:
	docker run -p 8080:8080 --env-file .env $(IMAGE):$(TAG)

k8s-secret:
	kubectl --context $(CLUSTER) -n $(NAMESPACE) create secret generic video-saver-env \
		--from-env-file=.env --dry-run=client -o yaml \
		| kubectl --context $(CLUSTER) apply -f -

k8s-deploy:
	kubectl --context $(CLUSTER) apply -f k8s/

deploy: build push k8s-secret k8s-deploy

.PHONY: build push run k8s-secret k8s-deploy deploy
