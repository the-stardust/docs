FROM golang:1.19 as builder
WORKDIR /go/src/proj/
COPY . .
ENV SPEECHSDK_ROOT="/root/speechsdk"
RUN wget http://archive.ubuntu.com/ubuntu/pool/main/o/openssl/libssl1.1_1.1.1f-1ubuntu2_amd64.deb && dpkg -i libssl1.1_1.1.1f-1ubuntu2_amd64.deb
RUN apt-get update -y && apt-get install -y build-essential libssl-dev ca-certificates libasound2 \
    && mkdir -p "$SPEECHSDK_ROOT" \
    && wget -O SpeechSDK-Linux.tar.gz https://aka.ms/csspeech/linuxbinary \
    && tar --strip 1 -xzf SpeechSDK-Linux.tar.gz -C "$SPEECHSDK_ROOT" \
    && rm SpeechSDK-Linux.tar.gz  \
    && chmod -R 0777 "$SPEECHSDK_ROOT"
ENV CGO_CFLAGS="-I$SPEECHSDK_ROOT/include/c_api"
ENV CGO_LDFLAGS="-L$SPEECHSDK_ROOT/lib/x64 -lMicrosoft.CognitiveServices.Speech.core"
ENV LD_LIBRARY_PATH="$SPEECHSDK_ROOT/lib/x64:$LD_LIBRARY_PATH"

RUN export GOPROXY=https://goproxy.io,direct && GOARCH=amd64 GOOS=linux CGO_ENABLED=1 go build -tags=jsoniter -a --ldflags="-s" -o job ./tools/job/job.go

FROM ubuntu:focal-20230801
WORKDIR /root/
RUN echo "deb http://th.archive.ubuntu.com/ubuntu jammy main" >> /etc/apt/sources.list && apt update -y && apt install libc6 wget build-essential libssl-dev ca-certificates libasound2 -y
COPY --from=builder /go/src/proj/job .
COPY --from=builder /root/speechsdk/ .
RUN wget http://archive.ubuntu.com/ubuntu/pool/main/o/openssl/libssl1.1_1.1.1f-1ubuntu2_amd64.deb && dpkg -i libssl1.1_1.1.1f-1ubuntu2_amd64.deb
ENV SSL_CERT_DIR=/etc/ssl/certs
ENV GO_ENV production
ENV CGO_CFLAGS="-I/root/include/c_api"
ENV CGO_LDFLAGS="-L/root/lib/x64 -lMicrosoft.CognitiveServices.Speech.core"
ENV LD_LIBRARY_PATH="/root/lib/x64:$LD_LIBRARY_PATH"
RUN chmod -R 0755 ./

CMD ["./job"]