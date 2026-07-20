# Карта задач (Roadmap)

Задачи сгруппированы по спринтам (S1–S7) из tasks.txt.

## S1 — Инфраструктура
| # | Задача | Issue |
|---|---|---|
| 9 | Project structure setup | PVTI...Zfg |
| 10 | Docker Compose infrastructure | PVTI...Zfs |
| 11 | Database schemas + migrations | PVTI...ZgI |
| 12 | CI/CD pipeline | PVTI...ZgM |

## S2 — Account Service
| # | Задача | Issue |
|---|---|---|
| 13 | Define account.proto | PVTI...Zic |
| 14 | Implement Account repository | PVTI...Zig |
| 15 | Implement Account service | PVTI...Zik |
| 16 | Account Service unit tests | PVTI...Zi4 |

## S3 — Transaction Service
| # | Задача | Issue |
|---|---|---|
| 17 | Define transaction.proto | PVTI...VkA |
| 18 | Implement Transaction repository | PVTI...VkI |
| 19 | Implement pessimistic locking | PVTI...VkU |
| 20 | Basic transfer logic | PVTI...VkY |
| 21 | Transaction Service unit tests | PVTI...Vkc |

## S4 — Saga
| # | Задача | Issue |
|---|---|---|
| 22 | Design Saga state machine | PVTI...Vlg |
| 23 | Implement Saga orchestrator | PVTI...Vlo |
| 24 | Compensating transactions | PVTI...Vl4 |
| 25 | Idempotency key handling | PVTI...VmA |
| 26 | Integration tests | PVTI...VmE |

## S5 — API Gateway
| # | Задача | Issue |
|---|---|---|
| 27 | Gin router + middleware | PVTI...Vm4 |
| 28 | REST endpoints | PVTI...Vm8 |
| 29 | gRPC client wrappers | PVTI...VnE |
| 30 | Request validation | PVTI...VnI |
| 31 | Error handling | PVTI...VnM |

## S6 — Messaging
| # | Задача | Issue |
|---|---|---|
| 32 | RabbitMQ setup | PVTI...VoI |
| 33 | Transaction events | PVTI...VoQ |
| 34 | Async notification handler | PVTI...Vog |
| 35 | Dead letter queue | PVTI...Vok |

## S7 — Observability
| # | Задача | Issue |
|---|---|---|
| 36 | OpenTelemetry integration | PVTI...VpI |
| 37 | Jaeger trace propagation | PVTI...VpM |
| 38 | Prometheus metrics | PVTI...VpQ |
