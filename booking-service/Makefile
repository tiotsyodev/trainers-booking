include .env
export

create-migration:
	migrate create -ext sql -dir .\migrations -seq $(name)
make-migrate:
	migrate -path ./migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" up
start-service:
	docker compose -f docker-compose.yml --env-file .env up -d