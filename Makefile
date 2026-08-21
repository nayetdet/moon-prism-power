.PHONY: mpp mpp-auth

mpp:
	go run ./cmd/mpp

mpp-auth:
	go run ./cmd/mpp-auth
