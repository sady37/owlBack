# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /build

# 复制 owl-common（依赖模块）
COPY owl-common ./owl-common

# 复制 wisefido-data 的 go mod 文件
COPY wisefido-data/go.mod wisefido-data/go.sum ./wisefido-data/

# 进入 wisefido-data 目录
WORKDIR /build/wisefido-data

# 下载依赖（replace 指令会自动使用 ../owl-common）
RUN go mod download

# 复制 wisefido-data 源代码
COPY wisefido-data/ .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o wisefido-data ./cmd/wisefido-data

# 运行阶段
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata wget

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/wisefido-data/wisefido-data .

# 暴露端口
EXPOSE 8080

# 运行服务
CMD ["./wisefido-data"]
