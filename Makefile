# Set the shell to bash always
SHELL := /bin/bash

build-examples:
	( cd examples/stack/crank-stack-gcp-dep/v0.0.1 && docker build . -t hasheddan/crank-stack-gcp-dep:v0.0.1 )
	( cd examples/stack/crank-stack-intermediate-one/v0.0.1 && docker build . -t hasheddan/crank-stack-intermediate-one:v0.0.1 )
	( cd examples/stack/crank-stack-intermediate-two/v0.0.1 && docker build . -t hasheddan/crank-stack-intermediate-two:v0.0.1 )
	( cd examples/stack/crank-stack-root/v0.0.1 && docker build . -t hasheddan/crank-stack-root:v0.0.1 )

push-examples:
	docker push hasheddan/crank-stack-gcp-dep:v0.0.1
	docker push hasheddan/crank-stack-intermediate-one:v0.0.1
	docker push hasheddan/crank-stack-intermediate-two:v0.0.1
	docker push hasheddan/crank-stack-root:v0.0.1

create-cluster:
	kind create cluster --name=crank
	kubectl apply -f config/crd/

delete-cluster:
	kind delete cluster --name=crank

manager:
	go run main.go manager

generate:
	go generate ./...

.PHONY: build-examples push-examples create-cluster delete-cluster manager generate