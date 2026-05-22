FROM golang:1.19-alpine AS build
WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags "-s -w -extldflags '-static'" -o ./app
RUN apk add upx
RUN upx ./app

FROM scratch
COPY --from=build /build/app /app
ENV PG_DNS "postgres://postgres:postgres@localhost:5432/postgres?replication=database"
ENV WEBHOOK_SERVICE_URL "http://localhost:8080"

ENTRYPOINT ["/app"]