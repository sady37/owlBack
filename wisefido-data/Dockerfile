# 构建阶段。context 须为 owlBack 根目录，且含 owl-common、wisefido-qinglan、wisefido-sensor
FROM golang:1.24-alpine AS builder

WORKDIR /build

# 复制本地 replace 依赖（与 go.mod replace 一致）
COPY owl-common ./owl-common
COPY wisefido-qinglan ./wisefido-qinglan
COPY wisefido-sensor ./wisefido-sensor

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
