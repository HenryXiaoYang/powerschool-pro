FROM node:24-alpine AS web
RUN npm install -g pnpm@11.25.0
WORKDIR /web
COPY web/package.json web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY web/ .
RUN pnpm build

FROM golang:1.27-alpine AS server
WORKDIR /server
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# distroless/static carries CA certificates (HTTPS to the school) and tzdata (SCHOOL_TIMEZONE).
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=server /out/server /server
COPY --from=web /web/dist /web
ENV WEB_DIR=/web
EXPOSE 8080
ENTRYPOINT ["/server"]
