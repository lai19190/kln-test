# KLN Code Test

This project implements two API scenarios using standard Go library features:

1. Pub/Sub subscription for shipping event updates
2. Multiple country public holiday data aggregation

## Project Structure

```
.
├── cmd/
│   └── api/           # Application entrypoint
├── internal/
│   ├── config/        # Configuration management
│   ├── handlers/      # HTTP request handlers
│   ├── holidays/      # Public holidays service
│   ├── middleware/    # HTTP middleware components
│   └── worker/        # Worker pool implementation
└── config.json        # Application configuration
```

## API Usage

### Subscribe to Shipping Events

```bash
curl -X POST http://localhost:8080/subscriptions \
  -H "Content-Type: application/json" \
  -H "Authorization: Basic YWRtaW46YWRtaW4=" \
  -d '{
    "consumerId": "client-123",
    "topics": ["shipping.created", "shipping.updated"],
    "deliveryUrl": "http://example.com/webhook"
  }'
```

### Get Public Holidays

```bash
curl "http://localhost:8080/public-holidays?year=2025&country=CA&country=DE" \
  -H "Authorization: Basic YWRtaW46YWRtaW4="
```

## Configuration

The application is configured via `config.json`:

```json
{
  "worker": {
    "poolSize": 5,
    "queueSize": 10,
    "retry": {
      "maxAttempts": 10,
      "initialTimeout": 1,
      "maxTimeout": 30
    }
  },
  "auth": {
    "username": "admin",
    "password": "admin"
  }
}
```

## Authentication

The API uses Basic Authentication. You need to include an `Authorization` header with your requests using the credentials configured in `config.json`.

## Running the Application

```bash
go run cmd/api/main.go
```

## Testing

Run all tests:

```bash
go test ./...
```
