FROM golang:1.24-alpine

RUN apk add --no-cache git postgresql-client && \
    apk add --no-cache --virtual .build-deps curl tar && \
    curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/ && \
    apk del .build-deps

RUN go install github.com/air-verse/air@latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./entrypoint.sh /app/entrypoint.sh
COPY . .

ENTRYPOINT ["/app/entrypoint.sh"]

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]