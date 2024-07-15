# MDAI CLI
MDAI command line tool that allows to install, update and manage MDAI clusters locally.

Prerequisites to build:
- [go](https://go.dev/doc/install)
- [docker](https://docs.docker.com/engine/install/)
- access to https://github.com/DecisiveAI/opentelemetry-operator 

Prerequisites to run local cluster:
- [docker](https://docs.docker.com/engine/install/)


# Build binary
```shell
make build
```

# Build docker image
```shell
make docker-build
```

# Local install
## go run
```shell
go mod vendor
GOPRIVATE=github.com/decisiveai/opentelemetry-operator go run main.go install
```


## binary
```shell
./mdai install
```

## docker
```shell
docker run --network host -v /var/run/docker.sock:/var/run/docker.sock -it --rm mdai-cli:latest install
```

# Remove kind cluster
```shell
kind delete cluster -n mdai-local
```
 ------

# Usage docs

See [Usage Docs Guide](https://github.com/DecisiveAI/mdai-cli/blob/main/docs/md/mdai.md)
