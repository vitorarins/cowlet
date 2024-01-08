TARGETS:=$(shell ls -d cmd/* 2>/dev/null)

GO_LDFLAGS:=-w -s -extldflags '-static'

all: $(TARGETS)

$(TARGETS): ## Build all applications from cmd/
	cd $@ && $(CGO) go build -ldflags "$(GO_LDFLAGS)"

.PHONY: $(TARGETS)