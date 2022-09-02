FROM golang

ENV PG_DNS "postgres://postgres:postgres@localhost:5432/postgres?replication=database"
ENV WEBHOOK_SERVICE_URL "http://localhost:8080"

COPY . /go/src/github.com/movinglake/pg_webhook
WORKDIR /go/src/github.com/movinglake/pg_webhook
RUN go build -o pg_webhook
CMD ["/go/src/github.com/movinglake/pg_webhook/pg_webhook"]