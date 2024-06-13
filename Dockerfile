FROM golang:1.21.0-alpine3.17 AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM scratch

COPY --from=build /app/main /main

USER 1001

EXPOSE 8080

CMD ["/main"]
