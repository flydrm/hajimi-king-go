# 构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的工具
RUN apk add --no-cache git ca-certificates

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o hajimi-king cmd/app/main.go

# 运行阶段
FROM alpine:latest

# 安装ca-certificates用于HTTPS请求
RUN apk --no-cache add ca-certificates tzdata

# 设置工作目录
WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/hajimi-king .

# 复制配置文件示例
COPY --from=builder /app/queries.example .
COPY --from=builder /app/.env.example .

# 创建数据目录
RUN mkdir -p data

# 暴露端口（如果需要）
EXPOSE 8080

# 设置环境变量
ENV TZ=Asia/Shanghai

# 运行应用
CMD ["./hajimi-king"]