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
ENV GOOS=linux

RUN GOARCH=arm go build -o bin/kube-janitor-arm main.go
RUN GOARCH=arm64 go build -o bin/kube-janitor-arm64 main.go
RUN GOARCH=386 go build -o bin/kube-janitor-386 main.go
RUN GOARCH=amd64 go build -o bin/kube-janitor-amd64 main.go

FROM scratch
COPY --from=build /go/src/github.com/theMagicalKarp/kube-janitor/bin /
ENTRYPOINT ["/kube-janitor-amd64"]
