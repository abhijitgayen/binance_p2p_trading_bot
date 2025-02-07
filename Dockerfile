# Use a minimal base image
FROM alpine:latest

# Install required system dependencies
RUN apk add --no-cache ca-certificates sqlite-libs

# Copy the pre-built Go binary into the container
COPY go_binance_bot /go_binance_bot

# Copy the .env file if needed (optional)
COPY .env /app/.env

# Set execution permissions
RUN chmod +x /go_binance_bot

# Set the working directory
WORKDIR /app

# Run the binary
CMD ["/go_binance_bot"]
