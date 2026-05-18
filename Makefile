.PHONY: mailpit-up mailpit-down mailpit-logs mailpit-clear

mailpit-up:
	docker compose -f docker-compose.mailpit.yml up -d
	@echo "Mailpit started - SMTP: localhost:1025, UI: http://localhost:8025"

mailpit-down:
	docker compose -f docker-compose.mailpit.yml down

mailpit-logs:
	docker compose -f docker-compose.mailpit.yml logs -f

mailpit-clear:
	@curl -s -X DELETE http://localhost:8025/api/v1/deleteall || true
	@echo "Mailpit messages cleared"