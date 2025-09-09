#!/bin/bash
# Ждём Postgres
./wait-for-it.sh postgres:5432 -t 40 -- echo "Postgres ready"

# Ждём Kafka
./wait-for-it.sh kafka-1:9092 -t 40 -- echo "Kafka ready"
./wait-for-it.sh kafka-2:9092 -t 40 -- echo "Kafka ready"
./wait-for-it.sh kafka-3:9092 -t 40 -- echo "Kafka ready"

# Запускаем сервис
./order-service