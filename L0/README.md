# Wildberries Test L0 - Микросервис для обработки заказов

## Описание проекта

Демонстрационный микросервис на Go для обработки и отображения данных о заказах. Сервис получает данные заказов из Kafka, сохраняет их в PostgreSQL и кэширует в памяти для быстрого доступа.

## Архитектура

Проект состоит из следующих компонентов:

- **Backend (order-back-end)** - Go микросервис
  - HTTP API для получения данных заказов
  - Kafka Consumer для получения сообщений
  - PostgreSQL для хранения данных
  - In-memory кэш для быстрого доступа
  - Автоматическое восстановление кэша при перезапуске

- **Frontend (front-end)** - Веб-интерфейс
  - HTML/JS страница для поиска заказов по ID
  - Подключение к Backend API

- **Infrastructure**
  - PostgreSQL (порт 5432)
  - Kafka кластер (3 брокера на портах 9092)
  - Docker контейнеры для всех компонентов

## Технологии

- **Backend**: Go, Gin, GORM, Sarama (Kafka), PostgreSQL
- **Frontend**: HTML, JavaScript
- **Infrastructure**: Docker, Docker Compose
- **Message Broker**: Apache Kafka
- **Database**: PostgreSQL

## Быстрый старт

### Предварительные требования

- Docker
- Docker Compose
- Make

### Запуск проекта

1. **Клонируйте репозиторий:**
```bash
git clone <repository-url>
cd wildberries-test-l0
```

2. **Запустите все сервисы:**
```bash
make up
```

Эта команда:
- Собирает и запускает контейнеры для frontend и order-service
- Настраивает PostgreSQL базу данных
- Запускает Kafka кластер
- Ждет полной готовности всех сервисов (wait-for-it логика)

3. **Остановите сервисы:**
```bash
make down
```

## Доступ к сервисам

- **Frontend**: http://localhost:8080
- **Backend API**: http://localhost:8081
- **PostgreSQL**: localhost:5432
- **Kafka**: localhost:9095 (kafka-1), localhost:9096 (kafka-2), localhost:9097 (kafka-3)

## API Endpoints

### Получение заказа по ID

```
GET /order/{order_uid}
```

**Пример запроса:**
```bash
curl http://localhost:8081/order/:id
```

**Ответ:**
```json
{
  "order_uid": "123456",
  "track_number": "WBILMTESTTRACK",
  "entry": "WBIL",
  "delivery": {
    "name": "Test Testov",
    "phone": "+9720000000",
    "zip": "2639809",
    "city": "Kiryat Mozkin",
    "address": "Ploshad Mira 15",
    "region": "Kraiot",
    "email": "test@gmail.com"
  },
  "payment": {
    "transaction": "b563feb7b2b84b6test",
    "request_id": "",
    "currency": "USD",
    "provider": "wbpay",
    "amount": 1817,
    "payment_dt": 1637907727,
    "bank": "alpha",
    "delivery_cost": 1500,
    "goods_total": 317,
    "custom_fee": 0
  },
  "items": [
    {
      "chrt_id": 9934930,
      "track_number": "WBILMTESTTRACK",
      "price": 453,
      "rid": "ab4219087a764ae0btest",
      "name": "Mascaras",
      "sale": 30,
      "size": "0",
      "total_price": 317,
      "nm_id": 2389212,
      "brand": "Vivienne Sabo",
      "status": 202
    }
  ],
  "locale": "en",
  "internal_signature": "",
  "customer_id": "test",
  "delivery_service": "meest",
  "shardkey": "9",
  "sm_id": 99,
  "date_created": "2021-11-26T06:22:19Z",
  "oof_shard": "1"
}
```

## Использование веб-интерфейса

1. Откройте http://localhost:8080 в браузере
2. Введите ID заказа в поле поиска
3. Нажмите кнопку поиска
4. Результат отобразится на странице

### Как найти Order ID

Order ID можно получить из логов Docker контейнера:

```bash
docker logs <container-name>
```

Или посмотреть логи order-service контейнера для получения ID обработанных заказов.

## Особенности реализации

### Кэширование
- Данные заказов кэшируются в памяти для быстрого доступа
- При перезапуске сервиса кэш автоматически восстанавливается из базы данных
- Повторные запросы по одному ID выполняются мгновенно

### Обработка ошибок
- Валидация входящих сообщений из Kafka
- Логирование некорректных сообщений
- Транзакционная обработка данных в PostgreSQL
- Подтверждение сообщений от Kafka брокера

### Надежность
- Использование транзакций для сохранения данных
- Механизм подтверждения сообщений от Kafka
- Автоматическое восстановление кэша при сбоях
- Wait-for-it логика для корректного запуска всех сервисов

## Структура проекта

```
wildberries-test-l0/
├── front-end/                 # Веб-интерфейс
│   ├── docker-compose.yaml
│   ├── Dockerfile
│   └── index.html
├── order-back-end/           # Go микросервис
│   ├── cmd/main.go          # Точка входа
│   ├── internal/
│   │   ├── cache/           # Кэширование
│   │   ├── config/          # Конфигурация
│   │   ├── handler/         # HTTP обработчики
│   │   ├── kafka/           # Kafka producer/consumer
│   │   ├── model/           # Модели данных
│   │   ├── postgres/        # Работа с БД
│   │   ├── repository/      # Репозиторий
│   │   └── service/         # Бизнес-логика
│   ├── migrations/          # Миграции БД
│   └── docker-compose.yaml
├── Makefile                 # Команды для сборки и запуска
└── README.md
```

## Разработка

### Локальная разработка

1. Убедитесь, что у вас установлен Go 1.24+
2. Настройте переменные окружения в `order-back-end/config/config.yaml`
3. Запустите PostgreSQL и Kafka локально или через Docker
4. Запустите сервис: `go run cmd/main.go`

### Тестирование

Для тестирования API можно использовать curl или Postman:

```bash
# Получить заказ по ID
curl http://localhost:8081/order/test-order-id
```

## Мониторинг

- Логи сервиса: `docker logs order-service`
- Логи frontend: `docker logs frontend`
- Статус контейнеров: `docker ps`

## Лицензия

Этот проект создан в рамках тестового задания.
