GOMINVERSION = 1.16
NERF_CMD_PATH = "./cmd/nerf"
NERF_SERVER_CMD_PATH = "./cmd/nerf-server"
GO111MODULE = on
export GO111MODULE

GOVERSION := $(shell go version | awk '{print substr($$3, 3)}')
GOISMIN := $(shell expr "$(GOVERSION)" ">=" "$(GOMINVERSION)")
ifneq "$(GOISMIN)" "1"
$(error "go version $(GOVERSION) is not supported, upgrade to $(GOMINVERSION) or above")
endif

vet:
	go vet -v ./...
proto:
	go build github.com/golang/protobuf/protoc-gen-go
	PATH="$(PWD):$(PATH)" protoc --go_out=plugins=grpc:. *.proto
	rm protoc-gen-go
bin:
	go build -o ./nerf ${NERF_CMD_PATH}
	go build -o ./nerf-server ${NERF_SERVER_CMD_PATH}
clean:
	rm -f ./nerf
	rm -f ./nerf-server

.DEFAULT_GOAL := bin
