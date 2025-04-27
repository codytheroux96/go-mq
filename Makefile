# Format the code using gofmt
neat:
	gofmt -w .

# Build the project
build:

# Run the application
run:
	go run ./cmd/go_mq/main.go

# Run test files
test:
	go test -v ./...