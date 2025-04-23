# 构建阶段
FROM golang:1.24-bullseye AS builder

WORKDIR /app

# 安装gpgme和sqlite3开发依赖
RUN apt-get update && apt-get install -y libgpgme-dev libsqlite3-dev && rm -rf /var/lib/apt/lists/*

# 拷贝 go.mod 和 go.sum 并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 拷贝项目源代码
COPY . .

# 构建可执行文件
RUN go build -o image-analyzer-go main.go

# 运行阶段
FROM debian:bullseye-slim

WORKDIR /app

# 运行时也需要gpgme和sqlite3库
RUN apt-get update && apt-get install -y libgpgme11 libsqlite3-0 && rm -rf /var/lib/apt/lists/*

# 拷贝可执行文件和配置文件
COPY --from=builder /app/image-analyzer-go .
# 如有默认配置文件
COPY config.yaml .

# 设置容器启动命令
CMD ["./image-analyzer-go"]
