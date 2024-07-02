# build binary
    GOPRIVATE=github.com/decisiveai/opentelemetry-operator go build -o mdai main.go

# build docker image
    go mod vendor && docker build -t mdai-cli:latest .

# local install
## go run
    GOPRIVATE=github.com/decisiveai/opentelemetry-operator go run main.go install

## binary
    ./mdai install

## docker
    docker run --network host -v /var/run/docker.sock:/var/run/docker.sock -it --rm mdai-cli:latest install

# remove kind cluster
    kind delete cluster -n mdai-local
