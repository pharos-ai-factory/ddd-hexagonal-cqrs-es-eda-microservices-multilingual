FROM golang:1.27.1-alpine AS build
WORKDIR /src
COPY services/storefront/go.mod services/storefront/go.sum ./
RUN go mod download
COPY services/storefront .
RUN CGO_ENABLED=0 go build -trimpath -o /out/ ./apps/storefront ./apps/topology ./apps/replay ./integrations/delivery/simulator

FROM alpine:3.23
RUN addgroup -g 10001 reference && adduser -D -u 10001 -G reference reference && mkdir /data && chown reference:reference /data
ARG SERVICE
COPY --from=build /out/${SERVICE} /app/service
USER 10001:10001
ENTRYPOINT ["/app/service"]
