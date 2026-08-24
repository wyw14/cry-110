FROM golang:1.26.2

ENV GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local
WORKDIR /src

COPY go.mod go.sum ./
COPY vendor ./vendor
COPY cmd ./cmd
COPY internal ./internal

RUN go build -mod=vendor -o /usr/local/bin/windlock ./cmd/windlock

EXPOSE 21210
CMD ["windlock", "-listen", "0.0.0.0:21210", "-data", "/var/lib/windlock"]
