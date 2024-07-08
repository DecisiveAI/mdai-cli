.PHONY: build
.SILENT: build
build: mdai

.SILENT: mdai
mdai:
	go mod vendor
	CGO_ENABLED=0 go build -o mdai main.go

.PHONY: docker-build
.SILENT: docker-build
docker-build:
	go mod vendor
	docker build -t mdai-cli:latest .

.PHONY: install
.SILENT: install
local: mdai
	./mdai install

.PHONY: demo
.SILENT: demo
local: mdai
	./mdai demo

.PHONY: docker-install
.SILENT: docker-install
docker-local: docker-build
	docker run --network host -v /var/run/docker.sock:/var/run/docker.sock -it --rm mdai-cli:latest install

.PHONY: docker-demo
.SILENT: docker-demo
docker-local: docker-build
	docker run --network host -v /var/run/docker.sock:/var/run/docker.sock -it --rm mdai-cli:latest demo


.PHONY: clean
.SILENT: clean
clean:
	rm mdai
	docker rmi -f mdai-cli:latest
