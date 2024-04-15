FROM golang:1.21.3-alpine3.18 as build
LABEL authors="Vorontsov Ilya"
WORKDIR /Auth
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o build

FROM alpine:3.18 as prod
COPY --from=build /Auth .
EXPOSE 8080
ENTRYPOINT ["/build"]