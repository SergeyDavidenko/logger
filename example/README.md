# Logger Usage Example with Logstash

This example demonstrates the logger working with a full ELK stack (Elasticsearch, Logstash, Kibana).

## Getting Started

1. **Start Docker Compose:**
   ```bash
   docker compose up -d
   ```

2. **Wait for all services to start (may take 1-2 minutes):**
   ```bash
   docker compose logs -f
   ```

3. **Run the Go application example:**
   ```bash
   go run main.go
   ```

## Verification

### Kibana (Web Interface)
- Open http://localhost:5601
- Go to Management → Stack Management → Index Management
- Create an index pattern for `logstash-*`
- Go to Analytics → Discover to view logs

### Log Files
Logs are also saved to files in the `./logs/` directory:
```bash
ls -la logs/dev/example-app/
```

### Logstash stdout
Logs are also output to Logstash console:
```bash
docker compose logs logstash
```

## Services

- **Elasticsearch**: http://localhost:9200
- **Kibana**: http://localhost:5601
- **Logstash TCP**: localhost:5008

## Stopping

```bash
docker compose down -v
```

## Log Structure

After processing in Logstash, logs will contain:
- `Time` - timestamp
- `Severity` - logging level (renamed from `level`)
- `Thread` - thread information (renamed from `thread_name`)
- `logger_info` - logger information
- `message` - log message
- `appid` - application identifier

## Example Output

```
2025-01-27T10:30:15.123Z INFO [main.go:25] [go-logger] Application started successfully
```
