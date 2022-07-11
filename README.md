#!/usr/bin/env bash

set -e

cd ~
git clone https://github.com/luciancurteanu/goproxy
cd goproxy

python3 install.py --path=/home/app --user=goproxy
