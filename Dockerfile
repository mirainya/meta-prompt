FROM node:20-alpine AS frontend
WORKDIR /app/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.26-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/web/dist ./cmd/server/dist
RUN CGO_ENABLED=0 go build -o /meta-prompt ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
COPY --from=backend /meta-prompt /usr/local/bin/meta-prompt
WORKDIR /app
EXPOSE 8080
CMD ["meta-prompt"]
