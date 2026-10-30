# Etapa 1: compilación del binario de Go.
FROM golang:1.26-alpine AS build

WORKDIR /src

# Se copian primero solo las dependencias para aprovechar la caché de capas de Docker.
COPY go.mod go.sum ./
RUN go mod download

# Se copia el resto del código y se compila un binario estático.
COPY . .
RUN CGO_ENABLED=0 go build -o /out/api .

# Etapa 2: imagen final mínima, solo con el binario.
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

# El binario ya trae todo lo necesario; no se copia el código fuente.
COPY --from=build /out/api /app/api

# Puerto en el que escucha la API (ver main.go).
EXPOSE 3001

CMD ["/app/api"]
