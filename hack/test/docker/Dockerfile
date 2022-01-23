FROM gcr.io/iguazio/golang:1.14

ENV GOPATH=/go
WORKDIR /v3io-go


RUN apt-get update --allow-releaseinfo-change \
    && apt-get install -y capnproto \
    && apt-get clean

# copy `go.mod` for definitions and `go.sum` to invalidate the next layer
# in case of a change in the dependencies
COPY go.mod go.sum ./

# copy everything
COPY . .

RUN ./hack/lint/install.sh
