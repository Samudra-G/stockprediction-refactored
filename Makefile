up:
	docker-compose up -d --build

down:
	docker-compose down

logs:
	docker-compose logs -f

restart:
	docker-compose down && docker-compose up -d --build

#For specific rebuilds
build-backend:
	docker-compose build backend-go

build-ml:
	docker-compose build ml-fastapi

build-frontend:
	docker-compose build frontend

#For local go-backend
local-server:
	cd backend-go && go run cmd/main.go

test-backend:
	cd backend-go && go test -v -cover ./...

test-ml:
	docker-compose exec ml-fastapi pytest -v --maxfail=1

.PHONY: up down logs build-backend build-ml build-frontend local-server test-backend test-ml