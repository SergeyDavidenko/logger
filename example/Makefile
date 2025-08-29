.PHONY: up down logs test clean

# Start ELK stack
up:
	mkdir -p logs
	docker compose up -d
	@echo "Waiting for services to start..."
	@sleep 30
	@echo "✅ ELK stack is running! ✅"
	@echo "✅ Kibana: http://localhost:5601 ✅"
	@echo "✅ Elasticsearch: http://localhost:9200 ✅"
	@echo "🐳 ELK stack is running! 🐳"

# Stop and remove containers
down:
	docker compose down -v
	@echo "🐳 ELK stack is stopped! 🐳"

# Show logs of all services
logs:
	docker compose logs -f

# Show only logstash logs
logs-logstash:
	docker compose logs -f logstash

# Run test application
test:
	go run main.go

# Check services status
status:
	@echo "=== Docker containers ==="
	@docker compose ps
	@echo "\n=== Elasticsearch health ==="
	@curl -s http://localhost:9200/_cluster/health?pretty || echo "❌ Elasticsearch is not available ❌"

# Clean logs
clean:
	rm -rf logs/*
	@echo "🧹 Logs cleaned! 🧹"
	docker compose down -v
	@echo "🐳 ELK stack is stopped! 🐳"

# Full restart
restart: down up

# Show recent logs from files
show-logs:
	@echo "📝=== Recent logs from files ===📝"
	@find logs -name "*.log" -type f -exec echo "=== {} ===" \; -exec tail -5 {} \;

help:
	@echo "Available commands:"
	@echo "  up           - Start ELK stack"
	@echo "  down         - Stop and remove containers"
	@echo "  logs         - Show logs of all services"
	@echo "  logs-logstash- Show only logstash logs"
	@echo "  test         - Run test application"
	@echo "  status       - Check services status"
	@echo "  clean        - Clean logs"
	@echo "  restart      - Full restart"
	@echo "  show-logs    - Show recent logs from files"
