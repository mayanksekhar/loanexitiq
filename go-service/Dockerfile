# ---- Build stage ----
FROM golang:1.26-alpine AS build

WORKDIR /src

# Cache dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a static binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/loanexitiq .

# ---- Runtime stage ----
FROM cgr.dev/chainguard/static:latest

WORKDIR /app

# The binary, plus the static assets served from disk at runtime
COPY --from=build /out/loanexitiq /app/loanexitiq
COPY --from=build /src/static /app/static

EXPOSE 8080

ENTRYPOINT ["/app/loanexitiq"]
