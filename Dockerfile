# 采用分段构建的方式
# 1. 构建基础镜像
FROM golang:1.22.4 AS builder
# 配置环境白能量
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=Linux \
    GOARCH=amd64

# 设置工作目录
WORKDIR /build

# 复制项目中的 go.mod 和 go.sum 文件
COPY go.mod .
COPY go.sum .
# 下载依赖
RUN go mod download
# 将源代码复制到工作目录
COPY . .
# 编译成二进制文件
RUN go build -o bluebell_app .

# 2. 构建镜像 -- 从 scratch 开始构建
FROM scratch
# 拷贝文件
COPY ./templates /templates
COPY ./static /static
COPY ./conf /conf
# 从上一个镜像中拷贝二进制文件到当前镜像
COPY --from=builder /build/bluebell_app /bluebell_app

# 暴露端口
EXPOSE 8888

# 需要运行的命令
ENTRYPOINT ["/bluebell_app", "--config ./conf/config.yaml"]