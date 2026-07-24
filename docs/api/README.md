# Mini Bank API Reference

## Base URL

```
http://localhost:8080
```

## System Endpoints

### Health Check

```
GET /healthz
```

**Response:**
```json
{
  "status": "ok",
  "service": "api-gateway"
}
```

### Readiness Check

```
GET /readyz
```

Same response as health check.

### Prometheus Metrics

```
GET /metrics
```

Returns Prometheus text format metrics.

---

## Account Endpoints

### Create Account

```
POST /accounts
```

**Request:**
```json
{
  "owner_name": "John Doe"
}
```

**Response (201):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "owner_name": "John Doe",
  "balance": 0,
  "status": "ACTIVE",
  "created_at": "2026-07-25T10:00:00Z",
  "updated_at": "2026-07-25T10:00:00Z"
}
```

**Errors:**
- `400` — Validation error (empty/missing owner_name)

---

### Get Account

```
GET /accounts/:id
```

**Response (200):** Same as create response.

**Errors:**
- `404` — Account not found

---

### Block Account

```
POST /accounts/:id/block
```

**Response (200):**
```json
{
  "message": "account blocked"
}
```

**Errors:**
- `404` — Account not found
- `409` — Account already blocked

---

### Unblock Account

```
POST /accounts/:id/unblock
```

**Response (200):**
```json
{
  "message": "account unblocked"
}
```

**Errors:**
- `404` — Account not found
- `409` — Account not blocked

---

### Close Account

```
POST /accounts/:id/close
```

Account must have zero balance to close.

**Response (200):**
```json
{
  "message": "account closed"
}
```

**Errors:**
- `404` — Account not found
- `409` — Balance not zero or already closed

---

## Transfer Endpoints

### Create Transfer

```
POST /transfers
```

**Request:**
```json
{
  "from_account_id": "550e8400-e29b-41d4-a716-446655440000",
  "to_account_id": "660e8400-e29b-41d4-a716-446655440001",
  "amount": 5000,
  "idempotency_key": "770e8400-e29b-41d4-a716-446655440002"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `from_account_id` | UUID | Yes | Sender account ID |
| `to_account_id` | UUID | Yes | Receiver account ID |
| `amount` | int64 | Yes | Amount in smallest currency unit (e.g., cents). Must be > 0 |
| `idempotency_key` | UUID | Yes | Unique key for idempotent requests |

**Response (201):**
```json
{
  "id": "880e8400-e29b-41d4-a716-446655440003",
  "from_account_id": "550e8400-e29b-41d4-a716-446655440000",
  "to_account_id": "660e8400-e29b-41d4-a716-446655440001",
  "amount": 5000,
  "status": "COMPLETED",
  "idempotency_key": "770e8400-e29b-41d4-a716-446655440002",
  "created_at": "2026-07-25T10:05:00Z",
  "updated_at": "2026-07-25T10:05:00Z"
}
```

**Saga Flow:**
1. Reserve funds on sender account
2. Credit funds to receiver account
3. Commit reservation on sender account

If any step fails, compensating transactions are executed automatically.

**Errors:**
- `400` — Validation error (missing fields, amount <= 0, same account)
- `409` — Insufficient funds, idempotency key already used

---

### Get Transfer

```
GET /transfers/:id
```

**Response (200):** Same as create response.

**Errors:**
- `404` — Transaction not found

---

## Error Format

All errors return JSON:

```json
{
  "error": "description of the error",
  "code": "ERROR_CODE"
}
```

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `NOT_FOUND` | 404 | Resource not found |
| `VALIDATION_ERROR` | 400 | Invalid request parameters |
| `CONFLICT` | 409 | Business logic conflict |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

## Architecture

```
┌─────────────┐     ┌──────────────┐     ┌─────────────────┐
│   Client    │────▶│  API Gateway │────▶│ Account Service  │
│  (HTTP)     │     │   :8080      │     │   (gRPC :9001)  │
└─────────────┘     │              │     └─────────────────┘
                    │              │     ┌─────────────────┐
                    │              │────▶│Transaction Svc   │
                    └──────────────┘     │  (gRPC :9002)   │
                                         └────────┬────────┘
                                                  │
                                         ┌────────▼────────┐
                                         │ Account Service  │
                                         │  (saga target)   │
                                         └─────────────────┘
```

**Inter-service communication:** gRPC with circuit breaker + retry + timeout.

**Data flow:**
1. Client → API Gateway (HTTP REST)
2. API Gateway → Account/Transaction (gRPC)
3. Transaction → Account (gRPC, Saga steps)
4. Transaction → RabbitMQ (event publishing)
5. RabbitMQ → Notification Service (event consumption)
