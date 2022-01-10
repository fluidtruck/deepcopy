##
# Stage 1 - Install dependencies
##
FROM golang:1.16-alpine AS build

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Configure git
RUN --mount=type=secret,id=github_token \
  git config --global url."https://$(cat /run/secrets/github_token):x-oauth-basic@github.com/fluidshare/".insteadOf "https://github.com/fluidshare/" && \
  git config --global url."https://$(cat /run/secrets/github_token):x-oauth-basic@github.com/fluidtruck/".insteadOf "https://github.com/fluidtruck/"

# Install go modules
COPY go.mod go.sum ./
RUN go mod verify
RUN go mod download

# Build the binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o bin/deepcopy

##
# Stage 2 - Final Image
##
FROM alpine:3.14

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy the binary
COPY --from=build /app/bin/deepcopy ./

# Wrap the binary in an entrypoint
COPY docker-entrypoint.sh ./
ENTRYPOINT ["./docker-entrypoint.sh"]
CMD ["./deepcopy"]
