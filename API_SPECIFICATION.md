# ProtobankBankC - API Specification

Complete REST API documentation for all microservices.

## Table of Contents

1. [General Information](#general-information)
2. [Authentication Service](#authentication-service)
3. [User Service](#user-service)
4. [Account Service](#account-service)
5. [Transaction Service](#transaction-service)
6. [Card Service](#card-service)
7. [Payment Service](#payment-service)
8. [Notification Service](#notification-service)
9. [Analytics Service](#analytics-service)

---

## General Information

### Base URLs

| Environment | Base URL |
|-------------|----------|
| Development | `http://localhost:3000/api/v1` |
| Staging | `https://staging-api.protobank.example.com/api/v1` |
| Production | `https://api.protobank.example.com/api/v1` |

### Authentication

All endpoints (except auth endpoints) require a Bearer token in the Authorization header:

```http
Authorization: Bearer <access_token>
```

### Common Response Formats

**Success Response:**
```json
{
  "success": true,
  "data": { ... },
  "message": "Operation successful"
}
```

**Error Response:**
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": [
      {
        "field": "email",
        "message": "Invalid email format"
      }
    ]
  }
}
```

### Standard Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `UNAUTHORIZED` | 401 | Missing or invalid authentication |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `VALIDATION_ERROR` | 400 | Invalid request data |
| `CONFLICT` | 409 | Resource conflict |
| `RATE_LIMIT_EXCEEDED` | 429 | Too many requests |
| `INTERNAL_ERROR` | 500 | Server error |

### Rate Limiting

- **Authentication endpoints**: 5 requests per minute
- **Standard endpoints**: 100 requests per minute
- **Transaction endpoints**: 50 requests per minute

---

## Authentication Service

Base path: `/auth`

### Register User

```http
POST /auth/register
```

**Request Body:**
```json
{
  "email": "john.doe@example.com",
  "phone": "+447700900000",
  "password": "SecurePass123!",
  "first_name": "John",
  "last_name": "Doe",
  "date_of_birth": "1990-01-15",
  "address": {
    "line1": "123 High Street",
    "line2": "Apt 4B",
    "city": "London",
    "postcode": "SW1A 1AA",
    "country": "GB"
  }
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "john.doe@example.com",
    "verification_required": true
  },
  "message": "User registered successfully. Please verify your email."
}
```

---

### Login

```http
POST /auth/login
```

**Request Body:**
```json
{
  "email": "john.doe@example.com",
  "password": "SecurePass123!",
  "device_id": "device-uuid-here",
  "device_type": "ios"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 900,
    "token_type": "Bearer",
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "john.doe@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "kyc_status": "verified"
    }
  }
}
```

---

### Refresh Token

```http
POST /auth/refresh-token
```

**Request Body:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 900,
    "token_type": "Bearer"
  }
}
```

---

### Logout

```http
POST /auth/logout
```

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Logged out successfully"
}
```

---

### Request Password Reset

```http
POST /auth/forgot-password
```

**Request Body:**
```json
{
  "email": "john.doe@example.com"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Password reset email sent"
}
```

---

### Reset Password

```http
POST /auth/reset-password
```

**Request Body:**
```json
{
  "token": "reset-token-from-email",
  "new_password": "NewSecurePass123!"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Password reset successfully"
}
```

---

## User Service

Base path: `/users`

### Get User Profile

```http
GET /users/{user_id}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "john.doe@example.com",
    "phone": "+447700900000",
    "first_name": "John",
    "last_name": "Doe",
    "date_of_birth": "1990-01-15",
    "address": {
      "line1": "123 High Street",
      "line2": "Apt 4B",
      "city": "London",
      "postcode": "SW1A 1AA",
      "country": "GB"
    },
    "kyc_status": "verified",
    "kyc_verified_at": "2026-01-15T10:30:00Z",
    "is_active": true,
    "created_at": "2026-01-10T14:20:00Z",
    "updated_at": "2026-01-15T10:30:00Z"
  }
}
```

---

### Update User Profile

```http
PUT /users/{user_id}
```

**Request Body:**
```json
{
  "phone": "+447700900001",
  "address": {
    "line1": "456 Main Street",
    "city": "Manchester",
    "postcode": "M1 1AA"
  }
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "phone": "+447700900001",
    "address": {
      "line1": "456 Main Street",
      "line2": "Apt 4B",
      "city": "Manchester",
      "postcode": "M1 1AA",
      "country": "GB"
    },
    "updated_at": "2026-01-30T11:00:00Z"
  }
}
```

---

### Submit KYC Documents

```http
POST /users/{user_id}/kyc
```

**Request (Multipart Form Data):**
```
document_type: "passport"
document_front: <file>
document_back: <file>
selfie: <file>
```

**Response (202 Accepted):**
```json
{
  "success": true,
  "data": {
    "kyc_status": "review",
    "submission_id": "kyc-123456",
    "estimated_review_time": "24-48 hours"
  },
  "message": "KYC documents submitted for review"
}
```

---

## Account Service

Base path: `/accounts`

### Create Account

```http
POST /accounts
```

**Request Body:**
```json
{
  "account_type": "personal",
  "currency": "GBP"
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "id": "acc-550e8400-e29b-41d4-a716-446655440000",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "account_number": "12345678",
    "sort_code": "040004",
    "account_type": "personal",
    "balance": 0.00,
    "currency": "GBP",
    "is_primary": true,
    "status": "active",
    "created_at": "2026-01-30T11:00:00Z"
  }
}
```

---

### Get Account Details

```http
GET /accounts/{account_id}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "acc-550e8400-e29b-41d4-a716-446655440000",
    "account_number": "12345678",
    "sort_code": "040004",
    "account_type": "personal",
    "balance": 1250.50,
    "currency": "GBP",
    "status": "active",
    "is_primary": true,
    "created_at": "2026-01-10T14:20:00Z"
  }
}
```

---

### Get Account Balance

```http
GET /accounts/{account_id}/balance
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "account_id": "acc-550e8400-e29b-41d4-a716-446655440000",
    "current_balance": 1250.50,
    "available_balance": 1180.50,
    "pending_balance": -70.00,
    "currency": "GBP",
    "overdraft_limit": 1000.00,
    "as_of": "2026-01-30T11:00:00Z"
  }
}
```

---

### List User Accounts

```http
GET /users/{user_id}/accounts
```

**Query Parameters:**
- `status` (optional): `active`, `frozen`, `closed`

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "accounts": [
      {
        "id": "acc-550e8400-e29b-41d4-a716-446655440000",
        "account_number": "12345678",
        "sort_code": "040004",
        "account_type": "personal",
        "balance": 1250.50,
        "currency": "GBP",
        "is_primary": true,
        "status": "active"
      }
    ],
    "total": 1
  }
}
```

