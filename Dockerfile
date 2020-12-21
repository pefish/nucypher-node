FROM golang:1.14
WORKDIR /app
ENV GO111MODULE=on
COPY ./ ./
RUN go get -u github.com/pefish/go-build-tool@v0.0.6
RUN make
ENV GO_CONFIG /app/config/pom.yaml
ENV GO_SECRET /app/secret/pom.yaml
CMD ["./build/bin/linux/main", "--help"]

# docker build -t pefish/nucypher-node:v0.0.7 .