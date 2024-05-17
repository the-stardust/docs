FROM golang:1.19 as builder
WORKDIR /go/src/proj/
COPY . .
RUN export GOPROXY=https://goproxy.io,direct && GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -tags=jsoniter -a --ldflags="-s" -o main .

FROM registry.cn-zhangjiakou.aliyuncs.com/xintujing/xtj-alpine:0.0.2
WORKDIR /root/
COPY --from=builder /go/src/proj/main .
ENV GO_ENV production
EXPOSE 8040:8040
CMD ["./main"]