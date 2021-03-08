GOMINVERSION = 1.16
NERF_CMD_PATH = "./cmd/nerf"
GO111MODULE = on
export GO111MODULE

GOVERSION := $(shell go version | awk '{print substr($$3, 3)}')
GOISMIN := $(shell expr "$(GOVERSION)" ">=" "$(GOMINVERSION)")
ifneq "$(GOISMIN)" "1"
$(error "go version $(GOVERSION) is not supported, upgrade to $(GOMINVERSION) or above")
endif

LDFLAGS = -X github.com/ton31337/nerf.OauthClientID=$(OAUTH_CLIENT_ID) \
		-X github.com/ton31337/nerf.OauthClientSecret=$(OAUTH_CLIENT_SECRET) \
		-X github.com/ton31337/nerf.OauthMasterToken=$(OAUTH_MASTER_TOKEN)

ALL = linux-amd64 \
	windows-amd64 \
	darwin-amd64 \
	darwin-arm64

all: $(ALL:%=build/%/nerf)
build/%/nerf: .FORCE
	GOOS=$(firstword $(subst -, , $*)) \
		GOARCH=$(word 2, $(subst -, ,$*)) $(GOENV) \
		go build -ldflags "$(LDFLAGS)" -o $@ ${NERF_CMD_PATH}
build/windows-%: LDFLAGS += -H=windowsgui
check:
	go fmt ./...
	go fix ./...
	go vet -v ./...
	go mod tidy
	golangci-lint run
proto:
	go build github.com/golang/protobuf/protoc-gen-go
	PATH="$(PWD):$(PATH)" protoc --go_out=plugins=grpc:. *.proto
	rm protoc-gen-go
bin:
	go build -ldflags "$(LDFLAGS)" -o ./nerf ${NERF_CMD_PATH}
clean:
	rm -rf ./build
	rm -f ./nerf

.FORCE:
.DEFAULT_GOAL := bin
