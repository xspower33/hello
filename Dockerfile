# --- build ---------------------------------------------------------------
FROM golang:1.24-alpine AS build
ENV GOTOOLCHAIN=local
WORKDIR /src
COPY . .
ARG TARGETOS=linux
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/hello .

# --- run -----------------------------------------------------------------
FROM alpine:3.20
RUN adduser -D -u 10001 app
COPY --from=build /out/hello /app/hello
USER app
ENV PORT=4380 \
    TZ=Europe/Istanbul
EXPOSE 4380
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- "http://127.0.0.1:${PORT}/healthz" || exit 1
ENTRYPOINT ["/app/hello"]
