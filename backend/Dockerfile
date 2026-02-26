# syntax=docker/dockerfile:1

FROM golang:1.25.7

WORKDIR /usr/src/app

COPY backend/go.mod backend/go.sum ./
RUN go mod download

ENV IS_DOCKER=true

COPY backend/ ./
COPY swagger/ ./swagger/

RUN go build -v -o /usr/local/bin/app ./

EXPOSE 8080

CMD ["app"]