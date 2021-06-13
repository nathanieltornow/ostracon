#!/bin/bash

set -e

sudo wget --continue --quiet https://golang.org/dl/go1.15.8.linux-amd64.tar.gz

sudo tar -C /usr/local -xzf go1.15.8.linux-amd64.tar.gz

export PATH=$PATH:/usr/local/go/bin

sudo sh -c  "echo 'export PATH=\$PATH:/usr/local/go/bin' >> /etc/profile"