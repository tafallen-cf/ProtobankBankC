# Docker Architecture - ProtobankBankC

Complete Docker infrastructure architecture for local development environment.

## Overview

The development environment uses Docker Compose to orchestrate 11 containers across 3 layers:
1. **Data Layer** - PostgreSQL, PgBouncer, Redis, RabbitMQ
2. **Application Layer** - 8 Go microservices
3. **Optional Layer** - Nginx, Prometheus, Grafana (commented out)

---

## Container Network Architecture

```mermaid
graph TB
    subgraph "Host Machine (localhost)"
        DEV[Developer<br/>Browser/Postman]
    end

    subgraph "Docker Network: protobank-network<br/>172.20.0.0/16"
        subgraph "Data Layer"
            PG[(PostgreSQL<br/>:5432<br/>protobank-postgres<br/>Volume: postgres_data)]
            PGB[PgBouncer<br/>:6432<br/>protobank-pgbouncer<br/>Connection Pooler]
            RD[(Redis<br/>:6379<br/>protobank-redis<br/>Volume: redis_data)]
            RMQ[RabbitMQ<br/>:5672, :15672<br/>protobank-rabbitmq<br/>Volume: rabbitmq_data]
        end

        subgraph "Application Layer - Microservices"
            AS[Auth Service<br/>:3001<br/>JWT + Sessions]
            US[User Service<br/>:3002<br/>Profiles + KYC]
            ACS[Account Service<br/>:3003<br/>Accounts + Balances]
            TS[Transaction Service<br/>:3004<br/>Transactions]
            CS[Card Service<br/>:3005<br/>Card Management]
            PS[Payment Service<br/>:3006<br/>Payments]
            NS[Notification Service<br/>:3007<br/>Push + Email + SMS]
            ANS[Analytics Service<br/>:3008<br/>Insights + Reports]
        end
    end

    DEV -->|HTTP :3001-3008| AS
    DEV -->|HTTP :3001-3008| US
    DEV -->|HTTP :3001-3008| ACS
    DEV -->|HTTP :3001-3008| TS
    DEV -->|HTTP :3001-3008| CS
    DEV -->|HTTP :3001-3008| PS
    DEV -->|HTTP :3001-3008| NS
    DEV -->|HTTP :3001-3008| ANS
    DEV -->|PostgreSQL :5432| PG
    DEV -->|PostgreSQL :6432| PGB
    DEV -->|Redis :6379| RD
    DEV -->|AMQP :5672| RMQ
    DEV -->|HTTP :15672| RMQ

    PGB -->|Connection Pool| PG

    AS -->|SQL| PGB
    AS -->|Cache/Session| RD

    US -->|SQL| PGB

    ACS -->|SQL| PGB
    ACS -->|Cache| RD

    TS -->|SQL| PGB
    TS -->|Cache| RD
    TS -->|Events| RMQ

    CS -->|SQL| PGB

    PS -->|SQL| PGB
    PS -->|Events| RMQ

    NS -->|SQL| PGB
    NS -->|Events| RMQ

    ANS -->|SQL| PGB
    ANS -->|Cache| RD
    ANS -->|Events| RMQ

    style PG fill:#336791
    style RD fill:#DC382D
    style RMQ fill:#FF6600
    style AS fill:#00ADD8
    style US fill:#00ADD8
    style ACS fill:#00ADD8
    style TS fill:#00ADD8
    style CS fill:#00ADD8
    style PS fill:#00ADD8
    style NS fill:#00ADD8
    style ANS fill:#00ADD8
```

---

## Service Dependency Graph

```mermaid
graph LR
    subgraph "Infrastructure Services (Always Start First)"
        PG[PostgreSQL]
        PGB[PgBouncer]
        RD[Redis]
        RMQ[RabbitMQ]
    end

    subgraph "Core Services (Authentication Layer)"
        AS[Auth Service]
    end

    subgraph "Business Services (Application Layer)"
        US[User Service]
        ACS[Account Service]
        TS[Transaction Service]
        CS[Card Service]
        PS[Payment Service]
        NS[Notification Service]
        ANS[Analytics Service]
    end

    PG --> PGB
    PGB --> AS
    RD --> AS

    AS --> US
    PGB --> US

    AS --> ACS
    PGB --> ACS
    RD --> ACS

    AS --> TS
    ACS --> TS
    PGB --> TS
    RD --> TS
    RMQ --> TS

    AS --> CS
    ACS --> CS
    PGB --> CS

    AS --> PS
    ACS --> PS
    TS --> PS
    PGB --> PS
    RMQ --> PS

    AS --> NS
    PGB --> NS
    RMQ --> NS

    AS --> ANS
    PGB --> ANS
    RD --> ANS
    RMQ --> ANS

    style PG fill:#336791,stroke:#fff,color:#fff
    style PGB fill:#336791,stroke:#fff,color:#fff
    style RD fill:#DC382D,stroke:#fff,color:#fff
    style RMQ fill:#FF6600,stroke:#fff,color:#fff
    style AS fill:#00ADD8,stroke:#fff,color:#fff
```

