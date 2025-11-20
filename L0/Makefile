COMPOSE_DIR_BACK=./order-back-end/build/docker
COMPOSE_DIR_FRONT=./front-end
up:
	cd $(COMPOSE_DIR_BACK) && docker-compose -f docker-compose.yaml up -d --build
	cd $(COMPOSE_DIR_FRONT) && docker-compose -f docker-compose.yaml up -d --build

down:
	cd $(COMPOSE_DIR_BACK) && docker-compose -f docker-compose.yaml down
	cd $(COMPOSE_DIR_FRONT) && docker-compose -f docker-compose.yaml down

restart: down up
