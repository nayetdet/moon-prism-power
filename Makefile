.PHONY: mpp mpp-auth format check install-hooks

mpp:
	go run ./cmd/mpp

mpp-auth:
	go run ./cmd/mpp-auth

format:
	gofmt -w $$(rg --files -g '*.go')

check:
	@test -z "$$(gofmt -l $$(rg --files -g '*.go'))" || (echo "Go files are not formatted. Run 'make format'." && exit 1)
	go test ./...
	go vet ./...
	golangci-lint run

install-hooks:
	git config core.hooksPath .githooks