---

### Freeze Account

```http
PUT /accounts/{account_id}/freeze
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "acc-550e8400-e29b-41d4-a716-446655440000",
    "status": "frozen",
    "updated_at": "2026-01-30T11:00:00Z"
  },
  "message": "Account frozen successfully"
}
```

---

### Unfreeze Account

```http
PUT /accounts/{account_id}/unfreeze
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "acc-550e8400-e29b-41d4-a716-446655440000",
    "status": "active",
    "updated_at": "2026-01-30T11:00:00Z"
  },
  "message": "Account unfrozen successfully"
}
```

---

## Transaction Service

Base path: `/transactions`

### Create Transaction

```http
POST /transactions
```

**Request Body:**
```json
{
  "account_id": "acc-550e8400-e29b-41d4-a716-446655440000",
  "transaction_type": "debit",
  "amount": 25.50,
  "currency": "GBP",
  "description": "Coffee at Starbucks",
  "merchant_id": "merchant-123",
  "category_id": "cat-eating-out"
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "id": "txn-550e8400-e29b-41d4-a716-446655440000",
    "account_id": "acc-550e8400-e29b-41d4-a716-446655440000",
    "transaction_type": "debit",
    "amount": -25.50,
    "currency": "GBP",
    "description": "Coffee at Starbucks",
    "running_balance": 1225.00,
    "status": "completed",
    "transaction_date": "2026-01-30T11:00:00Z",
    "created_at": "2026-01-30T11:00:00Z"
  }
}
```

