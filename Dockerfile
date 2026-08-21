# Build Stage
FROM golang:1.26-trixie AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/mpp ./cmd/mpp

# Runtime Stage
FROM gcr.io/distroless/static-debian13:nonroot AS runtime
COPY --from=build /out/mpp /usr/local/bin/mpp
WORKDIR /app
ENTRYPOINT ["mpp"]
