FROM golang

RUN apt-get update && \
    apt-get install sqlite3

WORKDIR /app
COPY . .

RUN make amd64