---

### Get Transaction Details

```http
GET /transactions/{transaction_id}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "txn-550e8400-e29b-41d4-a716-446655440000",
    "account_id": "acc-550e8400-e29b-41d4-a716-446655440000",
    "transaction_type": "debit",
    "amount": -25.50,
    "currency": "GBP",
    "description": "Coffee at Starbucks",
    "merchant": {
      "id": "merchant-123",
      "name": "Starbucks",
      "logo_url": "https://cdn.protobank.com/merchants/starbucks.png"
    },
    "category": {
      "id": "cat-eating-out",
      "name": "Eating Out",
      "icon": "🍽️",
      "color": "#EA4C89"
    },
    "running_balance": 1225.00,
    "status": "completed",
    "transaction_date": "2026-01-30T11:00:00Z",
    "settled_at": "2026-01-30T11:00:05Z",
    "created_at": "2026-01-30T11:00:00Z"
  }
}
```

---

### List Account Transactions

```http
GET /accounts/{account_id}/transactions
```

**Query Parameters:**
- `page` (default: 1)
- `limit` (default: 50, max: 100)
- `start_date` (optional): ISO 8601 date
- `end_date` (optional): ISO 8601 date
- `category_id` (optional): Filter by category
- `status` (optional): `pending`, `completed`, `declined`
- `type` (optional): `debit`, `credit`, `transfer_out`, `transfer_in`

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "transactions": [
      {
        "id": "txn-1",
        "transaction_type": "debit",
        "amount": -25.50,
        "description": "Coffee at Starbucks",
        "merchant": {
          "name": "Starbucks",
          "logo_url": "https://cdn.protobank.com/merchants/starbucks.png"
        },
        "category": {
          "name": "Eating Out",
          "icon": "🍽️"
        },
        "running_balance": 1225.00,
        "status": "completed",
        "transaction_date": "2026-01-30T11:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 50,
      "total_pages": 5,
      "total_records": 234
    }
  }
}
```

---

### Update Transaction Category

```http
PUT /transactions/{transaction_id}/category
```

**Request Body:**
```json
{
  "category_id": "cat-groceries"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "txn-550e8400-e29b-41d4-a716-446655440000",
    "category": {
      "id": "cat-groceries",
      "name": "Groceries",
      "icon": "🛒",
      "color": "#EE7B30"
    },
    "updated_at": "2026-01-30T11:05:00Z"
  }
}
```

---

### Add Transaction Notes

```http
PUT /transactions/{transaction_id}/notes
```

**Request Body:**
```json
{
  "notes": "Team lunch with colleagues"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "txn-550e8400-e29b-41d4-a716-446655440000",
    "notes": "Team lunch with colleagues",
    "updated_at": "2026-01-30T11:05:00Z"
  }
}
```

---

### List Categories

```http
GET /transactions/categories
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "categories": [
      {
        "id": "cat-groceries",
        "name": "Groceries",
        "icon": "🛒",
        "color": "#EE7B30",
        "is_system": true
      },
      {
        "id": "cat-eating-out",
        "name": "Eating Out",
        "icon": "🍽️",
        "color": "#EA4C89",
        "is_system": true
      }
    ],
    "total": 17
  }
}
```

---

## Card Service

Base path: `/cards`

### Create Physical Card

```http
POST /cards
```

**Request Body:**
```json
{
  "account_id": "acc-550e8400-e29b-41d4-a716-446655440000",
  "card_type": "debit",
  "spending_limit": 1000.00
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "id": "card-550e8400-e29b-41d4-a716-446655440000",
    "account_id": "acc-550e8400-e29b-41d4-a716-446655440000",
    "card_type": "debit",
    "last_four": "1234",
    "expiry_date": "2028-01-31",
    "is_frozen": false,
    "is_virtual": false,
    "spending_limit": 1000.00,
    "status": "active",
    "estimated_delivery": "5-7 business days",
    "created_at": "2026-01-30T11:00:00Z"
  },
  "message": "Physical card ordered. You'll receive it in 5-7 business days."
}
```

---

### Create Virtual Card

```http
POST /cards/virtual
```

**Request Body:**
```json
{
  "account_id": "acc-550e8400-e29b-41d4-a716-446655440000",
  "spending_limit": 500.00
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "id": "card-virtual-550e8400-e29b-41d4-a716-446655440000",
    "account_id": "acc-550e8400-e29b-41d4-a716-446655440000",
    "card_type": "virtual",
    "card_number": "4532 1234 5678 9010",
    "last_four": "9010",
    "cvv": "123",
    "expiry_date": "2028-01-31",
    "is_frozen": false,
    "is_virtual": true,
    "spending_limit": 500.00,
    "created_at": "2026-01-30T11:00:00Z"
  },
  "message": "Virtual card created successfully. Use immediately for online purchases."
}
```

---

### Get Card Details

```http
GET /cards/{card_id}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "card-550e8400-e29b-41d4-a716-446655440000",
    "account_id": "acc-550e8400-e29b-41d4-a716-446655440000",
    "card_type": "debit",
    "last_four": "1234",
    "expiry_date": "2028-01-31",
    "is_frozen": false,
    "is_virtual": false,
    "spending_limit": 1000.00,
    "created_at": "2026-01-20T10:00:00Z"
  }
}
```

---

### Freeze Card

```http
PUT /cards/{card_id}/freeze
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "card-550e8400-e29b-41d4-a716-446655440000",
    "is_frozen": true,
    "updated_at": "2026-01-30T11:00:00Z"
  },
  "message": "Card frozen successfully"
}
```

---

### Unfreeze Card

```http
PUT /cards/{card_id}/unfreeze
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "card-550e8400-e29b-41d4-a716-446655440000",
    "is_frozen": false,
    "updated_at": "2026-01-30T11:00:00Z"
  },
  "message": "Card unfrozen successfully"
}
```

---

### Update Spending Limit

```http
PUT /cards/{card_id}/limit
```

**Request Body:**
```json
{
  "spending_limit": 2000.00
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "card-550e8400-e29b-41d4-a716-446655440000",
    "spending_limit": 2000.00,
    "updated_at": "2026-01-30T11:00:00Z"
  }
}
```

---

### Replace Card

```http
POST /cards/{card_id}/replace
```

**Request Body:**
```json
{
  "reason": "lost",
  "delivery_address": {
    "line1": "123 High Street",
    "city": "London",
    "postcode": "SW1A 1AA"
  }
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "old_card_id": "card-550e8400-e29b-41d4-a716-446655440000",
    "new_card_id": "card-new-550e8400-e29b-41d4-a716-446655440000",
    "new_last_four": "5678",
    "estimated_delivery": "5-7 business days"
  },
  "message": "Card replacement ordered. Old card has been deactivated."
}
```

---

## Payment Service

Base path: `/payments`

### Create Payee

```http
POST /payees
```

**Request Body (UK Bank):**
```json
{
  "payee_type": "uk_bank",
  "name": "John Smith",
  "account_number": "12345678",
  "sort_code": "040004",
  "reference": "Rent payment"
}
```

**Request Body (International):**
```json
{
  "payee_type": "international",
  "name": "Marie Dubois",
  "iban": "FR1420041010050500013M02606",
  "swift_code": "BNPAFRPP",
  "reference": "Invoice 12345"
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "id": "payee-550e8400-e29b-41d4-a716-446655440000",
    "payee_type": "uk_bank",
    "name": "John Smith",
    "account_number": "12345678",
    "sort_code": "040004",
    "is_verified": false,
    "is_favorite": false,
    "created_at": "2026-01-30T11:00:00Z"
  }
}
```

---

### List Payees

```http
GET /payees
```

**Query Parameters:**
- `is_favorite` (optional): `true`, `false`

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "payees": [
      {
        "id": "payee-1",
        "name": "John Smith",
        "payee_type": "uk_bank",
        "account_number": "12345678",
        "sort_code": "040004",
        "is_favorite": true,
        "last_used_at": "2026-01-28T10:00:00Z"
      },
      {
        "id": "payee-2",
        "name": "Marie Dubois",
        "payee_type": "international",
        "iban": "FR1420041010050500013M02606",
        "is_favorite": false,
        "last_used_at": null
      }
    ],
    "total": 2
  }
}
```

