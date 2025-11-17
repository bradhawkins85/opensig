.PHONY: all api web relay up down

all: api web

api:
	cd server && docker build -t opensig-api:dev --target api .

relay:
	cd server && docker build -t opensig-relay:dev --target relay .

web:
	cd web && docker build -t opensig-web:dev .

up:
	docker compose -f deploy/docker-compose.yml up --build

down:
	docker compose -f deploy/docker-compose.yml down -v
