.PHONY: build proto test clean run stop

# Generate protobuf stubs for Python and Go
proto:
	@echo "Generating protobuf stubs..."
	# Python stubs
	python -m grpc_tools.protoc \
		-Iproto \
		--python_out=physics_environment/proto \
		--grpc_python_out=physics_environment/proto \
		proto/dsn.proto
	@touch physics_environment/proto/__init__.py
	# Go stubs (requires protoc-gen-go and protoc-gen-go-grpc)
	@mkdir -p array_controller/pkg/proto
	protoc \
		-Iproto \
		--go_out=array_controller/pkg/proto --go_opt=paths=source_relative \
		--go-grpc_out=array_controller/pkg/proto --go-grpc_opt=paths=source_relative \
		proto/dsn.proto || echo "Note: Go proto generation requires protoc-gen-go plugins"

# Build all services
build: proto
	docker-compose build

# Run everything
run:
	docker-compose up -d

# Stop everything
stop:
	docker-compose down

# Run Python tests
test-python:
	cd physics_environment && python -m pytest tests/ -v

# Run Go tests
test-go:
	cd array_controller && go test ./... -v

# Run all tests
test: test-python test-go
	@echo "All tests passed!"

# Clean up
clean:
	docker-compose down -v
	rm -rf physics_environment/proto/dsn_pb2*.py
	rm -rf array_controller/pkg/proto/*.go

# Development: run services locally
dev:
	docker-compose up --build

# Lint
lint:
	cd physics_environment && python -m flake8 app/ --max-line-length=120 || true
	cd array_controller && go vet ./...
