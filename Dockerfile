FROM pefish/ubuntu-go:v1.15 as builder
WORKDIR /app
ENV GO111MODULE=on
COPY ./ ./
RUN go get -u github.com/pefish/go-build-tool@v0.0.6
RUN make

FROM pefish/ubuntu18_04:v1.0
WORKDIR /app
COPY --from=builder /app/build/bin/linux/ /app/bin/
ENV GO_CONFIG /app/config/pom.yaml
ENV GO_SECRET /app/secret/pom.yaml
CMD ["/app/bin/main", "--help"]

# docker build -t pefish/nucypher-node:v0.0.12 .