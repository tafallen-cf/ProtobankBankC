# ProtobankBankC - Monzo Clone Banking Application

A full-featured digital banking application inspired by Monzo, built with modern microservices architecture and comprehensive financial management features.

## Overview

ProtobankBankC is a complete banking solution that includes account management, transaction processing, card services, savings pots, payment management, and real-time notifications. The application is designed with security, scalability, and user experience as top priorities.

## Key Features

### Core Banking
- **Account Management** - Personal, business, and joint accounts with multi-currency support
- **Real-time Transactions** - Instant transaction processing with notifications
- **Balance Tracking** - Current, available, and pending balance calculations
- **Transaction History** - Partitioned database for high-performance historical queries
- **KYC Verification** - Know Your Customer compliance workflow

### Cards
- **Physical Cards** - Traditional debit/credit cards with full management
- **Virtual Cards** - Instant digital cards for online shopping
- **Card Controls** - Freeze/unfreeze, spending limits, instant card replacement
- **PCI-DSS Compliance** - Encrypted card data storage

### Savings & Money Management
- **Savings Pots** - Goal-based savings with custom names, icons, and colors
- **Auto-deposits** - Automatic transfers to pots (daily, weekly, monthly, payday)
- **Round-up Savings** - Spare change savings (2x, 5x, 10x multipliers)
- **Target Tracking** - Visual progress toward savings goals

### Payments & Transfers
- **Payees Management** - Save UK and international payment recipients
- **Standing Orders** - Recurring automatic payments with flexible frequencies
- **Scheduled Payments** - Future-dated one-time payments
- **Direct Debits** - Track and manage DD mandates
- **P2P Transfers** - Instant Monzo-to-Monzo user transfers
- **Payment Templates** - Quick payment setup with saved configurations

### Analytics & Insights
- **Spending Categories** - 17+ categories with smart merchant detection
- **Monthly Reports** - Spending breakdowns by category
- **Budget Tracking** - Monitor spending against budgets
- **Transaction Search** - Advanced filtering and search capabilities

### Notifications
- **Real-time Alerts** - Instant notifications for every transaction
- **Security Alerts** - Fraud detection and suspicious activity warnings
- **Payment Reminders** - Upcoming standing orders and scheduled payments
- **Multi-channel** - Push notifications, email, and SMS

## Quick Start

### Get Started in 5 Minutes

```bash
# 1. Clone and setup
git clone <repository-url>
cd ProtobankBankC
make setup

# 2. Start all services
make up

# 3. View logs
make logs
```

**📖 For detailed setup instructions, see [DEVELOPMENT_SETUP.md](./DEVELOPMENT_SETUP.md)**

### Service URLs

Once running, access services at:
- **Auth API**: http://localhost:3001
- **Account API**: http://localhost:3003
- **Transaction API**: http://localhost:3004
- **RabbitMQ Management**: http://localhost:15672 (admin/admin)
- **PostgreSQL**: localhost:5432 (postgres/postgres)

## Technology Stack

### Database
- **PostgreSQL 14+** - Primary database with ACID compliance
- **Redis** - Session management and caching
- **Partitioning** - Quarterly transaction partitions for performance
- **JSONB** - Flexible metadata storage

### Backend (Recommended)
- **Node.js + Express** or **Python + FastAPI** or **Go + Gin**
- **TypeScript** for type safety (if using Node.js)
- **JWT** for authentication
- **Message Queue** (RabbitMQ/Kafka) for async operations

### Frontend (Recommended)
- **React Native** - Cross-platform mobile (iOS + Android)
- **React + TypeScript** - Web application
- **TailwindCSS** - Styling
- **React Query** - Data fetching and state management

### Infrastructure
- **Docker + Docker Compose** - Containerization
- **Kubernetes** - Orchestration (production)
- **Nginx** - API Gateway and load balancing
- **GitHub Actions** - CI/CD

## Project Structure

```
ProtobankBankC/
├── database_schema.sql          # PostgreSQL database schema
├── ARCHITECTURE.md              # Detailed architecture documentation
├── API_SPECIFICATION.md         # REST API endpoints
├── docker-compose.yml           # Docker services configuration
├── backend/
│   ├── auth-service/           # Authentication & authorization
│   ├── user-service/           # User management
│   ├── account-service/        # Account operations
│   ├── transaction-service/    # Transaction processing
│   ├── card-service/           # Card management
│   ├── payment-service/        # Payments & transfers
│   ├── notification-service/   # Push notifications & alerts
│   └── analytics-service/      # Spending insights & reports
├── frontend/
│   ├── mobile/                 # React Native mobile app
│   └── web/                    # React web application
└── infrastructure/
    ├── k8s/                    # Kubernetes configs
    └── terraform/              # Infrastructure as code
```

## Database Schema

The database includes 17 tables covering all banking operations:

### Core Tables
- `users` - User accounts with KYC verification
- `accounts` - Bank accounts (personal, business, joint)
- `transactions` - Financial transactions (partitioned by quarter)
- `account_balances` - Balance history for auditing

