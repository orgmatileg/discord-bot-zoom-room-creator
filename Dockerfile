############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder

# Install:
# - git (Required for fetching depedencies)
# - tzdata (Required for change timezone)
# - ca-certificates (Required for fetching data from https)
# - upx (Required for compress binary)
# - binutils (Required for strip unnecessary code)
RUN apk update && apk add --no-cache git tzdata ca-certificates upx binutils

# Update CA-Certificate
RUN update-ca-certificates

# Create appuser.
RUN adduser -D -g '' app
WORKDIR /app/

# Copy depedencies
COPY go.mod go.sum ./

# Get dependancies - will also be cached if we won't change mod/sum
RUN go mod download

# COPY the source code as the last step
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/app

# Strip Unnecessary binary code
RUN strip --strip-unneeded /go/bin/app
# Compress binary
RUN upx /go/bin/app

# Check if the binary not error after compressed with upx
RUN upx -t /go/bin/app


############################
# STEP 2 build a small image
############################
FROM scratch

# Import from builder.
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
# Import the user and group files from the builder.
COPY --from=builder /etc/passwd /etc/passwd
# Import the ca-certificate from the builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
# Copy our static executable.
COPY --from=builder /go/bin/app /app/app


# Use an unprivileged user.
USER app
# Run the binary.
ENTRYPOINT ["/app/app"]