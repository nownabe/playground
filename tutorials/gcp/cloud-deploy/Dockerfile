FROM golang:1.21-bookworm AS builder

ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOARCH amd64

WORKDIR /go/src/app

COPY . .

# -a: force rebuilding of packages that are already up-to-date
# -trimpath: remove all file system paths from the resulting executable
RUN go build -a -trimpath -o /go/bin/app .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /go/bin/app /app
ENTRYPOINT ["/app"]