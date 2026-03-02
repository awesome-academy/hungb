#!/bin/sh
set -e

./server -migrate -seed || true

exec ./server
