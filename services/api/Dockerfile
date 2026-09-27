FROM golang:1.27.1-alpine AS build
WORKDIR /src
COPY services/api/go.mod services/api/go.sum ./
RUN go mod download
COPY services/api ./
RUN CGO_ENABLED=0 go build -trimpath -o /out/api ./apps/api
FROM alpine:3.23
RUN addgroup -g 10001 reference && adduser -D -u 10001 -G reference reference
COPY --from=build /out/api /app/api
USER 10001:10001
ENTRYPOINT ["/app/api"]
