FROM golang:1.13-alpine3.10 AS builder

RUN apk add --update gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -a -o main .

FROM alpine:3.10

WORKDIR /app
ARG RUNTIME_DIR="/app/runtime"
ARG SERVER_PORT=9000
ARG LOG_DIR=${RUNTIME_DIR}/log
ARG DB_DIR=${RUNTIME_DIR}/db
RUN    mkdir -p ${LOG_DIR} \
    && mkdir -p ${DB_DIR}
ENV AE_LOG_PATH=${LOG_DIR}/main.log \
    AE_SQLITE_PATH=${DB_DIR}/main.db \
    AE_SERVER_PORT=${SERVER_PORT}

RUN apk --no-cache add ca-certificates 

COPY --from=builder /app/main .

EXPOSE ${SERVER_PORT}
VOLUME [${RUNTIME_DIR}]
CMD ["./main"]
