# MDAI CLI

MDAI command line tool that allows to install, update and manage MDAI clusters locally.

## Prerequisite
Since we are using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/) cluster you have to have [Docker](https://docs.docker.com/engine/install/) installed.

## Download and install MDAI CLI 
### Install prebuilt binary via shell script
This command will download and install release v.0.1.0, you can check and download the latest release [here](https://github.com/DecisiveAI/mdai-cli/releases).
```shell
curl --proto '=https' --tlsv1.2 -LsSf https://github.com/decisiveai/mdai-cli/releases/download/v0.1.0/mdai-installer.sh | sh
```
### Install via Homebrew
To install prebuilt binaries via homebrew
```Shell
brew install decisiveai/tap/mdai
```
