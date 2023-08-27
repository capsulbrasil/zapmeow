FROM golang:1.20-alpine

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY . .

RUN ls -la

RUN go mod download

ENV CGO_ENABLED=1
RUN go build -o main .

EXPOSE 8900

CMD ["./main"]
