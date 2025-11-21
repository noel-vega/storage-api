dev:
	@echo "Starting storage API in development mode..."
	PORT=5000 STORAGE_ROOT_PATH=./storage-dev air 

	@echo "Building storage-api..."
build:
	go build -o bin/storage-api main.go

fmt:
	go fmt ./...
