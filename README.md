# Go Support ID Error - Example Implementation

Demo aplikasi Gin server yang menggunakan library [`go-support-id-error`](https://github.com/isaui/go-support-id-error) untuk error tracking dengan integrasi ELK cluster dan Discord webhook notifications.

## Features

### Core Features
- **Unique Error IDs** - Setiap error dapat unique tracking ID (format: `ERR-20251023-A3F9B2`)
- **ELK Integration** - Custom logger yang mengirim error logs ke ELK cluster (Elasticsearch/Logstash/Kibana)
- **Discord Notifications** - Callback yang mengirim error alerts ke Discord channel via webhook
- **Error Bot** - Goroutine yang secara berkala hit error endpoints untuk testing
- **Multiple Error Types** - Simulasi berbagai jenis error (database, validation, network, auth, payment, panic)

### Integration Details
- **Custom ELK Logger**: Implements `errorid.Logger` interface, mengirim logs ke ELK cluster
- **Discord Webhook**: Mengirim rich embeds dengan error details, metadata, dan stack traces
- **Automated Testing**: Bot goroutine yang hit random error endpoints setiap interval tertentu
- **Graceful Shutdown**: Proper cleanup saat server shutdown

## Installation

### Prerequisites
- Go 1.24.4 or higher
- Discord webhook URL (optional, untuk notifications)
- ELK cluster (optional, untuk centralized logging)

### Setup

1. **Clone atau download project**
```bash
cd go-support-id-example
```

2. **Install dependencies**
```bash
go mod download
```

3. **Configure environment variables**
```bash
cp .env.example .env
```

Edit `.env` file:
```env
PORT=8080
ENVIRONMENT=development
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/YOUR_WEBHOOK_ID/YOUR_WEBHOOK_TOKEN
ELK_URL=http://localhost:9200/logs/_doc
BOT_INTERVAL=30s
```

4. **Run the server**
```bash
go run .
```

Server akan berjalan di `http://localhost:8080`

## Usage

### API Endpoints

#### Health Check
```bash
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "time": "2025-10-23T12:27:00Z"
}
```

#### Error Endpoints

Semua error endpoints akan:
1. Generate unique error ID
2. Log ke stdout/stderr dan ELK cluster
3. Trigger Discord webhook notification
4. Return error response dengan error ID

**1. Database Error**
```bash
GET /api/error/database
```

Simulasi database connection timeout.

**2. Validation Error**
```bash
GET /api/error/validation
```

Simulasi input validation failure.

**3. Network Error**
```bash
GET /api/error/network
```

Simulasi external API call failure.

**4. Authentication Error**
```bash
GET /api/error/auth
```

Simulasi authentication failure dengan user metadata.

**5. Payment Error**
```bash
GET /api/error/payment
```

Simulasi payment processing error.

**6. Panic Error**
```bash
GET /api/error/panic
```

Simulasi panic (index out of range) yang akan di-catch oleh Gin recovery middleware.

**7. Uncaught Panic Error**
```bash
GET /api/error/uncaught-panic
```

Simulasi panic yang di-catch oleh `errorid.RecoveryMiddleware` dari library. Middleware ini akan:
- Catch panic dan wrap sebagai error dengan error ID
- Trigger OnError callback (Discord notification)
- Return proper JSON response dengan error ID
- **Tidak crash server** karena panic sudah di-handle

### Example Response

```json
{
  "error": "failed to connect to PostgreSQL: connection to database timed out after 30s",
  "error_id": "ERR-20251023-A3F9B2",
  "type": "database_error"
}
```

## Error Bot

Server otomatis menjalankan background goroutine yang secara berkala hit random error endpoints.

**Configuration:**
- Interval bisa di-set via `BOT_INTERVAL` environment variable
- Default: 30 detik
- Bot akan hit random endpoint dari semua available error endpoints

**Log Output:**
```
Error Bot started - hitting endpoints every 30s
Bot hitting: /api/error/database
Bot triggered error (Status: 500)
   Response: {"error":"...","error_id":"ERR-20251023-A3F9B2"}
```

## ELK Integration

### Custom Logger Implementation

File: `elk_logger.go`

Logger implements `errorid.Logger` interface dengan 2 methods:
- `Error(errorID, err, context, details)` - For error logging
- `Info(msg)` - For info logging

Logs dikirim ke ELK dalam **structured format** untuk better Kibana filtering:

```json
{
  "@timestamp": "2025-10-23T12:27:00Z",
  "error_id": "ERR-20251023-A3F9B2",
  "error_type": "tracked",
  "context": "database query",
  "error": "connection timeout",
  "service": "go-support-id-example",
  "level": "error",
  "environment": "production",
  "database": "postgres",
  "host": "db.example.com",
  "port": 5432
}
```

**Benefits:**
- Semua details jadi separate fields di Kibana
- Easy filtering: `error_id: "ERR-*"`, `database: "postgres"`, `port: 5432`
- No need regex parsing!

### ELK Setup Options

**Option 1: Elasticsearch Direct**
```env
ELK_URL=http://localhost:9200/logs/_doc
ELK_USERNAME=elastic
ELK_PASSWORD=your_password
```

**Option 2: Logstash HTTP Input**
```env
ELK_URL=http://localhost:5000
```

Logstash config example:
```ruby
input {
  http {
    port => 5000
    codec => json
  }
}

output {
  elasticsearch {
    hosts => ["localhost:9200"]
    index => "go-support-id-errors-%{+YYYY.MM.dd}"
  }
}
```

## Discord Integration

### Discord Webhook Setup

1. Di Discord server, pergi ke **Server Settings > Integrations > Webhooks**
2. Click **New Webhook**
3. Set nama dan channel untuk notifications
4. Copy webhook URL
5. Paste ke `.env` file

### Discord Notification Format

Bot mengirim rich embed dengan:
- **Title**: Error ID (max 256 chars)
- **Description**: Error message dan context (max 2048 chars)
- **Fields**: 
  - Details (custom data dari WrapWithDetails, max 1024 chars per field)
  - Environment (production/development)
  - Stack trace (jika enabled, truncated to 900 chars)
- **Color**: Red (15158332)

**Field Limits:**
Semua fields di-validate dan truncate sesuai Discord API limits untuk prevent 400 errors.

Example Discord notification:

```
Error Alert: ERR-20251023-A3F9B2

Error: connection to database timed out after 30s
Context: failed to connect to PostgreSQL

Metadata:
• database: postgres
• host: db.example.com
• port: 5432
• timeout: 30s

Environment: production
```

## Architecture

### File Structure

```
go-support-id-example/
├── main.go              # Main entry point, server initialization
├── routes.go            # Route definitions and grouping
├── handlers.go          # HTTP handlers with error handling
├── services.go          # Business logic services (return errors)
├── adapter.go           # GinRecoveryMiddleware adapter for library
├── elk_logger.go        # Custom ELK logger implementation
├── discord.go           # Discord webhook integration
├── bot.go               # Error bot goroutine
├── go.mod               # Go dependencies
├── .env.example         # Environment variables template
└── README.md            # This file
```

### Components

**1. Main Server (`main.go`)**
- Application initialization
- Error-ID library configuration
- Bot initialization
- Graceful shutdown
- Startup info display

**2. Routes (`routes.go`)**
- Route definitions and grouping
- Middleware application per route group
- API endpoint organization

**3. Handlers (`handlers.go`)**
- HTTP request handlers
- Service layer integration
- Error handling with `if err != nil` pattern
- Error wrapping with error-id

**4. Services (`services.go`)**
- Business logic layer
- Database, user, payment, auth services
- Returns errors for handlers to handle
- Simulates real-world service operations

**5. Adapter (`adapter.go`)**
- `GinRecoveryMiddleware()` - Bridges `errorid.RecoveryMiddleware` to Gin
- Wraps library's RecoveryMiddleware for use as Gin middleware
- Catches panics and returns proper JSON with error ID

**6. ELK Logger (`elk_logger.go`)**
- Implements `errorid.Logger` interface:
  - `Error(errorID, err, context, details)` - Structured error logging
  - `Info(msg)` - Info logging
- Sends logs to ELK cluster via HTTP POST
- **Fully structured JSON** - all details as separate fields
- Async sending to prevent blocking
- Supports both Elasticsearch direct and Logstash HTTP input

**7. Discord Webhook (`discord.go`)**
- OnError callback handler
- Rich embed formatting with Discord API limits
- Field validation and truncation (prevent 400 errors)
- Details and stack trace inclusion
- Async notification sending

**8. Error Bot (`bot.go`)**
- Background goroutine
- Periodic endpoint testing
- Random endpoint selection
- HTTP client for API calls

### Error Flow (New Architecture)

```
1. HTTP Request → Handler
   ↓
2. Handler calls Service layer
   ↓
3. Service returns error (if err != nil)
   ↓
4. Handler wraps error with errorid.WrapWithDetails()
   ↓
5. Unique ID generated (ERR-20251023-A3F9B2)
   ↓
6. Middleware logs request/response
   ↓
7. ELK Logger: Send to ELK cluster (async)
   ↓
8. Discord Callback: Send notification (async)
   ↓
9. Return error response to client
```

### Real-World Error Handling Pattern

```go
// Service layer - returns error
func (s *DatabaseService) Connect() error {
    return errors.New("connection timeout")
}

// Handler - handles error with if err != nil
func (h *Handlers) HandleDatabaseError(c *gin.Context) {
    err := h.dbService.Connect()
    
    if err != nil {  // Real error handling pattern!
        wrappedErr := errorid.WrapWithDetails(err, "context", metadata)
        c.JSON(500, gin.H{"error_id": wrappedErr.ID})
        return
    }
    
    c.JSON(200, gin.H{"message": "success"})
}
```

## Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `PORT` | Server port | `8080` | No |
| `ENVIRONMENT` | Environment name | `development` | No |
| `DISCORD_WEBHOOK_URL` | Discord webhook URL | - | No |
| `ELK_URL` | ELK cluster endpoint | - | No |
| `ELK_USERNAME` | ELK authentication username | - | No |
| `ELK_PASSWORD` | ELK authentication password | - | No |
| `BOT_INTERVAL` | Bot hit interval (e.g., `30s`, `1m`) | `30s` | No |

### Error-ID Configuration

In `main.go`:

```go
errorid.Configure(errorid.Config{
    OnError:            discordCallback,  // Discord webhook
    AsyncCallback:      true,             // Non-blocking
    Logger:             elkLogger,        // Custom ELK logger
    IncludeStackTrace:  true,            // Capture stack traces
    Environment:        "production",     // Environment
})
```

## Examples

### Testing with curl

```bash
# Test database error
curl http://localhost:8080/api/error/database

# Test validation error
curl http://localhost:8080/api/error/validation

# Test payment error
curl http://localhost:8080/api/error/payment
```

### Custom Error Handler Example

```go
func handleCustomError(c *gin.Context) {
    // Your business logic
    err := doSomething()
    
    if err != nil {
        // Wrap with error-id
        wrappedErr := errorid.WrapWithDetails(
            err,
            "custom operation failed",
            map[string]interface{}{
                "user_id": 12345,
                "action": "process_data",
            },
        )
        
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": wrappedErr.Error(),
            "error_id": wrappedErr.ID,
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"status": "success"})
}
```

### Recovery Middleware Example

**Global Recovery Middleware:**
```go
// In setupServer() - main.go
router := gin.New()
router.Use(gin.Logger())
router.Use(GinRecoveryMiddleware()) // Library's RecoveryMiddleware!

// Now all panics are caught and wrapped with error ID
// Example panic response:
{
  "error_id": "ERR-20251023-A3F9B2",
  "message": "[ERR-20251023-A3F9B2] panic recovered in HTTP handler: index out of range",
  "timestamp": 1761200072
}
```

**How it works:**
1. Panic occurs in handler
2. `errorid.RecoveryMiddleware` catches it
3. Wraps panic as error with error ID
4. Triggers OnError callback (Discord notification)
5. Returns JSON response (tidak crash server)

**Adapter Implementation:**
```go
// adapter.go
func GinRecoveryMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        handler := errorid.RecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            c.Next()
        }))
        handler.ServeHTTP(c.Writer, c.Request)
    }
}
```

## Production Considerations

1. **ELK Authentication**: Always use credentials untuk production ELK cluster
2. **Discord Rate Limits**: Discord webhooks limited to 30 requests/minute per webhook
3. **Async Callbacks**: Enable `AsyncCallback: true` untuk prevent blocking
4. **Bot Interval**: Adjust `BOT_INTERVAL` based on your testing needs (longer interval for production)
5. **Environment**: Set `ENVIRONMENT=production` untuk production deployments

## Resources

- **go-support-id-error Library**: https://github.com/isaui/go-support-id-error
- **Gin Framework**: https://gin-gonic.com/
- **Discord Webhooks**: https://discord.com/developers/docs/resources/webhook
- **ELK Stack**: https://www.elastic.co/what-is/elk-stack

## License

MIT License

## Contributing

Feel free to open issues atau pull requests untuk improvements!