---

### Update Payee

```http
PUT /payees/{payee_id}
```

**Request Body:**
```json
{
  "name": "John Smith Jr.",
  "is_favorite": true,
  "reference": "Monthly rent"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "payee-550e8400-e29b-41d4-a716-446655440000",
    "name": "John Smith Jr.",
    "is_favorite": true,
    "reference": "Monthly rent",
    "updated_at": "2026-01-30T11:00:00Z"
  }
}
```

---

### Delete Payee

```http
DELETE /payees/{payee_id}
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Payee deleted successfully"
}
```

---

### Make Payment

```http
POST /payments
```

**Request Body:**
```json
{
  "account_id": "acc-550e8400-e29b-41d4-a716-446655440000",
  "payee_id": "payee-550e8400-e29b-41d4-a716-446655440000",
  "amount": 1200.00,
  "currency": "GBP",
  "reference": "January rent"
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "payment_id": "pay-550e8400-e29b-41d4-a716-446655440000",
    "transaction_id": "txn-550e8400-e29b-41d4-a716-446655440000",
    "amount": 1200.00,
    "currency": "GBP",
    "status": "completed",
    "payee": {
      "name": "John Smith",
      "account_number": "12345678",
      "sort_code": "040004"
    },
    "created_at": "2026-01-30T11:00:00Z"
  },
  "message": "Payment sent successfully"
}
```

