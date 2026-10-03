FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN go mod tidy && CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /worker ./apps/worker/cmd/worker
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /worker /worker
USER nonroot:nonroot
ENTRYPOINT ["/worker"]
