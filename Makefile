-include .env
export

IMAGE     ?= video-saver
TAG       ?= latest
NAMESPACE ?= video-saver

build:
	docker build --platform linux/amd64 -t $(IMAGE):$(TAG) .

push:
	docker push $(IMAGE):$(TAG)

run:
	docker run -p $(PORT):8080 --env-file .env $(IMAGE):$(TAG)

k8s-deploy:
	kubectl --context $(CLUSTER) apply -f k8s/

deploy: build push k8s-deploy

.PHONY: build push run k8s-deploy deploy
