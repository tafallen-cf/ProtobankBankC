module github.com/protobankbankc/auth-service

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/google/uuid v1.5.0
	github.com/jackc/pgx/v5 v5.5.1
	github.com/redis/go-redis/v9 v9.3.1
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/viper v1.18.2
	github.com/stretchr/testify v1.8.4
	golang.org/x/crypto v0.18.0
)

require (
	// Testing dependencies
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/go-redis/redismock/v9 v9.2.0
	github.com/golang/mock v1.6.0
	github.com/testcontainers/testcontainers-go v0.27.0
)

require (
	// Security scanning
	github.com/securego/gosec/v2 v2.18.2
)
