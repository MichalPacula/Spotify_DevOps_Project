FROM golang:1.23.1 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /go_website

FROM alpine:3.20 AS build-release-stage

WORKDIR /

COPY --from=build-stage /go_website /go_website

COPY ./static ./static
COPY ./templates ./templates

EXPOSE 8080

RUN addgroup -S nonrootuser && adduser -S nonrootuser -G nonrootuser

USER nonrootuser

ENTRYPOINT ["/go_website"]
