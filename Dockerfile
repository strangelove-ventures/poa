# make local-image
# docker run --rm -it poa:local q

FROM golang:1.21-alpine3.18 as builder

RUN set -eux; apk add --no-cache git libusb-dev linux-headers gcc musl-dev make go;

ENV GOPATH=""

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN cd simapp && go mod download
RUN cd simapp && make build

FROM alpine:3.18

COPY --from=builder /go/simapp/build/* /bin/poad

ENTRYPOINT ["/bin/poad"]