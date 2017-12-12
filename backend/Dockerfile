FROM golang:1.11 as builder
WORKDIR /go/src
COPY . .
RUN CGO_ENABLED=0 go build -a --installsuffix cgo --ldflags="-s" -o whoami

# Create a minimal container to run a Golang static binary
FROM scratch
COPY --from=builder /go/src/ .
ENTRYPOINT ["/whoami"]
EXPOSE 80
