# Build-Phase
FROM golang:1.21-alpine AS builder
LABEL org.opencontainers.image.source="https://github.com/codalf/microzoo"

WORKDIR /app

# Abhängigkeiten herunterladen
COPY go.mod .
COPY go.sum .
RUN go mod download

# Quellcode kopieren
COPY src src

# Build
RUN go build -o /go-service src/main.go

# Finales Image
FROM alpine:latest
LABEL org.opencontainers.image.source="https://github.com/codalf/microzoo"

# Installiere netcat für den wait-for-Befehl
RUN apk --no-cache add netcat-openbsd

WORKDIR /app

# Kopiere die ausführbare Datei und die util-Skripte
COPY --from=builder /go-service .
COPY util/wait-start .
COPY util/wait-for .

# Berechtigungen setzen
RUN chmod +x /app/go-service && chmod +x /app/wait-start && chmod +x /app/wait-for

EXPOSE 8080

# ENTRYPOINT ist das wait-start Skript, wie im spring-boot-service
ENTRYPOINT ["/app/wait-start"]
