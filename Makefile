.PHONY: check-env
check-env:
ifndef V3IO_DATAPLANE_URL
		$(error V3IO_DATAPLANE_URL is undefined)
endif
ifndef V3IO_DATAPLANE_USERNAME
		$(error V3IO_DATAPLANE_USERNAME is undefined)
endif
ifndef V3IO_DATAPLANE_ACCESS_KEY
		$(error V3IO_DATAPLANE_ACCESS_KEY is undefined)
endif
ifndef V3IO_CONTROLPLANE_URL
		$(error V3IO_CONTROLPLANE_URL is undefined)
endif
ifndef V3IO_CONTROLPLANE_USERNAME
		$(error V3IO_CONTROLPLANE_USERNAME is undefined)
endif
ifndef V3IO_CONTROLPLANE_PASSWORD
		$(error V3IO_CONTROLPLANE_PASSWORD is undefined)
endif

.PHONY: generate-capnp
generate-capnp:
	pushd ./pkg/dataplane/schemas/; ./build; popd

.PHONY: clean
clean:
	pushd ./pkg/dataplane/schemas/; ./clean; popd

.PHONY: test
test: check-env
	GO111MODULE=on \
		go test -race -tags unit -count 1 ./...

.PHONY: lint
lint:
	docker run --rm \
		--volume ${shell pwd}:/go/src/github.com/v3io/v3io-go \
		--env GOPATH=/go \
		--env GO111MODULE=off \
		golang:1.12 \
		bash /go/src/github.com/v3io/v3io-go/hack/lint.sh

	@echo Done.

.PHONY: build
build: clean generate-capnp lint test