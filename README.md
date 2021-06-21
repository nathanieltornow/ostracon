# ostracon

Ostracon is a shared-log system build on the persistence-first strategy introduced 
by [Scalog](https://www.usenix.org/conference/nsdi20/presentation/ding).

## Quickstart

Prerequisites: [Go](https://golang.org/) (version >= 1.15) and 
[Make](https://www.gnu.org/software/make/)

1. Clone this repository and go in the root-folder.
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

# Print usage
./ostracon --help
```

## Configuration

The cluster can be configured by modifying `cluster.config.yaml`, where a list of 
shards (sequencer-shards and record-shards) are defined, which will be started on
`make start-cluster`.

**Important:** Don't modify the system at runtime of a cluster.

A shard is described using the following properties:

```text
type:       If it is a record- or a sequencer-shard (sequencer or record).
ip:         The IP-address, where the shard should be served.
parent_ip:  The IP-address of the overlying shard.
root:       If the shard is the root of the system (only for sequencer-shards).
            Exactly one shard has to be root!
color:      Distinct color, that a shard represents (only for sequencer-shards).
interval:   The intervall with which order-requests should be sent to parents.
            E.g. 1ms, 10s, 900us, ...
disk:       Path to where the log should be stored (only for record-shards).
            Has to be distinct in order to prevent overriding!
```
