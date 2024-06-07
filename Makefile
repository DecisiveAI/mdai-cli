.PHONY: build
.SILENT: build
build: mdai

.SILENT: mdai
mdai:
	go build -o mdai main.go

.PHONY: docker-build
.SILENT: docker-build
docker-build:
	docker build -t mdai-cli:latest .

.PHONY: demo
.SILENT: demo
local: mdai
	./mdai install

.PHONY: docker-local
.SILENT: docker-local
docker-local: docker-build
	docker run --network host -v /var/run/docker.sock:/var/run/docker.sock -it --rm mdai-cli:latest engine demo

.PHONY: clean
.SILENT: clean
clean:
	rm mdai
	docker rmi -f mdai-cli:latest
