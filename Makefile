.PHONY: up down

run:
	sudo truncate -s 0 /var/log/mysql/mysql-slow.log
	sudo truncate -s 0 /var/log/nginx/access.log
	sudo chmod +rx /var/log/mysql/
	sudo chmod +rx /var/log/nginx/
	bash run.sh

up:
	cd webapp && docker compose -f docker-compose.local.yml up -d

down:
	cd webapp && docker compose -f docker-compose.local.yml down

pp.start:
	sudo systemctl start pprotein

pp.status:
	sudo systemctl status pprotein

pp.restart:
	sudo systemctl restart pprotein

ppa:
	sudo systemctl start pprotein-agent

ppa.status:
	sudo systemctl status pprotein-agent

ppa.restart:
	sudo systemctl restart pprotein-agent

daemon.reload:
	sudo systemctl daemon-reload
