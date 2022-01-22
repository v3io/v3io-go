# To setup env locally on your dev box, run:
# $ export $(grep -v '^#' ./hack/test.env | xargs)
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
ifndef V3IO_CONTROLPLANE_IGZ_ADMIN_PASSWORD
		$(error V3IO_CONTROLPLANE_IGZ_ADMIN_PASSWORD is undefined)
endif
	@echo "All required env vars populated"

.PHONY: generate-capnp
generate-capnp:
	pushd ./pkg/dataplane/schemas/; ./build; popd

.PHONY: clean
clean:
	pushd ./pkg/dataplane/schemas/; ./clean; popd

.PHONY: test
test: check-env
	go test -race -tags unit -count 1 ./...

.PHONY: fmt
fmt:
	gofmt -s -w .

.PHONY: lint
lint:
	./hack/lint/install.sh
	./hack/lint/run.sh

.PHONY: test-controlplane
test-controlplane: check-env
	go test -test.v=true -race -tags unit -count 1 ./pkg/controlplane/...

.PHONY: build
build: clean generate-capnp lint test
