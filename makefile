.PHONY: build push compose deploy update

IMAGE_NAME = apaluch/gozurite
TAG = latest

build:
	docker build -t $(IMAGE_NAME):$(TAG) .

push:
	docker push $(IMAGE_NAME):$(TAG)

update: build push