**Startup Order**:
1. PostgreSQL (with health check)
2. PgBouncer (depends on PostgreSQL)
3. Redis (independent)
4. RabbitMQ (independent)
5. Auth Service (depends on PostgreSQL + Redis)
6. All other services (depend on Auth Service)

---

## Data Flow Architecture

### Authentication Flow

```mermaid
sequenceDiagram
    participant Client as Client<br/>(Browser/Mobile)
    participant Auth as Auth Service<br/>:3001
    participant Redis as Redis<br/>:6379
    participant PGB as PgBouncer<br/>:6432
    participant PG as PostgreSQL<br/>:5432

    Client->>Auth: POST /auth/login<br/>{email, password}

    Auth->>PGB: Query user by email
    PGB->>PG: SELECT * FROM users<br/>WHERE email = ?
    PG-->>PGB: User record
    PGB-->>Auth: User data

    Auth->>Auth: Verify password<br/>(bcrypt compare)

    Auth->>Auth: Generate JWT tokens<br/>(access + refresh)

    Auth->>Redis: Store refresh token<br/>KEY: session:{user_id}<br/>TTL: 7 days
    Redis-->>Auth: OK

    Auth-->>Client: 200 OK<br/>{access_token, refresh_token,<br/>expires_in: 900}

    Note over Client,Auth: Access token expires in 15 minutes
    Note over Redis: Refresh token stored for 7 days
```

### Transaction Creation Flow

```mermaid
sequenceDiagram
    participant Client as Client
    participant Auth as Auth Service
    participant Txn as Transaction Service<br/>:3004
    participant Acc as Account Service<br/>:3003
    participant PGB as PgBouncer
    participant PG as PostgreSQL
    participant Redis as Redis
    participant RMQ as RabbitMQ
    participant Notif as Notification Service<br/>:3007
    participant Analy as Analytics Service<br/>:3008

    Client->>Txn: POST /transactions<br/>Authorization: Bearer {token}

    Txn->>Auth: Validate JWT token
    Auth-->>Txn: Token valid, user_id

    Txn->>Acc: GET /accounts/{account_id}/balance
    Acc->>Redis: Get cached balance
    Redis-->>Acc: Cache miss
    Acc->>PGB: Query balance
    PGB->>PG: SELECT balance FROM accounts<br/>WHERE id = ?
    PG-->>PGB: Balance: £1,250.50
    PGB-->>Acc: Balance data
    Acc->>Redis: Cache balance<br/>TTL: 60 seconds
    Acc-->>Txn: Balance OK (sufficient funds)

    Txn->>PGB: BEGIN TRANSACTION
    PGB->>PG: BEGIN

    Txn->>PGB: INSERT INTO transactions
    PGB->>PG: INSERT transaction record
    PG-->>PGB: Transaction ID
    PGB-->>Txn: Created

    Txn->>PGB: UPDATE accounts SET balance = balance - amount
    PGB->>PG: UPDATE balance
    PG-->>PGB: Updated
    PGB-->>Txn: Balance updated

    Txn->>PGB: COMMIT TRANSACTION
    PGB->>PG: COMMIT
    PG-->>PGB: Committed
    PGB-->>Txn: Transaction committed

    Txn->>Redis: Invalidate balance cache
    Redis-->>Txn: Deleted

    Txn->>RMQ: Publish message<br/>Exchange: transactions<br/>Event: transaction.created<br/>Payload: {transaction_id, amount, ...}
    RMQ-->>Txn: Message queued

    Txn-->>Client: 201 Created<br/>{transaction}

    RMQ->>Notif: Consume: transaction.created
    Notif->>Notif: Format notification
    Notif->>Client: Push notification<br/>"£25.50 spent at Starbucks"

    RMQ->>Analy: Consume: transaction.created
    Analy->>Redis: Update spending stats<br/>KEY: analytics:{user_id}:daily
    Redis-->>Analy: Updated
```

