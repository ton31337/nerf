GOMINVERSION = 1.16
NERF_CMD_PATH = "./cmd/nerf"
NERF_API_CMD_PATH = "./cmd/nerf-api"
NERF_SERVER_CMD_PATH = "./cmd/nerf-server"
GO111MODULE = on
export GO111MODULE

GOVERSION := $(shell go version | awk '{print substr($$3, 3)}')
GOISMIN := $(shell expr "$(GOVERSION)" ">=" "$(GOMINVERSION)")
ifneq "$(GOISMIN)" "1"
$(error "go version $(GOVERSION) is not supported, upgrade to $(GOMINVERSION) or above")
endif

LDFLAGS = -X github.com/ton31337/nerf.OauthClientID=$(OAUTH_CLIENT_ID) \
		-X github.com/ton31337/nerf.OauthClientSecret=$(OAUTH_CLIENT_SECRET) \
		-X github.com/ton31337/nerf.OauthMasterToken=$(OAUTH_MASTER_TOKEN) \
		-X github.com/ton31337/nerf.OauthOrganization=$(OAUTH_ORGANIZATION) \
		-X github.com/ton31337/nerf.DNSAutoDiscoverZone=$(DNS_AUTODISCOVER_ZONE)

check:
	go fmt ./...
	go fix ./...
	go vet -v ./...
	go mod tidy
	golines -w ./
	golangci-lint run
proto:
	go build github.com/golang/protobuf/protoc-gen-go
	PATH="$(PWD):$(PATH)" protoc --go_out=plugins=grpc:. *.proto
	rm protoc-gen-go
linux-client:
	@-go build -ldflags "$(LDFLAGS)" -o ./nerf ${NERF_CMD_PATH}
	@-go build -ldflags "$(LDFLAGS)" -o ./nerf-api ${NERF_API_CMD_PATH}
darwin-client:
	@-go build -ldflags "$(LDFLAGS)" -o ./osx/Nerf.app/Contents/MacOS/nerf ${NERF_CMD_PATH}
	@-go build -ldflags "$(LDFLAGS)" -o ./nerf-api ${NERF_API_CMD_PATH}
server:
	@-go build -ldflags "$(LDFLAGS)" -o ./nerf-server ${NERF_SERVER_CMD_PATH}
clean:
	rm -f ./osx/Nerf.app/Contents/MacOS/*
	rm -f ./nerf
	rm -f ./nerf-api
	rm -f ./nerf-server

.DEFAULT_GOAL := linux-client
