FROM golang:1.24-alpine AS build
ENV GOTOOLCHAIN=local
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/hello .

FROM alpine:3.20
RUN adduser -D -u 10001 app
COPY --from=build /out/hello /app/hello
USER app
ENV PORT=4380 \
    TZ=Europe/Istanbul
EXPOSE 4380
ENTRYPOINT ["/app/hello"]
