build:
	go build -o BangumiBuddy
unit:
	go test ./...
ci-lint:
	golangci-lint run -c .golangci.yml --exclude-use-default=false
generate:
	go generate ./...
ci: unit ci-lint build