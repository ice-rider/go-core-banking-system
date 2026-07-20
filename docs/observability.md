# Наблюдаемость (Observability)

## Трассировка
- OpenTelemetry SDK во всех сервисах
- Сквозная трассировка: Gin (HTTP) → gRPC → PostgreSQL
- Визуализация в Jaeger

## Метрики (Prometheus)
- RPS по эндпоинтам
- Latency (p50, p95, p99)
- Ошибки (БД, gRPC-таймауты, откаты Saga)
- Активные соединения с БД и RabbitMQ

## Логирование
- Структурированные JSON-логи
- Обязательные поля: trace_id, span_id

## Визуализация
- Grafana: дашборды с графиками RPS, Latency, ошибок
