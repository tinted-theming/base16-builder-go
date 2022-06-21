# Stage 1: Build the application
FROM golang:1.18 as builder

RUN go install github.com/goreleaser/goreleaser@v1.9.2

RUN mkdir /build

WORKDIR /src/base16-builder-go

ADD ./go.mod ./go.sum ./
RUN go mod download

ADD . ./
RUN git clone https://github.com/base16-project/base16-schemes.git schemes
RUN goreleaser build --single-target --snapshot --rm-dist -o /build/base16-builder-go

RUN ls /build

# Stage 2: Copy files and configure what we need
FROM debian:buster-slim

ADD entrypoint.sh /bin

# Copy the built binary into the container
COPY --from=builder /build /bin

ENTRYPOINT ["/bin/entrypoint.sh"]
