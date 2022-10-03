.PHONY: generate
generate:
	go get -d github.com/vektra/mockery/cmd/mockery
	go generate ./...
