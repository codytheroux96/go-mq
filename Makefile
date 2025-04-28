# Format the code using gofmt
neat:
	gofmt -w .

# Build the project
build:
	go build -o bin/go-mq ./cmd/go_mq

# Run the application
run:
	go run ./cmd/go_mq/main.go

# Build and run the project
bnr:
	go build -o bin/go-mq ./cmd/go_mq && ./bin/go-mq

# Run test files
test:
	go test -v ./...