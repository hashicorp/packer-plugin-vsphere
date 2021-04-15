NAME=vsphere
BINARY=packer-plugin-${NAME}

COUNT?=1
TEST?=$(shell go list ./...)

.PHONY: dev

build:
	@go build -o ${BINARY}

dev: build
	@mkdir -p ~/.packer.d/plugins/
	@mv ${BINARY} ~/.packer.d/plugins/${BINARY}

#generate:
	#@go install github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@latest
	#@go generate -v ./...
	#@packer-sdc renderdocs -src content-files/docs -partials content-files/partials -dst docs/

run-example: dev
	@packer build ./example

test:
	@go test -count $(COUNT) $(TEST) -timeout=3m

testacc: dev
	@PACKER_ACC=1 go test -count $(COUNT) -v $(TEST) -timeout=120m