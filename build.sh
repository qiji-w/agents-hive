#!/bin/sh
# 本地开发：生成当前平台的 ./server（macOS/Linux 本机跑用这个）
# 交叉编译 Linux 镜像：sh build.sh linux  → ./hive-linux
set -e
cd "$(dirname "$0")"

case "${1:-local}" in
  local|"")
    go build -o server ./cmd/server
    echo "OK: ./server ($(go env GOOS)/$(go env GOARCH))"
    ;;
  linux)
    GOOS=linux GOARCH=amd64 go build -o hive-linux ./cmd/server
    echo "OK: ./hive-linux (linux/amd64)"
    ;;
  *)
    echo "用法: sh build.sh [local|linux]" >&2
    exit 1
    ;;
esac