### Payment Processing Flow (Standing Orders)

```mermaid
sequenceDiagram
    participant Cron as Cron Job<br/>(Scheduler)
    participant Pay as Payment Service<br/>:3006
    participant PGB as PgBouncer
    participant PG as PostgreSQL
    participant Txn as Transaction Service<br/>:3004
    participant RMQ as RabbitMQ
    participant Notif as Notification Service<br/>:3007

    Cron->>Pay: Trigger: Check due payments

    Pay->>PGB: Query due standing orders
    PGB->>PG: SELECT * FROM standing_orders<br/>WHERE status = 'active'<br/>AND next_payment_date <= CURRENT_DATE
    PG-->>PGB: 5 standing orders due
    PGB-->>Pay: Standing orders list

    loop For each standing order
        Pay->>Txn: POST /transactions<br/>{payee_id, amount, reference}

        Txn->>PGB: Check account balance
        PGB->>PG: SELECT balance
        PG-->>PGB: Balance: £2,500
        PGB-->>Txn: Balance OK

        Txn->>PGB: Create transaction
        PGB->>PG: INSERT + UPDATE balance
        PG-->>PGB: Transaction created
        PGB-->>Txn: Transaction ID

        Txn-->>Pay: 201 Created

        Pay->>PGB: Update standing order
        PGB->>PG: UPDATE standing_orders<br/>SET payments_made = payments_made + 1,<br/>next_payment_date = next_payment_date + INTERVAL '1 month',<br/>last_executed_at = CURRENT_TIMESTAMP
        PG-->>PGB: Updated
        PGB-->>Pay: Standing order updated

        Pay->>RMQ: Publish: payment.executed
        RMQ-->>Pay: Queued

        RMQ->>Notif: Consume: payment.executed
        Notif->>Notif: Format notification
        Notif-->>Pay: Push notification sent
    end

    Pay-->>Cron: 5 payments processed
```

---

## Container Resource Allocation

### Memory Limits

| Container | Memory Limit | Memory Reserve | Notes |
|-----------|-------------|----------------|-------|
| PostgreSQL | 512 MB | 256 MB | Database engine |
| PgBouncer | 64 MB | 32 MB | Connection pooler |
| Redis | 512 MB | 256 MB | In-memory cache |
| RabbitMQ | 512 MB | 256 MB | Message queue |
| Auth Service | 128 MB | 64 MB | Go service |
| User Service | 128 MB | 64 MB | Go service |
| Account Service | 128 MB | 64 MB | Go service |
| Transaction Service | 256 MB | 128 MB | Go service (higher load) |
| Card Service | 128 MB | 64 MB | Go service |
| Payment Service | 128 MB | 64 MB | Go service |
| Notification Service | 128 MB | 64 MB | Go service |
| Analytics Service | 256 MB | 128 MB | Go service (data processing) |
| **Total** | **~2.8 GB** | **~1.5 GB** | Minimum 8 GB host RAM recommended |

### CPU Allocation

- **Data Layer**: 2 CPUs total (0.5 per service)
- **Application Layer**: 4 CPUs total (0.5 per service)
- **Host Overhead**: 2 CPUs
- **Total Recommended**: 8+ CPU cores

---

## Network Configuration

### Docker Network

```yaml
Name: protobank-network
Driver: bridge
Subnet: 172.20.0.0/16
Gateway: 172.20.0.1
```

### Port Bindings

```
Host Port → Container Port

External Access (Host → Container):
5432  → postgres:5432          (PostgreSQL)
6432  → pgbouncer:5432         (PgBouncer)
6379  → redis:6379             (Redis)
5672  → rabbitmq:5672          (RabbitMQ AMQP)
15672 → rabbitmq:15672         (RabbitMQ Management UI)
3001  → auth-service:3001      (Auth API)
3002  → user-service:3002      (User API)
3003  → account-service:3003   (Account API)
3004  → transaction-service:3004  (Transaction API)
3005  → card-service:3005      (Card API)
3006  → payment-service:3006   (Payment API)
3007  → notification-service:3007  (Notification API)
3008  → analytics-service:3008 (Analytics API)

Internal Only (Container → Container):
All services can communicate via service names
Example: auth-service → postgres:5432
         transaction-service → rabbitmq:5672
```

