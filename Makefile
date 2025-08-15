# Пути к docker-compose файлов
COMPOSE_ORDER=./front-end/docker-compose.yaml
COMPOSE_MIGRATION=./order-back-end/docker-compose.yaml

# Запуск всех сервисов
up:
	docker-compose -f $(COMPOSE_ORDER) up -d --build
	docker-compose -f $(COMPOSE_MIGRATION) up -d --build

# Остановка всех сервисов
down:
	docker-compose -f $(COMPOSE_ORDER) down
	docker-compose -f $(COMPOSE_MIGRATION) down

# Перезапуск
restart: down up

# Просмотр логов всех сервисов
logs:
	docker-compose -f $(COMPOSE_ORDER) logs -f &
	docker-compose -f $(COMPOSE_MIGRATION) logs -f &
