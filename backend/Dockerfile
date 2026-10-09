FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/fs ./cmd/api

FROM golang:1.26-alpine AS migrate
WORKDIR /src
COPY migrate-tool/go.mod migrate-tool/go.sum ./
RUN go install -tags='no_mysql,no_sqlite3,no_ydb,no_clickhouse,no_mssql,no_vertica,no_libsql,no_azuresql' github.com/pressly/goose/v3/cmd/goose
COPY migrations /src/migrations
CMD ["sh", "-c", "/go/bin/goose -dir /src/migrations postgres \"$DB_URL\" up"]

FROM alpine:3.21 AS runtime
RUN apk add --no-cache ca-certificates && adduser -D -u 10001 appuser
WORKDIR /app
COPY --from=build /out/fs /app/fs
COPY migrations /app/migrations
USER appuser
EXPOSE 8080 2525
CMD ["/app/fs"]
