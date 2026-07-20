# Технологический стек

| Компонент | Технология |
|---|---|
| Язык | Go 1.26 |
| REST API | Gin |
| Межсервисное взаимодействие | gRPC (protobuf3) |
| База данных | PostgreSQL 15+ (database-per-service) |
| Кэширование / блокировки | Redis |
| Брокер сообщений | RabbitMQ |
| Трассировка | OpenTelemetry, Jaeger |
| Метрики | Prometheus, Grafana |
| Инфраструктура | Docker, Docker Compose |
| CI/CD | GitHub Actions |

## Зависимости инфраструктуры
- 2 инстанса PostgreSQL (отдельные БД для Account и Transaction сервисов)
- Redis (кэширование справочников, распределённые блокировки)
- RabbitMQ (асинхронные события, Saga)
- Jaeger, Prometheus, Grafana (мониторинг)
