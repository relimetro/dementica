FROM golang:1.25
# requires /usr/src/app/firebase.json from docker-compose

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -v -o /usr/local/bin/app ./server.go

CMD ["app"]
