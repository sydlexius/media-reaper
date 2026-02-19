# Stage 1: Build frontend
FROM node:22-bookworm-slim AS frontend
WORKDIR /build/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 2: Build backend with embedded frontend
FROM golang:1.25-bookworm AS backend
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /build/web/dist ./web/dist
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /media-reaper ./cmd/server/
RUN mkdir /data

# Stage 3: Final minimal image
FROM gcr.io/distroless/static-debian12
COPY --from=backend /media-reaper /media-reaper
COPY --from=backend --chown=65532:65532 /data /data
USER 65532:65532
ENV MEDIA_REAPER_DB_PATH=/data/media-reaper.db
EXPOSE 8080
VOLUME ["/data"]
ENTRYPOINT ["/media-reaper"]
