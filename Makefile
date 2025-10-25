.PHONY: help build-manager build-proxy build-all run-manager run-proxy clean

help:
	@echo "Available targets:"
	@echo "  build-manager    - Build bin-manager Docker image"
	@echo "  build-proxy      - Build bin-proxy Docker image"
	@echo "  build-all        - Build both Docker images"
	@echo "  run-manager      - Run bin-manager container"
	@echo "  run-proxy        - Run bin-proxy container"
	@echo "  clean            - Remove Docker images"

build-manager:
	docker build -f Dockerfile.bin-manager -t bin-manager:latest .

build-proxy:
	docker build -f Dockerfile.binproxy -t binproxy:latest .

build-all: build-manager build-proxy

run-manager:
	docker run -p 8081:8081 bin-manager:latest

run-proxy:
	docker run --name binproxy --net=host -d binproxy:latest

clean:
	docker rmi bin-manager:latest binproxy:latest