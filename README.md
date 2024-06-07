# build binary
    go build -o mdai main.go

# build docker image
    docker build -t mdai-cli:latest .

# local install
## go run
    go run main.go install

## binary
    ./mdai install

## docker
    docker run --network host -v /var/run/docker.sock:/var/run/docker.sock -it --rm mdai-cli:latest install

# remove kind cluster
    kind delete cluster -n mdai-local
