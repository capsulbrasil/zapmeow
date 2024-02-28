FROM golang:1.20-alpine

RUN apk add --no-cache gcc musl-dev
RUN apk add mailcap

WORKDIR /app

COPY . .

RUN ls -la

RUN go mod download

ENV CGO_ENABLED=1
ENV CGO_CFLAGS="-D_LARGEFILE64_SOURCE"
RUN go build -o server cmd/server/main.go

EXPOSE 8900

CMD ["./server"]
