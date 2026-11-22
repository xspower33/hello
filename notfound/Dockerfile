# --- build ---------------------------------------------------------------
# Build context: depo kökü (go build ./notfound)
FROM golang:1.24-alpine AS build
ENV GOTOOLCHAIN=local
WORKDIR /src
COPY . .
ARG TARGETOS=linux
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/notfound ./notfound

# --- run -----------------------------------------------------------------
FROM alpine:3.20
RUN adduser -D -u 10001 app
COPY --from=build /out/notfound /app/notfound
USER app
ENV PORT=4381 \
    TZ=Europe/Istanbul
EXPOSE 4381
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- "http://127.0.0.1:${PORT}/healthz" || exit 1
ENTRYPOINT ["/app/notfound"]
