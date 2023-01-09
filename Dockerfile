FROM golang:1.19 as builder
WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./...
RUN apt-get update \
	&& apt-get clean \
	&& rm -rf /var/lib/apt/lists/* \
	&& curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
	| sh -s v1.50.1
RUN make build-webserver

FROM gcr.io/distroless/static-debian11 as app
WORKDIR /app
COPY --from=builder --chown=1000:1000 /go/src/app/target/webserver webserver
USER 1000:1000
