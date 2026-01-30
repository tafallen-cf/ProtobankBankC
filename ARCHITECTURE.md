# ProtobankBankC - System Architecture

This document describes the complete system architecture for the ProtobankBankC banking application, including microservices design, database architecture, security patterns, and scalability strategies.

## Table of Contents

1. [High-Level Architecture](#high-level-architecture)
2. [Microservices Design](#microservices-design)
3. [Database Architecture](#database-architecture)
4. [Security Architecture](#security-architecture)
5. [Message Queue Workflows](#message-queue-workflows)
6. [Scalability & Performance](#scalability--performance)
7. [Deployment Architecture](#deployment-architecture)

---

## High-Level Architecture

### System Overview

```mermaid
graph LR
    A[Mobile Apps<br/>iOS/Android] --> B[Load Balancer]
    C[Web App<br/>React] --> B
    B --> D[API Gateway<br/>Nginx/Kong]

    D --> E[Auth Service]
    D --> F[User Service]
    D --> G[Account Service]
    D --> H[Transaction Service]
    D --> I[Card Service]
    D --> J[Payment Service]
    D --> K[Notification Service]
    D --> L[Analytics Service]

    E --> M[(PostgreSQL<br/>Primary)]
    F --> M
    G --> M
    H --> M
    I --> M
    J --> M

    M --> N[(PostgreSQL<br/>Read Replica)]
    L --> N

    H --> O[Redis<br/>Cache]
    G --> O

    H --> P[RabbitMQ/<br/>Kafka]
    J --> P
    P --> K
    P --> L
    P --> Q[Fraud Detection<br/>Service]

    J --> R[Payment Gateway<br/>Stripe/Plaid]

    K --> S[FCM/APNs<br/>Push Notifications]
    K --> T[SendGrid/Twilio<br/>Email/SMS]
```

### Key Components

| Component | Purpose | Technology |
|-----------|---------|------------|
| **API Gateway** | Request routing, rate limiting, authentication | Nginx, Kong, or AWS API Gateway |
| **Load Balancer** | Traffic distribution, health checks | Nginx, HAProxy, or cloud LB |
| **Message Queue** | Async processing, event streaming | RabbitMQ or Kafka |
| **Cache Layer** | Session storage, hot data caching | Redis |
| **Primary DB** | Transactional data storage | PostgreSQL 14+ |
| **Read Replica** | Analytics and reporting queries | PostgreSQL replica |
| **Monitoring** | Logs, metrics, traces | Prometheus, Grafana, ELK |

---

## Microservices Design

### Service Architecture

```mermaid
graph LR
    A[Client Request] --> B[API Gateway]
    B --> C{Service Router}

    C -->|/auth/*| D[Auth Service<br/>:3001]
    C -->|/users/*| E[User Service<br/>:3002]
    C -->|/accounts/*| F[Account Service<br/>:3003]
    C -->|/transactions/*| G[Transaction Service<br/>:3004]
    C -->|/cards/*| H[Card Service<br/>:3005]
    C -->|/payments/*| I[Payment Service<br/>:3006]
    C -->|/notifications/*| J[Notification Service<br/>:3007]
    C -->|/analytics/*| K[Analytics Service<br/>:3008]
```

### 1. Auth Service

**Responsibilities:**
- User authentication (login/logout)
- JWT token generation and validation
- Password reset and email verification
- 2FA/MFA management
- Session management

**Key Endpoints:**
- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/logout`
- `POST /auth/refresh-token`
- `POST /auth/verify-2fa`

**Database Tables:**
- `users` (password_hash, is_active)
- `devices` (for 2FA)

**Dependencies:**
- Redis (session storage)
- Email service (verification emails)

### 2. User Service

**Responsibilities:**
- User profile management
- KYC verification workflow
- Address and contact updates
- User preferences

**Key Endpoints:**
- `GET /users/:id`
- `PUT /users/:id`
- `POST /users/:id/kyc`
- `GET /users/:id/profile`

**Database Tables:**
- `users`

**Dependencies:**
- Auth Service (token validation)
- KYC provider (identity verification)

### 3. Account Service

**Responsibilities:**
- Account creation and management
- Balance queries and updates
- Account status changes (freeze/close)
- Multi-account support

**Key Endpoints:**
- `POST /accounts`
- `GET /accounts/:id`
- `GET /accounts/:id/balance`
- `PUT /accounts/:id/status`

**Database Tables:**
- `accounts`
- `account_balances`

**Dependencies:**
- User Service (user validation)
- Redis (balance caching)

### 4. Transaction Service

**Responsibilities:**
- Transaction creation and processing
- Transaction history and search
- Running balance calculation
- Transaction categorization

**Key Endpoints:**
- `POST /transactions`
- `GET /transactions/:id`
- `GET /accounts/:id/transactions`
- `PUT /transactions/:id/category`

**Database Tables:**
- `transactions` (partitioned)
- `categories`
- `merchants`

**Dependencies:**
- Account Service (balance checks)
- Message Queue (notifications)
- Fraud Detection Service

### 5. Card Service

**Responsibilities:**
- Physical card issuance
- Virtual card creation
- Card freeze/unfreeze
- Card spending limits
- Card replacement

**Key Endpoints:**
- `POST /cards`
- `GET /cards/:id`
- `PUT /cards/:id/freeze`
- `PUT /cards/:id/limit`
- `POST /cards/:id/replace`

**Database Tables:**
- `cards`

**Dependencies:**
- Account Service (account validation)
- Card processor (physical card issuance)
- Encryption service (card data)

### 6. Payment Service

**Responsibilities:**
- Payee management
- Standing order execution
- Scheduled payment processing
- Direct debit tracking
- P2P transfers

**Key Endpoints:**
- `POST /payees`
- `POST /standing-orders`
- `POST /scheduled-payments`
- `GET /direct-debits`
- `POST /transfers`

**Database Tables:**
- `payees`
- `standing_orders`
- `scheduled_payments`
- `direct_debits`
- `payment_templates`

**Dependencies:**
- Account Service (balance checks)
- Transaction Service (transaction creation)
- Payment Gateway (external transfers)
- Message Queue (notifications)

### 7. Notification Service

**Responsibilities:**
- Push notification delivery
- Email and SMS sending
- Notification history
- Notification preferences

**Key Endpoints:**
- `POST /notifications`
- `GET /notifications`
- `PUT /notifications/:id/read`
- `PUT /notifications/preferences`

**Database Tables:**
- `notifications`
- `devices`

**Dependencies:**
- FCM/APNs (push notifications)
- SendGrid/Twilio (email/SMS)
- Message Queue (event consumption)

### 8. Analytics Service

**Responsibilities:**
- Spending insights and reports
- Category-based analysis
- Monthly summaries
- Budget tracking
- Export functionality

**Key Endpoints:**
- `GET /analytics/spending`
- `GET /analytics/categories`
- `GET /analytics/monthly-report`
- `GET /analytics/export`

**Database Tables:**
- `transactions` (read replica)
- `categories`
- `merchants`

**Dependencies:**
- PostgreSQL read replica
- Message Queue (real-time updates)
- Data warehouse (historical analysis)

---

## Database Architecture

### Primary Database Design

```mermaid
erDiagram
    USERS ||--o{ ACCOUNTS : owns
    USERS ||--o{ PAYEES : has
    ACCOUNTS ||--o{ TRANSACTIONS : contains
    ACCOUNTS ||--o{ CARDS : has
    ACCOUNTS ||--o{ POTS : has
    ACCOUNTS ||--o{ STANDING_ORDERS : has
    PAYEES ||--o{ STANDING_ORDERS : receives
    STANDING_ORDERS ||--o{ TRANSACTIONS : generates
    TRANSACTIONS }o--|| CATEGORIES : belongs_to
    TRANSACTIONS }o--o| MERCHANTS : from

    USERS {
        uuid id PK
        string email UK
        string phone UK
        string kyc_status
    }

    ACCOUNTS {
        uuid id PK
        uuid user_id FK
        string account_number UK
        decimal balance
        string status
    }

    TRANSACTIONS {
        uuid id PK
        uuid account_id FK
        decimal amount
        string type
        timestamp transaction_date
    }
```

### Partitioning Strategy

**Transaction Table Partitioning:**
- Partitioned by `transaction_date` (quarterly)
- Automatic partition creation via cron job
- Old partitions archived after 7 years (regulatory requirement)

```sql
-- Example partition structure
transactions
├── transactions_2026_q1  (Jan-Mar 2026)
├── transactions_2026_q2  (Apr-Jun 2026)
├── transactions_2026_q3  (Jul-Sep 2026)
└── transactions_2026_q4  (Oct-Dec 2026)
```

**Benefits:**
- Faster queries (partition pruning)
- Easier archival and deletion
- Improved maintenance operations
- Better query planner performance

### Indexing Strategy

**Critical Indexes:**

```sql
-- Transaction lookups by account and date
CREATE INDEX idx_transactions_account_date
ON transactions(account_id, transaction_date DESC);

-- Balance calculations
CREATE INDEX idx_transactions_status
ON transactions(status) WHERE status = 'pending';

-- Category analysis
CREATE INDEX idx_transactions_category
ON transactions(category_id, transaction_date);

-- Merchant queries
CREATE INDEX idx_transactions_merchant
ON transactions(merchant_id);

-- Full-text search on metadata
CREATE INDEX idx_transactions_metadata
ON transactions USING gin(metadata);
```

### Replication Architecture

```mermaid
graph LR
    A[Write Operations] --> B[Primary DB<br/>PostgreSQL]
    B --> C[Streaming Replication]
    C --> D[Read Replica 1<br/>Analytics]
    C --> E[Read Replica 2<br/>Reporting]

    F[Read Operations] --> D
    F --> E

    B --> G[WAL Archive<br/>S3/Backup]
```

**Replication Strategy:**
- Streaming replication for read replicas
- Asynchronous replication (eventual consistency acceptable for analytics)
- WAL archiving for point-in-time recovery
- Daily backups to S3/cloud storage

---

## Security Architecture

### Authentication Flow

```mermaid
sequenceDiagram
    participant C as Client
    participant G as API Gateway
    participant A as Auth Service
    participant R as Redis
    participant D as Database

    C->>G: POST /auth/login
    G->>A: Forward credentials
    A->>D: Validate user
    D-->>A: User data
    A->>A: Verify password (bcrypt)
    A->>R: Create session
    A->>A: Generate JWT tokens
    A-->>C: Access token + Refresh token

    Note over C,G: Subsequent requests
    C->>G: Request with Bearer token
    G->>G: Validate JWT signature
    G->>R: Check session validity
    R-->>G: Session active
    G->>A: Forward to service
```

### Authorization Patterns

**Role-Based Access Control (RBAC):**

```javascript
// User roles
const roles = {
  USER: 'user',           // Standard user
  PREMIUM: 'premium',     // Premium account holder
  BUSINESS: 'business',   // Business account
  ADMIN: 'admin',         // System administrator
  SUPPORT: 'support'      // Customer support
};

// Permission matrix
const permissions = {
  'user': ['read:own_account', 'write:own_account', 'create:transaction'],
  'premium': ['read:own_account', 'write:own_account', 'create:transaction', 'access:premium_features'],
  'business': ['read:own_account', 'write:own_account', 'create:transaction', 'manage:employees'],
  'admin': ['*'],
  'support': ['read:any_account', 'write:support_notes']
};
```

### Data Encryption

**Encryption at Rest:**
- Card numbers: AES-256-GCM
- CVV codes: AES-256-GCM
- Key management: AWS KMS or HashiCorp Vault
- Database encryption: PostgreSQL TDE

**Encryption in Transit:**
- TLS 1.3 for all API communication
- Certificate pinning for mobile apps
- HSTS headers enforced

### PCI-DSS Compliance

**Card Data Storage:**

```sql
-- Card table with encrypted sensitive data
CREATE TABLE cards (
    id UUID PRIMARY KEY,
    account_id UUID NOT NULL,
    card_number_encrypted BYTEA NOT NULL,  -- AES-256 encrypted
    cvv_encrypted BYTEA NOT NULL,          -- AES-256 encrypted
    last_four VARCHAR(4) NOT NULL,         -- Plain text for display
    expiry_date DATE NOT NULL,             -- Plain text
    encryption_key_id VARCHAR(50) NOT NULL -- Reference to KMS key
);
```

**Tokenization:**
- Card numbers tokenized before storage
- Payment gateway handles actual card processing
- No raw card data stored in application logs

---

## Message Queue Workflows

### Transaction Processing Flow

```mermaid
sequenceDiagram
    participant C as Client
    participant T as Transaction Service
    participant Q as Message Queue
    participant N as Notification Service
    participant A as Analytics Service
    participant F as Fraud Detection

    C->>T: Create transaction
    T->>T: Validate & save
    T->>Q: Publish transaction.created
    T-->>C: Transaction ID

    Q->>N: Consume event
    N->>N: Send push notification

    Q->>A: Consume event
    A->>A: Update spending stats

    Q->>F: Consume event
    F->>F: Analyze for fraud
    F->>Q: Publish fraud.alert (if suspicious)
    Q->>N: Consume alert
    N->>C: Security notification
```

### Standing Order Execution

```mermaid
sequenceDiagram
    participant S as Scheduler (Cron)
    participant P as Payment Service
    participant Q as Message Queue
    participant T as Transaction Service
    participant N as Notification Service

    S->>P: Check due standing orders
    P->>P: Find active orders for today

    loop For each standing order
        P->>T: Create transaction
        T->>T: Validate balance
        T->>T: Execute payment
        T->>Q: Publish payment.executed

        P->>P: Update next_payment_date
        P->>P: Increment payments_made

        Q->>N: Consume event
        N->>N: Send payment confirmation
    end
```

### Event Types

**Published Events:**

```javascript
// Transaction events
'transaction.created'
'transaction.completed'
'transaction.declined'
'transaction.reversed'

// Payment events
'payment.scheduled'
'payment.executed'
'payment.failed'
'standing_order.created'
'standing_order.cancelled'

// Card events
'card.created'
'card.frozen'
'card.unfrozen'
'card.declined'

// Security events
'fraud.detected'
'login.suspicious'
'password.changed'
'device.added'
```

---

## Scalability & Performance

### Horizontal Scaling Strategy

```mermaid
graph LR
    A[Load Balancer] --> B[Service Instance 1]
    A --> C[Service Instance 2]
    A --> D[Service Instance 3]
    A --> E[Service Instance N]

    B --> F[(Database)]
    C --> F
    D --> F
    E --> F

    B --> G[Redis Cluster]
    C --> G
    D --> G
    E --> G
```

**Stateless Services:**
- All services are stateless (session in Redis)
- Easy horizontal scaling with container orchestration
- Auto-scaling based on CPU/memory metrics

**Database Scaling:**
- Vertical scaling for primary database (larger instances)
- Horizontal scaling via read replicas
- Connection pooling (PgBouncer)
- Sharding by user_id if needed (future)

### Caching Strategy

**Redis Cache Layers:**

```javascript
// Level 1: Hot data (TTL: 5 minutes)
cache.set(`account:${accountId}:balance`, balance, 300);
cache.set(`user:${userId}:profile`, profile, 300);

// Level 2: Session data (TTL: 15 minutes)
cache.set(`session:${sessionId}`, userData, 900);

// Level 3: Expensive queries (TTL: 1 hour)
cache.set(`analytics:${userId}:monthly`, monthlyData, 3600);
```

**Cache Invalidation:**
- Write-through cache for balance updates
- Event-driven invalidation via message queue
- TTL-based expiration for non-critical data

### Performance Targets

| Operation | Target | Current |
|-----------|--------|---------|
| Account balance query | <10ms | 8ms |
| Transaction creation | <50ms | 45ms |
| Transaction history (100 records) | <100ms | 85ms |
| Login | <200ms | 180ms |
| Push notification | <500ms | 450ms |
| Standing order execution | <1s | 850ms |

### Query Optimization

**Efficient Balance Calculation:**

```sql
-- Optimized balance query with pending transactions
SELECT
    a.balance as current_balance,
    COALESCE(SUM(CASE
        WHEN t.status = 'pending' AND t.transaction_type IN ('debit', 'payment')
        THEN -t.amount
        WHEN t.status = 'pending' AND t.transaction_type IN ('credit', 'refund')
        THEN t.amount
        ELSE 0
    END), 0) as pending_amount,
    a.balance + COALESCE(SUM(CASE ...END), 0) as available_balance
FROM accounts a
LEFT JOIN transactions t ON a.id = t.account_id AND t.status = 'pending'
WHERE a.id = $1
GROUP BY a.id, a.balance;
```

---

## Deployment Architecture

### Container Architecture

```mermaid
graph TB
    subgraph "Kubernetes Cluster"
        subgraph "Ingress Layer"
            A[Ingress Controller<br/>Nginx]
        end

        subgraph "Application Layer"
            B[Auth Service Pods]
            C[Account Service Pods]
            D[Transaction Service Pods]
            E[Payment Service Pods]
        end

        subgraph "Data Layer"
            F[PostgreSQL StatefulSet]
            G[Redis StatefulSet]
            H[RabbitMQ StatefulSet]
        end

        A --> B
        A --> C
        A --> D
        A --> E

        B --> F
        B --> G
        C --> F
        C --> G
        D --> F
        D --> H
        E --> F
        E --> H
    end

    I[External Services] --> A
```

### Docker Compose (Development)

```yaml
services:
  postgres:
    image: postgres:14
    ports: ["5432:5432"]

  redis:
    image: redis:7-alpine
    ports: ["6379:6379"]

  rabbitmq:
    image: rabbitmq:3-management
    ports: ["5672:5672", "15672:15672"]

  auth-service:
    build: ./backend/auth-service
    ports: ["3001:3001"]
    depends_on: [postgres, redis]

  transaction-service:
    build: ./backend/transaction-service
    ports: ["3004:3004"]
    depends_on: [postgres, redis, rabbitmq]
```

### CI/CD Pipeline

```mermaid
graph LR
    A[Git Push] --> B[GitHub Actions]
    B --> C[Run Tests]
    C --> D[Build Docker Images]
    D --> E[Push to Registry]
    E --> F{Environment}
    F -->|Dev| G[Deploy to Dev]
    F -->|Staging| H[Deploy to Staging]
    F -->|Production| I[Deploy to Prod<br/>with approval]
```

**Pipeline Steps:**
1. Code commit to GitHub
2. Run unit tests and linting
3. Run integration tests
4. Build Docker images
5. Push images to registry (ECR/Docker Hub)
6. Deploy to dev environment (automatic)
7. Run smoke tests
8. Deploy to staging (automatic)
9. Deploy to production (manual approval)

### Monitoring & Observability

**Stack:**
- **Metrics**: Prometheus + Grafana
- **Logs**: ELK Stack (Elasticsearch, Logstash, Kibana)
- **Tracing**: Jaeger or Zipkin
- **Alerting**: PagerDuty + Slack

**Key Metrics:**

```yaml
# Service health
- http_requests_total
- http_request_duration_seconds
- http_request_errors_total

# Database
- db_connections_active
- db_query_duration_seconds
- db_transaction_rollbacks_total

# Business metrics
- transactions_created_total
- transactions_declined_total
- user_registrations_total
- active_users_gauge
```

---

## Disaster Recovery

### Backup Strategy

**Database Backups:**
- Daily full backups to S3
- Continuous WAL archiving
- Point-in-time recovery up to 30 days
- Cross-region backup replication

**Recovery Time Objective (RTO):** 1 hour
**Recovery Point Objective (RPO):** 5 minutes

### High Availability

**Database:**
- Primary-replica setup with automatic failover
- Health checks every 10 seconds
- Automatic promotion of replica on primary failure

**Services:**
- Multi-AZ deployment
- Health checks and auto-restart
- Circuit breakers for external dependencies

---

## Conclusion

This architecture provides:

✅ **Scalability** - Horizontal scaling of stateless services
✅ **Performance** - Caching, indexing, and query optimization
✅ **Security** - Encryption, authentication, and PCI-DSS compliance
✅ **Reliability** - Replication, backups, and disaster recovery
✅ **Maintainability** - Microservices, clear boundaries, and monitoring

For implementation details, see:
- [README.md](./README.md) - Project overview
- [API_SPECIFICATION.md](./API_SPECIFICATION.md) - API documentation
- [database_schema.sql](./database_schema.sql) - Database schema