---

### Create Standing Order

```http
POST /standing-orders
```

**Request Body:**
```json
{
  "account_id": "acc-550e8400-e29b-41d4-a716-446655440000",
  "payee_id": "payee-550e8400-e29b-41d4-a716-446655440000",
  "amount": 1200.00,
  "currency": "GBP",
  "reference": "Monthly rent",
  "frequency": "monthly",
  "start_date": "2026-02-01",
  "end_date": null,
  "total_payments": null
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "id": "so-550e8400-e29b-41d4-a716-446655440000",
    "account_id": "acc-550e8400-e29b-41d4-a716-446655440000",
    "payee": {
      "id": "payee-550e8400-e29b-41d4-a716-446655440000",
      "name": "John Smith"
    },
    "amount": 1200.00,
    "frequency": "monthly",
    "start_date": "2026-02-01",
    "next_payment_date": "2026-02-01",
    "status": "active",
    "created_at": "2026-01-30T11:00:00Z"
  },
  "message": "Standing order created successfully"
}
```

---

### List Standing Orders

```http
GET /standing-orders
```

**Query Parameters:**
- `status` (optional): `active`, `paused`, `completed`, `cancelled`
- `account_id` (optional): Filter by account

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "standing_orders": [
      {
        "id": "so-1",
        "payee": {
          "name": "John Smith",
          "account_number": "12345678"
        },
        "amount": 1200.00,
        "currency": "GBP",
        "frequency": "monthly",
        "next_payment_date": "2026-02-01",
        "payments_made": 12,
        "status": "active"
      }
    ],
    "total": 1
  }
}
```

---

### Cancel Standing Order

```http
DELETE /standing-orders/{standing_order_id}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "so-550e8400-e29b-41d4-a716-446655440000",
    "status": "cancelled",
    "cancelled_at": "2026-01-30T11:00:00Z"
  },
  "message": "Standing order cancelled successfully"
}
```

---

### Schedule Payment

```http
POST /scheduled-payments
```

**Request Body:**
```json
{
  "account_id": "acc-550e8400-e29b-41d4-a716-446655440000",
  "payee_id": "payee-550e8400-e29b-41d4-a716-446655440000",
  "amount": 500.00,
  "currency": "GBP",
  "reference": "Birthday gift",
  "scheduled_date": "2026-02-15"
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "id": "sp-550e8400-e29b-41d4-a716-446655440000",
    "account_id": "acc-550e8400-e29b-41d4-a716-446655440000",
    "payee": {
      "name": "Jane Doe"
    },
    "amount": 500.00,
    "scheduled_date": "2026-02-15",
    "status": "pending",
    "created_at": "2026-01-30T11:00:00Z"
  },
  "message": "Payment scheduled for 2026-02-15"
}
```

---

### List Direct Debits

```http
GET /direct-debits
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "direct_debits": [
      {
        "id": "dd-1",
        "mandate_reference": "DD-123456",
        "originator_name": "British Gas",
        "originator_id": "BG123",
        "status": "active",
        "created_date": "2025-01-15"
      },
      {
        "id": "dd-2",
        "mandate_reference": "DD-789012",
        "originator_name": "Netflix",
        "originator_id": "NF456",
        "status": "active",
        "created_date": "2025-03-20"
      }
    ],
    "total": 2
  }
}
```

---

### Cancel Direct Debit

```http
DELETE /direct-debits/{direct_debit_id}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "dd-1",
    "status": "cancelled",
    "cancelled_date": "2026-01-30"
  },
  "message": "Direct debit cancelled successfully"
}
```

---

## Notification Service

Base path: `/notifications`

### List Notifications

```http
GET /notifications
```

**Query Parameters:**
- `page` (default: 1)
- `limit` (default: 50)
- `is_read` (optional): `true`, `false`
- `type` (optional): `transaction`, `security`, `system`

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "notifications": [
      {
        "id": "notif-1",
        "type": "transaction",
        "title": "Payment sent",
        "message": "£25.50 sent to Starbucks",
        "is_read": false,
        "sent_at": "2026-01-30T11:00:00Z",
        "transaction_id": "txn-1"
      },
      {
        "id": "notif-2",
        "type": "security",
        "title": "New device login",
        "message": "Your account was accessed from a new device",
        "is_read": true,
        "sent_at": "2026-01-29T15:30:00Z",
        "read_at": "2026-01-29T15:31:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 50,
      "total_pages": 3,
      "total_records": 142
    }
  }
}
```

