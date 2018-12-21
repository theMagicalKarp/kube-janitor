FROM golang:1.11 as build

WORKDIR /src
COPY . /src
ENV CGO_ENABLED=0
RUN go test -v ./...
RUN go build -o /src/out/kube-janitor

FROM scratch
COPY --from=build /src/out/kube-janitor /kube-janitor
ENTRYPOINT ["/kube-janitor"]
