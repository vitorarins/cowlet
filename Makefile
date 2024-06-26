TARGETS:=$(shell ls -d cmd/* 2>/dev/null)
VERSION:=latest

CGO:=CGO_ENABLED=0
GO_LDFLAGS:=-w -s -extldflags '-static'

REGISTRY:=vitorarins

TAG:=$(REGISTRY)/cowlet:$(VERSION)

all: $(TARGETS)

$(TARGETS): ## Build all applications from cmd/
	cd $@ && $(CGO) go build -ldflags "$(GO_LDFLAGS)"

docker: ## Build docker image from build/docker
	docker build -t $(TAG) -f build/docker/Dockerfile .

push:
	docker push $(TAG)

.PHONY: $(TARGETS) docker