---

### Mark Notification as Read

```http
PUT /notifications/{notification_id}/read
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "notif-1",
    "is_read": true,
    "read_at": "2026-01-30T11:05:00Z"
  }
}
```

---

### Mark All as Read

```http
PUT /notifications/read-all
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "marked_read": 15
  },
  "message": "All notifications marked as read"
}
```

---

### Get Notification Preferences

```http
GET /notifications/preferences
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "push_enabled": true,
    "email_enabled": true,
    "sms_enabled": false,
    "preferences": {
      "transaction": {
        "push": true,
        "email": false,
        "sms": false
      },
      "security": {
        "push": true,
        "email": true,
        "sms": true
      },
      "marketing": {
        "push": false,
        "email": false,
        "sms": false
      }
    }
  }
}
```

---

### Update Notification Preferences

```http
PUT /notifications/preferences
```

**Request Body:**
```json
{
  "push_enabled": true,
  "preferences": {
    "transaction": {
      "push": true,
      "email": false
    },
    "marketing": {
      "push": false,
      "email": false
    }
  }
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Notification preferences updated"
}
```

---

## Analytics Service

Base path: `/analytics`

### Get Spending Summary

```http
GET /analytics/spending
```

**Query Parameters:**
- `account_id` (required)
- `start_date` (optional): ISO 8601 date
- `end_date` (optional): ISO 8601 date
- `period` (optional): `day`, `week`, `month`, `year`

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "period": "month",
    "start_date": "2026-01-01",
    "end_date": "2026-01-31",
    "total_spent": 1245.75,
    "total_income": 2500.00,
    "net": 1254.25,
    "transaction_count": 67,
    "average_transaction": 18.59,
    "categories": [
      {
        "category_id": "cat-eating-out",
        "name": "Eating Out",
        "icon": "🍽️",
        "color": "#EA4C89",
        "total": 325.50,
        "percentage": 26.1,
        "transaction_count": 12
      },
      {
        "category_id": "cat-groceries",
        "name": "Groceries",
        "icon": "🛒",
        "color": "#EE7B30",
        "total": 280.00,
        "percentage": 22.5,
        "transaction_count": 8
      }
    ]
  }
}
```

---

### Get Monthly Report

```http
GET /analytics/monthly-report
```

**Query Parameters:**
- `account_id` (required)
- `year` (required): e.g., 2026
- `month` (required): 1-12

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "month": "January 2026",
    "summary": {
      "total_spent": 1245.75,
      "total_income": 2500.00,
      "savings": 1254.25,
      "savings_rate": 50.17
    },
    "top_categories": [
      {
        "name": "Eating Out",
        "total": 325.50,
        "percentage": 26.1
      },
      {
        "name": "Groceries",
        "total": 280.00,
        "percentage": 22.5
      }
    ],
    "top_merchants": [
      {
        "name": "Tesco",
        "total": 150.00,
        "transaction_count": 5
      },
      {
        "name": "Starbucks",
        "total": 125.50,
        "transaction_count": 8
      }
    ],
    "daily_spending": [
      {"date": "2026-01-01", "amount": 45.20},
      {"date": "2026-01-02", "amount": 12.50}
    ]
  }
}
```