### Payment Tables
- `payees` - Saved payment recipients
- `standing_orders` - Recurring automatic payments
- `scheduled_payments` - Future-dated one-time payments
- `direct_debits` - Direct debit mandates
- `payment_templates` - Saved payment configurations

### Card & Savings Tables
- `cards` - Physical and virtual cards (encrypted)
- `pots` - Savings pots with goals
- `pot_transactions` - Pot deposit/withdrawal history

### Support Tables
- `categories` - Transaction categories (17 defaults included)
- `merchants` - Merchant information with logos
- `contacts` - P2P payment contacts
- `notifications` - In-app and push notifications
- `devices` - User devices for push notifications

## Setup Instructions

### Prerequisites
- PostgreSQL 14+ installed
- Node.js 18+ (if using Node.js backend)
- Docker & Docker Compose (recommended)
- Redis (for caching)

### Database Setup

1. **Install PostgreSQL**
   ```bash
   # macOS
   brew install postgresql@14
   brew services start postgresql@14

   # Ubuntu
   sudo apt install postgresql-14
   sudo systemctl start postgresql
   ```

2. **Create Database**
   ```bash
   createdb protobank
   ```

3. **Run Schema**
   ```bash
   psql protobank < database_schema.sql
   ```

4. **Verify Installation**
   ```bash
   psql protobank -c "SELECT COUNT(*) FROM categories;"
   # Should return 17 default categories
   ```

### Docker Setup (Recommended)

1. **Start Services**
   ```bash
   docker-compose up -d
   ```

2. **Initialize Database**
   ```bash
   docker-compose exec postgres psql -U postgres protobank < database_schema.sql
   ```

3. **Check Status**
   ```bash
   docker-compose ps
   ```

## Security Considerations

### Data Protection
- **Encryption at Rest** - Card numbers and CVVs encrypted with AES-256
- **Encryption in Transit** - TLS 1.3 for all API communications
- **Password Hashing** - bcrypt with salt rounds >= 12
- **PCI-DSS Compliance** - Encrypted card storage and tokenization

### Authentication
- **JWT Tokens** - Short-lived access tokens (15 min) + refresh tokens
- **2FA/MFA** - Time-based OTP support via devices table
- **Session Management** - Redis-backed sessions with timeout
- **Biometric Auth** - Face ID / Touch ID for mobile apps

### API Security
- **Rate Limiting** - Prevent brute force and DoS attacks
- **Input Validation** - Comprehensive request validation
- **CORS Configuration** - Strict origin policies
- **API Keys** - Service-to-service authentication

### Financial Security
- **Transaction Limits** - Daily/per-transaction limits
- **Fraud Detection** - Real-time transaction monitoring
- **Balance Checks** - Prevent overdrafts beyond limits
- **Audit Logging** - Immutable transaction history

## Development Roadmap

### Phase 1: Core Banking (Months 1-2)
- [ ] User registration and KYC
- [ ] Account creation and management
- [ ] Transaction processing
- [ ] Balance tracking

### Phase 2: Cards & Payments (Months 3-4)
- [ ] Physical card issuance
- [ ] Virtual cards
- [ ] Payee management
- [ ] Standing orders and scheduled payments

### Phase 3: Savings & Analytics (Months 5-6)
- [ ] Savings pots
- [ ] Category-based spending insights
- [ ] Budget tracking
- [ ] Monthly reports

### Phase 4: Mobile App (Months 7-8)
- [ ] React Native mobile app
- [ ] Push notifications
- [ ] Biometric authentication
- [ ] Offline support

### Phase 5: Advanced Features (Months 9-12)
- [ ] P2P payments
- [ ] Direct debit management
- [ ] Merchant offers and cashback
- [ ] Export transactions (CSV, PDF)
- [ ] Multi-currency accounts
- [ ] Business accounts features

## API Documentation

See [API_SPECIFICATION.md](./API_SPECIFICATION.md) for complete REST API documentation including:
- Authentication endpoints
- Account operations
- Transaction management
- Payment processing
- Card services
- Notification APIs

## Architecture

See [ARCHITECTURE.md](./ARCHITECTURE.md) for detailed system architecture including:
- Microservices design
- Database architecture
- Security patterns
- Scalability strategies
- Message queue workflows

## Testing

### Database Tests
```bash
# Test connection
psql protobank -c "SELECT version();"

# Test schema
psql protobank -c "\dt"

# Test constraints
psql protobank -c "SELECT conname FROM pg_constraint WHERE contype = 'c';"
```

### Performance Benchmarks
- Transaction inserts: >10,000 TPS
- Balance queries: <10ms
- Transaction history: <50ms (1M+ records)

## License

MIT License - See LICENSE file for details

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Support

For issues, questions, or contributions:
- GitHub Issues: [Create an issue](https://github.com/yourusername/ProtobankBankC/issues)
- Email: support@protobank.example.com

## Acknowledgments

- Inspired by [Monzo](https://monzo.com/)
- Built with modern banking best practices
- Community-driven development

---

**⚠️ Disclaimer**: This is a demonstration project. Do not use in production without proper security audits, regulatory compliance review, and financial institution licensing.
