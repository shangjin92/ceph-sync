PROJDIR=$(dir $(realpath $(firstword $(MAKEFILE_LIST))))

# change to project dir so we can express all as relative paths
$(shell cd $(PROJDIR))

VERSION ?= $(shell scripts/git-version.sh)

REPO_PATH=github.com/shangjin92/ceph-sync
LD_FLAGS="-w -X $(REPO_PATH)/cmd.Version=$(VERSION)"

$(shell mkdir -p bin )

.PHONY: all
all: build

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags $(LD_FLAGS) -o $(PROJDIR)/bin/ceph-sync main.go

.PHONY: build-mac
build-mac:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags $(LD_FLAGS) -o $(PROJDIR)/bin/ceph-sync main.go
