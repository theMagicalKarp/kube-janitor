FROM golang:1.11.0-stretch as build

RUN go get github.com/Masterminds/glide && \
    cd /go/src/github.com/Masterminds/glide && \
    git checkout tags/v0.13.1 && \
    go install .

WORKDIR /go/src/github.com/theMagicalKarp/kube-janitor
COPY glide.yaml glide.yaml
COPY glide.lock glide.lock
RUN glide install

COPY main.go main.go

RUN ["/bin/bash", "-c", "diff -u <(echo -n) <(gofmt -d -s main.go)"]
ENV CGO_ENABLED=0

RUN go build -o kube-janitor main.go

FROM scratch
COPY --from=build /go/src/github.com/theMagicalKarp/kube-janitor/kube-janitor /kube-janitor
ENTRYPOINT ["/kube-janitor"]
