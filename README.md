[![go](https://github.com/nathanieltornow/ostracon/actions/workflows/go.yml/badge.svg)](https://github.com/nathanieltornow/ostracon/actions/workflows/go.yml)

# ostracon

Ostracon is a shared-log system build on the persistence-first strategy introduced 
by [Scalog](https://www.usenix.org/conference/nsdi20/presentation/ding).

## Quickstart

Prerequisites: [Go](https://golang.org/) (version >= 1.15) and 
[Make](https://www.gnu.org/software/make/)

1. Clone this repository.
```shell
git clone https://github.com/nathanieltornow/ostracon.git && \
cd ostracon
```

2. Start a sample cluster.
```shell
make start-cluster
```

3. In another terminal, build the client-library.
```shell
make
```

4. Append and subscribe on the shared log. For example:
```shell
# Append the record "HelloWorld" to color 0
./ostracon -a -record HelloWorld -color 0 

# Subscribe to all records on color 1, starting from sequence-number 8
./ostracon -s -gsn 8 -color 1
```

- Appending
    ```shell
    ./ostracon -a -record HelloWorld -color 0 
    ```
- Subscribing
     ```shell
    ./ostracon -s -gsn  
    ```