### Service Discovery

Containers discover each other by **service name** (DNS):

```go
// Example: Transaction Service connecting to Auth Service
authServiceURL := os.Getenv("AUTH_SERVICE_URL")  // http://auth-service:3001

// Example: Connecting to PostgreSQL via PgBouncer
dbURL := "postgres://postgres:postgres@pgbouncer:5432/protobank"

// Example: Connecting to Redis
redisURL := "redis://:redis@redis:6379/0"

// Example: Connecting to RabbitMQ
rabbitURL := "amqp://admin:admin@rabbitmq:5672/protobank"
```

---

## Volume Persistence

### Data Volumes

```mermaid
graph LR
    subgraph "Docker Host"
        PGV[postgres_data<br/>Volume]
        RDV[redis_data<br/>Volume]
        RMQV[rabbitmq_data<br/>Volume]
    end

    subgraph "Containers"
        PG[PostgreSQL<br/>Container]
        RD[Redis<br/>Container]
        RMQ[RabbitMQ<br/>Container]
    end

    PGV -->|Mount| PG
    RDV -->|Mount| RD
    RMQV -->|Mount| RMQ

    style PGV fill:#336791
    style RDV fill:#DC382D
    style RMQV fill:#FF6600
```

**Volume Details**:

```bash
# List volumes
docker volume ls | grep protobank

# Inspect volume
docker volume inspect protobankbankc_postgres_data

# Location on host
/var/lib/docker/volumes/protobankbankc_postgres_data/_data

# Backup volume
docker run --rm -v protobankbankc_postgres_data:/data \
  -v $(pwd):/backup alpine \
  tar czf /backup/postgres_backup.tar.gz -C /data .

# Restore volume
docker run --rm -v protobankbankc_postgres_data:/data \
  -v $(pwd):/backup alpine \
  tar xzf /backup/postgres_backup.tar.gz -C /data
```

### Read-Only Mounts

```
./database_schema.sql → /docker-entrypoint-initdb.d/01-schema.sql (ro)
./scripts/init-db.sh  → /docker-entrypoint-initdb.d/02-init.sh (ro)
```

These are mounted as **read-only** to prevent accidental modification.

---

## Health Checks

### PostgreSQL Health Check

```bash
# Docker Compose health check
test: ["CMD-SHELL", "pg_isready -U postgres -d protobank"]
interval: 10s
timeout: 5s
retries: 5

# Manual check
docker-compose exec postgres pg_isready -U postgres -d protobank
```

### Redis Health Check

```bash
# Docker Compose health check
test: ["CMD", "redis-cli", "--raw", "incr", "ping"]
interval: 10s
timeout: 5s
retries: 5

# Manual check
docker-compose exec redis redis-cli --pass redis ping
```

### RabbitMQ Health Check

```bash
# Docker Compose health check
test: ["CMD", "rabbitmq-diagnostics", "ping"]
interval: 10s
timeout: 5s
retries: 5

# Manual check
docker-compose exec rabbitmq rabbitmq-diagnostics ping
```

### Service Health Checks

Each Go service should implement a `/health` endpoint:

```go
// Example health check endpoint
func HealthCheckHandler(c *gin.Context) {
    // Check database connection
    if err := db.Ping(); err != nil {
        c.JSON(503, gin.H{"status": "unhealthy", "database": "down"})
        return
    }

    // Check Redis connection
    if err := redisClient.Ping().Err(); err != nil {
        c.JSON(503, gin.H{"status": "unhealthy", "redis": "down"})
        return
    }

    c.JSON(200, gin.H{"status": "healthy"})
}
```

---

## Environment Variable Injection

### Variable Flow

```
.env file → Docker Compose → Container Environment
```

**Example**:

```bash
# .env
POSTGRES_PASSWORD=supersecret
JWT_SECRET=my-jwt-secret

# docker-compose.yml
environment:
  POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-postgres}
  JWT_SECRET: ${JWT_SECRET:-change-this}

# Inside container
echo $POSTGRES_PASSWORD  # supersecret
echo $JWT_SECRET         # my-jwt-secret
```

### Service Environment Variables

Each microservice receives:

