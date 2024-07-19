ifndef CARGO_DIST_TARGET
	GOARCH = $(shell uname -m)
	GOOS = $(shell uname -o | tr '[:upper:]' '[:lower:]')
else
ifeq ($(CARGO_DIST_TARGET),aarch64-apple-darwin)
	GOARCH = arm64
	GOOS = darwin
else
ifeq ($(CARGO_DIST_TARGET),x86_64-apple-darwin)
	GOARCH = amd64
	GOOS = darwin
else
ifeq ($(CARGO_DIST_TARGET),x86_64-unknown-linux-gnu)
	GOARCH = amd64
	GOOS = linux
else
ifeq ($(CARGO_DIST_TARGET),x86_64-unknown-linux-musl)
	GOARCH = amd64
	GOOS = linux
else
ifeq ($(CARGO_DIST_TARGET),x86_64-pc-windows-msvc)
	GOARCH = amd64
	GOOS = windows
else
$(error "unsupported target platform $(CARGO_DIST_TARGET)")
endif
endif
endif
endif
endif
endif


.PHONY: build
.SILENT: build
build: mdai

.PHONY: mdai
.SILENT: mdai
mdai:
	rm -f mdai
	go mod vendor
	CGO_ENABLED=0 go build -o mdai main.go

.PHONY: docker-build
.SILENT: docker-build
docker-build:
	go mod vendor
	docker build -t mdai-cli:latest .

.PHONY: install
.SILENT: install
install: mdai
	./mdai install

.PHONY: demo
.SILENT: demo
demo: mdai
	./mdai demo

.PHONY: docker-install
.SILENT: docker-install
docker-install: docker-build
	docker run --network host -v /var/run/docker.sock:/var/run/docker.sock -it --rm mdai-cli:latest install

.PHONY: docker-demo
.SILENT: docker-demo
docker-demo: docker-build
	docker run --network host -v /var/run/docker.sock:/var/run/docker.sock -it --rm mdai-cli:latest demo


.PHONY: clean
.SILENT: clean
clean:
	rm -f mdai
	docker rmi -f mdai-cli:latest &> /dev/null

.PHONY: ci-build
.SILENT: ci-build
ci-build: git-setup build

.PHONY: git-setup
.SILENT: git-setup
git-setup:
	git config --global url."https://user:${TOKEN}@github.com".insteadOf "https://github.com"

