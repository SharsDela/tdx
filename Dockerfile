FROM golang:1.23-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /tdx-api ./cmd/server/

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

ENV TZ=Asia/Shanghai
ENV LISTEN=:8080
ENV POOL_SIZE=3
ENV NO_INTERACTIVE=1
# ENV HOSTS=124.71.187.122,122.51.120.217   # 自定义服务器地址(逗号分隔)
# ENV TDX_PORT=7709                          # 通达信服务器端口

COPY --from=builder /tdx-api /usr/local/bin/tdx-api

EXPOSE 8080

CMD ["tdx-api"]
