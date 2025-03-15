FROM golang:1.20-alpine AS builder

WORKDIR /app

# Copier les fichiers de dépendances et télécharger les dépendances
COPY go.mod go.sum* ./
RUN go mod download

# Copier le reste du code source
COPY . .

# Compiler l'application
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

# Image finale plus légère
FROM alpine:latest

WORKDIR /app

# Installer les certificats pour HTTPS
RUN apk --no-cache add ca-certificates

# Copier l'exécutable compilé depuis l'étape de build
COPY --from=builder /app/server .

# Exposer le port que l'application utilise
EXPOSE 8080

# Exécuter l'application
CMD ["./server"]
