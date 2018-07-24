# Colab Dataset 

Simple voting web app to help build a dataset for deep learning / ML based on user colaboration.

# Installation

To install fastgate, you can download the latest release binary from the [**Dowload page**](https://github.com/auyer/fastgate/releases/latest)
, or compile it from source with GO.

## Install Golang

If you need to install GO, please refer to the [golang.org](https://golang.org/dl/) Download Page, and follow instructions, or use a package manager (Most are very outdated). 

> For macOS users, I do recommend installing from homebrew. The mantainers are doing a amazing job keeping up with updates. Note that you still need to configure home path, but brew itself will teach you on how to do it.   Run : `brew install go`

## Run Colab Dataset from Source

Add all the picture folders to the folder configured in the model file ( ./static/ is the default )

Run with:

```bash
go get github.com/auyer/colab-dataset
cd $GOPATH/src/github.com/auyer/colab-dataset
go run main.go
```
OR

```bash
go run main.go -config ./path_to_config_file
```
  A sample to the configuration file can be found in [config.model.json](config.model.json)


<!-- # Deploy with Docker

To run with docker, it would be necessary to add the pictures to a shared volume

By default, the Dockerfile picks the configuration file, TLS key and TLS cert from the same folder as the sourcecode.
```sh
  docker build -t fastgate .
  docker run -p YOUR_HTTP:8000 -p YOUR_HTTPS:8443 -d fastgate
``` -->
