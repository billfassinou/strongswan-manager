# Build multi-stage : SPA React (node) -> binaire Go statique -> image minimale.
FROM node:22-alpine AS web
WORKDIR /web
COPY web/package.json web/package-lock.json* ./
RUN npm install
COPY web/ ./
RUN npm run build

FROM golang:1.23-alpine AS build
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum* ./
RUN go mod download || true
COPY . .
# Récupère le build du front (dist/) produit par le stage node.
COPY --from=web /web/dist ./web/dist
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.20
RUN adduser -D -u 10001 app
COPY --from=build /out/server /usr/local/bin/server
USER app
EXPOSE 7926
ENTRYPOINT ["/usr/local/bin/server"]
