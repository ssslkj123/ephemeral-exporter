FROM golang:1.20.4-alpine3.18 as builder
WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -a -installsuffix cgo -ldflags="-w -s" -o /usr/local/bin/pod-ephemeral-storage-exporter main.go
RUN cd /usr/local/bin/ && apk update \
    && apk add xz binutils \
    && wget https://github.com/upx/upx/releases/download/v4.0.2/upx-4.0.2-amd64_linux.tar.xz \
    && xz -d upx-4.0.2-amd64_linux.tar.xz \
    && tar -xvf upx-4.0.2-amd64_linux.tar \
    && cp upx-4.0.2-amd64_linux/upx /usr/bin/upx \
    && chmod +x /usr/bin/upx \
    && strip --strip-unneeded pod-ephemeral-storage-exporter \
    && upx pod-ephemeral-storage-exporter \
    && chmod a+x pod-ephemeral-storage-exporter

FROM alpine:3.18
COPY --from=builder /usr/local/bin/pod-ephemeral-storage-exporter /pod-ephemeral-storage-exporter
ENTRYPOINT ["/pod-ephemeral-storage-exporter"]
# docker build --no-cache -t caocao-acr-registry.cn-hangzhou.cr.aliyuncs.com/caocao/pod-ephemeral-storage-exporter:v2 .