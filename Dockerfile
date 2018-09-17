FROM golang:1.11.0-stretch as build

RUN go get github.com/golang/dep && \
    cd /go/src/github.com/golang/dep/cmd/dep && \
    git checkout tags/v0.5.0 && \
    go install .

WORKDIR /go/src/github.com/theMagicalKarp/kube-janitor
COPY Gopkg.toml Gopkg.toml
COPY Gopkg.lock Gopkg.lock
RUN dep ensure --vendor-only

COPY ./*.go /go/src/github.com/theMagicalKarp/kube-janitor/

ENV CGO_ENABLED=0

RUN go tool vet -all *.go
RUN ["/bin/bash", "-c", "diff -u <(echo -n) <(gofmt -d -s *.go)"]
RUN go test -v github.com/theMagicalKarp/kube-janitor/...
RUN go build -o kube-janitor

FROM scratch
COPY --from=build /go/src/github.com/theMagicalKarp/kube-janitor/kube-janitor /kube-janitor
ENTRYPOINT ["/kube-janitor"]
