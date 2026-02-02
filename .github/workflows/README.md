# CI/CD Workflows

Automated workflows for continuous integration, security scanning, and deployment.

## Workflows

### 1. CI Workflow (`ci.yml`)

**Triggers**: Push to main/develop, Pull Requests

**Jobs**:
- **Test**: Run all tests with PostgreSQL and Redis services
  - Unit tests with race detector
  - Coverage reporting (80% threshold)
  - Upload to Codecov
  - Generate HTML coverage reports

- **Lint**: Code quality checks
  - golangci-lint with comprehensive linters
  - 30+ linters enabled

- **Security**: Security scanning
  - gosec for security vulnerabilities
  - govulncheck for known vulnerabilities
  - SARIF reports uploaded to GitHub Security

- **Build**: Binary compilation
  - Static binary build
  - Cross-platform compilation
  - Artifact upload

**Required Secrets**: None (uses GitHub token)

**Status Badge**:
```markdown
![CI](https://github.com/tafallen-cf/ProtobankBankC/workflows/CI/badge.svg)
```

### 2. Docker Build Workflow (`docker.yml`)

**Triggers**: Push to main, version tags, Pull Requests

**Jobs**:
- **Build**: Container image build and push
  - Multi-stage Docker build
  - Push to GitHub Container Registry (ghcr.io)
  - Multiple image tags (latest, SHA, version)
  - Layer caching for faster builds
  - Trivy vulnerability scanning
  - SBOM (Software Bill of Materials) generation

**Image Tags**:
- `latest` - Latest main branch build
- `main-{sha}` - Specific commit on main
- `v1.0.0` - Semantic version tags
- `pr-123` - Pull request builds

**Registry**: `ghcr.io/tafallen-cf/protobankbankc/auth-service`

**Required Secrets**:
- Automatic via `GITHUB_TOKEN` (packages permission required)

**Status Badge**:
```markdown
![Docker](https://github.com/tafallen-cf/ProtobankBankC/workflows/Docker%20Build/badge.svg)
```

### 3. CodeQL Workflow (`codeql.yml`)

**Triggers**: Push, Pull Requests, Weekly schedule (Mondays 6 AM UTC)

**Jobs**:
- **Analyze**: Advanced security analysis
  - CodeQL semantic code analysis
  - Security-extended query suite
  - Security and quality queries
  - Automated vulnerability detection

**Languages**: Go

**Status Badge**:
```markdown
![CodeQL](https://github.com/tafallen-cf/ProtobankBankC/workflows/CodeQL/badge.svg)
```

## Dependabot

Automated dependency updates configured in `.github/dependabot.yml`:

- **Go modules**: Weekly updates (Mondays 9 AM)
- **GitHub Actions**: Weekly updates
- **Docker base images**: Weekly updates

Pull requests are automatically created for:
- Security updates
- Version updates
- Dependency updates

## Configuration Files

### `.golangci.yml`

Linting configuration with 30+ linters enabled:
- Code quality checks
- Security analysis
- Style enforcement
- Best practices
- Performance optimization

### Coverage Requirements

- **Minimum**: 80% code coverage
- **Target**: 90% code coverage
- **CI Failure**: < 80% coverage

## Badges

Add these badges to your README:

```markdown
![CI](https://github.com/tafallen-cf/ProtobankBankC/workflows/CI/badge.svg)
![Docker](https://github.com/tafallen-cf/ProtobankBankC/workflows/Docker%20Build/badge.svg)
![CodeQL](https://github.com/tafallen-cf/ProtobankBankC/workflows/CodeQL/badge.svg)
[![codecov](https://codecov.io/gh/tafallen-cf/ProtobankBankC/branch/main/graph/badge.svg)](https://codecov.io/gh/tafallen-cf/ProtobankBankC)
```

## Secrets Required

### Repository Secrets

Currently, all workflows use `GITHUB_TOKEN` which is automatically provided.

### Optional Secrets (for future expansion)

- `CODECOV_TOKEN`: For private repositories on Codecov
- `SLACK_WEBHOOK`: For Slack notifications
- `SONAR_TOKEN`: For SonarQube analysis
- `AWS_ACCESS_KEY_ID`: For AWS deployments
- `AWS_SECRET_ACCESS_KEY`: For AWS deployments

## Adding New Services

When adding new microservices, duplicate and modify workflows:

1. Copy `ci.yml` and adjust paths
2. Copy `docker.yml` and update image names
3. Update Dependabot configuration
4. Add service-specific linting rules

## Troubleshooting

### Tests Failing

```bash
# Run tests locally with the same environment
docker-compose up -d postgres redis
cd backend/auth-service
go test -v -race -coverprofile=coverage.out ./...
```

### Coverage Below Threshold

```bash
# Check coverage locally
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Linting Failures

```bash
# Run linter locally
golangci-lint run --timeout=5m
```

### Docker Build Failures

```bash
# Build locally
cd backend/auth-service
docker build -t auth-service:test .
docker run --rm auth-service:test
```

## Performance

- **Average CI time**: ~5 minutes
- **Docker build time**: ~3 minutes (with cache)
- **Full security scan**: ~2 minutes

## Best Practices

1. **All PRs must pass CI** before merge
2. **Coverage must be >= 80%**
3. **No critical security issues**
4. **All linters must pass**
5. **Docker images must scan clean**
6. **SBOM generated for all releases**

## Future Enhancements

- [ ] E2E testing workflow
- [ ] Performance testing
- [ ] Automated rollback on failures
- [ ] Multi-environment deployments
- [ ] Release automation
- [ ] Changelog generation
