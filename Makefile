.PHONY: up down

up:
	cd webapp && docker compose -f docker-compose.local.yml up -d

down:
	cd webapp && docker compose -f docker-compose.local.yml down

pprotein.logs:
	sudo journalctl -u pprotein-agent

daemon.reload:
	sudo systemctl daemon-reload
