FROM golang:1.16-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY ./ ./

RUN go build -o /docker-api-ping

EXPOSE 8000

CMD [ "/docker-api-ping" ]