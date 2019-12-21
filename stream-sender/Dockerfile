# Stage 1: build executable
FROM golang:1.13-alpine AS builder

WORKDIR /root

# Git is required for fetching the dependencies.
RUN apk add --no-cache git gcc musl-dev sqlite
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

# Fetch dependencies first; they are less susceptible to change on every build
# and will therefore be cached for speeding up the next build
COPY ./go.mod ./go.sum ./
RUN go mod download 

# Import the code
COPY ./ ./ 

# Build executable
RUN go build -o /streamsender .

# Stage 2: run container
FROM golang:1.13-alpine AS runtime
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

ARG HTTP_PORT=5000

COPY --from=builder /streamsender /streamsender

EXPOSE ${HTTP_PORT}

ENTRYPOINT ["/streamsender"]