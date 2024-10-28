FROM golang:1.23.1

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./

COPY ./static ./static
COPY ./templates ./templates

EXPOSE 8080

ENTRYPOINT ["go"]
CMD ["test"]
