FROM golang:1.17.5

RUN go install github.com/cespare/reflex@latest

WORKDIR /app
COPY . .

CMD ["/bin/sh","docker-entrypoint.sh"]
