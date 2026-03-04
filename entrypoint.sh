#!/bin/sh
set -e

./server -migrate -seed

exec ./server