| Variable | Example | Purpose |
|----------|---------|---------|
| `SERVICE_NAME` | `auth-service` | Service identifier |
| `SERVICE_PORT` | `3001` | Port to listen on |
| `DATABASE_URL` | `postgres://...@pgbouncer:5432/protobank` | Database connection |
| `REDIS_URL` | `redis://:redis@redis:6379/0` | Redis connection |
| `RABBITMQ_URL` | `amqp://admin:admin@rabbitmq:5672/protobank` | RabbitMQ connection |
| `AUTH_SERVICE_URL` | `http://auth-service:3001` | Auth service URL |
| `LOG_LEVEL` | `debug` | Logging level |

---

## Security Considerations

### Network Isolation

- All services run in private network (`172.20.0.0/16`)
- Only exposed ports are accessible from host
- Services cannot access host network directly

### Secrets Management

**Development**:
- Secrets in `.env` file (NOT committed to Git)
- Default values in `docker-compose.yml` (weak, for dev only)

**Production**:
- Use Docker Secrets or HashiCorp Vault
- Never use `.env` files in production
- Rotate secrets regularly

### Container Security

```dockerfile
# Non-root user in containers
USER appuser

# Read-only filesystem (where possible)
read_only: true

# No privileged mode
privileged: false

# Limited capabilities
cap_drop:
  - ALL
cap_add:
  - NET_BIND_SERVICE
```

---

## Monitoring & Logging

### Container Logs

```bash
# View all logs
docker-compose logs -f

# View specific service
docker-compose logs -f auth-service

# View last 100 lines
docker-compose logs --tail=100 postgres

# Follow logs with timestamps
docker-compose logs -f -t redis
```

### Log Drivers

```yaml
# Configure in docker-compose.yml
logging:
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "3"
```

### Resource Monitoring

```bash
# Monitor container resource usage
docker stats

# Output:
CONTAINER                     CPU %     MEM USAGE / LIMIT     NET I/O
protobank-postgres           2.5%      256MB / 512MB         1.2kB / 850B
protobank-auth-service       0.5%      64MB / 128MB          2.5kB / 1.2kB
```

---

## Troubleshooting

### Common Issues

**Problem**: Port already in use

```bash
# Find process using port
lsof -i :5432

# Change port in .env
POSTGRES_PORT=5433
```

**Problem**: Container won't start

```bash
# Check logs
docker-compose logs postgres

# Remove and recreate
docker-compose down -v
docker-compose up -d
```

**Problem**: Database connection refused

```bash
# Check PostgreSQL is running
docker-compose ps postgres

# Check health
docker-compose exec postgres pg_isready

# Check network
docker network inspect protobankbankc_protobank-network
```

---

## Performance Optimization

### Connection Pooling

**PgBouncer Configuration**:
```
pool_mode = transaction
max_client_conn = 1000
default_pool_size = 25
```

This allows:
- 1000 client connections
- 25 actual PostgreSQL connections
- 40x connection efficiency

### Redis Memory Policy

```
maxmemory 512mb
maxmemory-policy allkeys-lru
```

Evicts least recently used keys when memory limit reached.

### RabbitMQ Tuning

```yaml
environment:
  RABBITMQ_VM_MEMORY_HIGH_WATERMARK: 0.8
  RABBITMQ_DISK_FREE_LIMIT: 2GB
```

---

## Scaling Considerations

### Horizontal Scaling (Multiple Instances)

```yaml
# Scale transaction service to 3 instances
docker-compose up -d --scale transaction-service=3

# Add load balancer (Nginx)
nginx:
  image: nginx:alpine
  volumes:
    - ./nginx/nginx.conf:/etc/nginx/nginx.conf
  ports:
    - "80:80"
```

### Database Read Replicas

```yaml
postgres-replica:
  image: postgres:14-alpine
  environment:
    POSTGRES_PRIMARY_HOST: postgres
    POSTGRES_REPLICATION_USER: replicator
  command: ...streaming replication...
```

---

## Next Steps

1. ✅ Docker Compose setup complete
2. 🔄 Implement Go microservices
3. 🔄 Add Nginx API gateway
4. 🔄 Add monitoring (Prometheus + Grafana)
5. 🔄 Kubernetes migration for production

---

**Last Updated**: 2026-01-30
**Version**: 1.0.0
**Maintainer**: ProtobankBankC Team
