# Пути к docker-compose файлов
COMPOSE_ORDER=./front-end/docker-compose.yaml
COMPOSE_MIGRATION=./order-back-end/docker-compose.yaml
COMPOSE_NAME=order-back-end_
POSTGRES_VOLUME := ${COMPOSE_NAME}postgres_volume
# Запуск всех сервисов
up:
	docker-compose -f $(COMPOSE_ORDER) up -d --build
	docker-compose -f $(COMPOSE_MIGRATION) up -d --build

# Остановка всех сервисов
down:
	docker-compose -f $(COMPOSE_ORDER) down
	docker-compose -f $(COMPOSE_MIGRATION) down
	for vol in $$(docker volume ls -q --filter name=$(COMPOSE_NAME)); do \
		if [ "$$vol" != "$(POSTGRES_VOLUME)" ]; then \
			docker volume rm $$vol || true; \
		fi; \
	done




# Перезапуск
restart: down up

# Просмотр логов всех сервисов
logs:
	docker-compose -f $(COMPOSE_ORDER) logs -f &
	docker-compose -f $(COMPOSE_MIGRATION) logs -f &
