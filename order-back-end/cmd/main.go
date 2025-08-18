package main

import (
	"context"
	"fmt"
	"net/http"
	"order-back-end/internal/cache"
	"order-back-end/internal/config"
	hand "order-back-end/internal/handler"
	consumer "order-back-end/internal/kafka/consumer"
	producer "order-back-end/internal/kafka/producer"
	"order-back-end/internal/logger"
	"order-back-end/internal/postgres"
	repo "order-back-end/internal/repository"
	serv "order-back-end/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background() // создаём базовый контекст

	ctx, err := logger.New(ctx) // инициализируем логгер
	if err != nil {
		fmt.Println(err)
	}

	cfg, err := config.NewConfig() // загружаем конфигурацию
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "config.New error", zap.Error(err))
	}

	db, err := postgres.New(ctx, cfg.Postgres) // создаём подключение к базе
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "postgres.New error", zap.Error(err))
	}
	db.Ping(ctx)     // проверяем соединение с базой
	defer db.Close() // закрываем базу при выходе

	cacheIn := cache.NewCache() // создаём кэш для хранения заказов

	repo := repo.NewRepository(db) // создаём репозиторий для работы с базой

	orders, err := repo.GetAllOrders(ctx) // получаем все заказы из базы
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "repo.GetAllOrders error", zap.Error(err))
	}

	for _, orderElem := range orders { // наполняем кэш существующими заказами
		cacheIn.Orders[orderElem.OrderUID] = orderElem
	}

	cons := consumer.NewConsumer(db, cacheIn) // создаём консьюмера для kafka

	go producer.StartProducer(ctx, cfg.Kafka.Brokers, cfg.Kafka.Topic) // запускаем продьюсера в отдельной горутине

	go func() { // запускаем консьюмера в отдельной горутине
		err = cons.StartConsuming(ctx, cfg.Kafka.Brokers, cfg.Kafka.GroupID, cfg.Kafka.Topic)
		if err != nil {
			logger.GetLoggerFromCtx(ctx).Fatal(ctx, "consumer.StartConsuming error", zap.Error(err))
		}
	}()

	router := gin.Default()          // создаём новый gin router
	router.Use(cors.New(cors.Config{ // настраиваем cors для фронтенда
		AllowOrigins:     []string{"http://localhost:8080"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true,
	}))

	service := serv.NewOrderService(repo) // создаём сервис для работы с заказами

	handler := hand.NewHandler(service, router, cacheIn) // создаём обработчик http запросов

	handler.RegisterRoutes() // регистрируем маршруты

	go func() { // запускаем http сервер в отдельной горутине
		fmt.Println("start http server on :8081")
		if err := http.ListenAndServe(":"+cfg.HTTP.Port, router); err != nil {
			logger.GetLoggerFromCtx(ctx).Fatal(ctx, "http.ListenAndServe error", zap.Error(err))
		}
	}()

	select {} // блокируем основной поток, чтобы приложение не завершилось
}
