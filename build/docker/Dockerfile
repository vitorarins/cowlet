ARG GO_MODULE_NAME=github.com/vitorarins/cowlet

### Prepare the build image
FROM golang:1.22 as build
ARG GO_MODULE_NAME

WORKDIR /go/src/${GO_MODULE_NAME}
COPY . ./
RUN make

### Build the target image
FROM scratch
ARG GO_MODULE_NAME

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /go/src/${GO_MODULE_NAME}/cmd/cowlet/cowlet .

CMD ["./cowlet"]