---

### Export Transactions

```http
GET /analytics/export
```

**Query Parameters:**
- `account_id` (required)
- `start_date` (required): ISO 8601 date
- `end_date` (required): ISO 8601 date
- `format` (required): `csv`, `pdf`, `json`

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "download_url": "https://api.protobank.com/downloads/transactions-jan-2026.csv",
    "expires_at": "2026-01-30T12:00:00Z",
    "format": "csv",
    "file_size": 45678
  }
}
```

---

## Webhooks (Optional)

Base path: `/webhooks`

### Subscribe to Webhook

```http
POST /webhooks
```

**Request Body:**
```json
{
  "url": "https://your-app.com/webhook",
  "events": ["transaction.created", "payment.executed"],
  "secret": "your-webhook-secret"
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "id": "webhook-550e8400-e29b-41d4-a716-446655440000",
    "url": "https://your-app.com/webhook",
    "events": ["transaction.created", "payment.executed"],
    "status": "active",
    "created_at": "2026-01-30T11:00:00Z"
  }
}
```

### Webhook Payload Example

When an event occurs, we'll send a POST request to your webhook URL:

```json
{
  "event": "transaction.created",
  "timestamp": "2026-01-30T11:00:00Z",
  "data": {
    "transaction_id": "txn-550e8400-e29b-41d4-a716-446655440000",
    "account_id": "acc-550e8400-e29b-41d4-a716-446655440000",
    "amount": -25.50,
    "description": "Coffee at Starbucks",
    "status": "completed"
  }
}
```

---

## Postman Collection

Import this URL into Postman to get started quickly:
```
https://api.protobank.example.com/api/v1/postman-collection
```

---

## Support

For API support and questions:
- Documentation: https://docs.protobank.example.com
- Email: api-support@protobank.example.com
- Status: https://status.protobank.example.com
