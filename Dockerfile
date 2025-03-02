FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

# Add git, curl and upx support
RUN apk add --no-cache git curl upx ca-certificates 

WORKDIR /src

# Pull modules
COPY go.* ./
RUN go mod download

# Copy code into image
COPY . ./

# Build application for deployment
RUN --mount=type=cache,target=/root/.cache/go-build \
  --mount=type=cache,target=/go/pkg \
  CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags '-s -w' -o /tmp/chatter .

# Compress binary
RUN upx --best --lzma /tmp/chatter

# Create minimal image
FROM --platform=$TARGETPLATFORM gcr.io/distroless/base

# Add the binary
COPY --from=builder /tmp/chatter /chatter

EXPOSE 80/tcp

ENTRYPOINT ["/chatter"]
