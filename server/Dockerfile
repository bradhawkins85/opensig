# ---------- build ----------
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build     go build -o /out/opensig-api ./cmd/opensig-api &&     go build -o /out/opensig-relay ./cmd/opensig-relay

# ---------- api runtime ----------
FROM alpine:3.20 AS api
WORKDIR /app
COPY --from=build /out/opensig-api /app/opensig-api
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/app/opensig-api"]

# ---------- relay runtime ----------
FROM alpine:3.20 AS relay
WORKDIR /app
COPY --from=build /out/opensig-relay /app/opensig-relay
EXPOSE 2525
ENTRYPOINT ["/app/opensig-relay"]
