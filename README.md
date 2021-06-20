[![go](https://github.com/nathanieltornow/ostracon/actions/workflows/go.yml/badge.svg)](https://github.com/nathanieltornow/ostracon/actions/workflows/go.yml)

# ostracon

Ostracon is a shared-log system build on the persistence-first strategy introduced 
by [Scalog](https://www.usenix.org/conference/nsdi20/presentation/ding).

## Quickstart

1. Clone this repository
```shell
git clone https://github.com/nathanieltornow/ostracon.git && \
cd ostracon
```
2. Install [Go](https://golang.org/), if you haven't already:
```shell
# For Ubuntu:
./scripts/install_go.sh && source ~/.profile
```

3. Install [goreman](https://github.com/mattn/goreman):
```shell
go get github.com/mattn/goreman
```

4. Run the ostracon-cluster:
```shell
goreman start
```