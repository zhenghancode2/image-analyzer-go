# Image Analyzer Go

一个简单的容器镜像分析工具

## 功能特点

- 提供 RESTful API 接口
- 支持容器镜像分析

## 安装

1. 克隆仓库：

```bash
git clone https://github.com/yourusername/image-analyzer-go.git
cd image-analyzer-go
```

2. 安装依赖：

```bash
go mod download
```

## 构建

```bash
go build -o bin/image-analyzer .
```

## 使用方法

### 命令行模式

```bash
# 分析镜像
./bin/image-analyzer analyze nginx:latest -o report.json -f json --check-os --check-python --check-tools --commands "python,java,node"
```

### API 服务器模式

```bash
# 启动 API 服务器
./bin/image-analyzer server
```

服务器默认在 `:8080` 端口启动。

## API 端点

- `POST /api/v1/analyze` - 分析镜像
- `GET /api/v1/health` - 健康检查

## Makefile 使用说明

- 构建项目：

  ```bash
  make build
  ```

- 启动 server（默认模式）：

  ```bash
  make run
  # 或
  make run MODE=server
  ```

- 分析指定镜像：

  ```bash
  make run MODE=analyze IMAGE=镜像名
  # 例如
  make run MODE=analyze IMAGE=nginx:latest
  ```

- 清理构建产物：

  ```sh
  make clean
  ```

- 构建 Docker 镜像：
  ```bash
  make docker
  ```
