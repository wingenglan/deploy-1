FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /flowboard-api .

FROM alpine:3.20
COPY --from=build /flowboard-api /flowboard-api
EXPOSE 8081
ENTRYPOINT ["/flowboard-api"]
