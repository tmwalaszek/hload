FROM golang:1.22.2

RUN apt-get update && \
    apt-get install sqlite3

WORKDIR /app
COPY . .

RUN make amd64
