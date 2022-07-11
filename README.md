#!/usr/bin/env bash

set -e

cd ~
git clone https://github.com/luciancurteanu/go
cd golang

python3 install.py --path=/home/go/go --user=go