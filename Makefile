.PHONY: lint
lint:
	@echo Installing linters...
	go get -u github.com/pavius/impi/cmd/impi
	go get -u gopkg.in/alecthomas/gometalinter.v2
	@$(GOPATH)/bin/gometalinter.v2 --install

	@echo Verifying imports...
	$(GOPATH)/bin/impi \
        --local github.com/v3io/v3io-go \
        --scheme stdLocalThirdParty \
        ./cmd/... ./controlplane/... ./internal/...

	@echo Linting...
	@$(GOPATH)/bin/gometalinter.v2 \
		--deadline=300s \
		--disable-all \
		--enable-gc \
		--enable=deadcode \
		--enable=goconst \
		--enable=gofmt \
		--enable=golint \
		--enable=gosimple \
		--enable=ineffassign \
		--enable=interfacer \
		--enable=misspell \
		--enable=staticcheck \
		--enable=unconvert \
		--enable=varcheck \
		--enable=vet \
		--enable=vetshadow \
		--enable=errcheck \
		--exclude="_test.go" \
		--exclude="comment on" \
		--exclude="error should be the last" \
		--exclude="should have comment" \
		--skip=pkg/platform/kube/apis \
		--skip=pkg/platform/kube/client \
		./cmd/... ./controlplane/... ./internal/...

	@echo Done.
