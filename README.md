#!/usr/bin/env bash

set -e

cd ~
git clone https://github.com/luciancurteanu/golang
cd golang

python3 install.py --path=/home/golang/go --user=go
