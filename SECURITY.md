# Security Policy

## GitHub Code Scanning Status

### ✅ **Repository is Public - Full Code Scanning Enabled**

This repository is **public** and has **full GitHub code scanning** enabled for free, including:

- ✅ **CodeQL** - Advanced semantic code analysis
- ✅ **Secret scanning** - Automatic credential detection
- ✅ **Dependency graph** - Dependency visualization
- ✅ **Dependabot alerts** - Vulnerability notifications
- ✅ **Security advisories** - CVE tracking

The CodeQL workflow (`.github/workflows/codeql.yml`) runs automatically:
- On every push to `main` branch
- On all pull requests
- Weekly on Monday at 6:00 AM UTC

### Current Security Scanning Setup

#### ✅ **Advanced Code Scanning** (Enabled)

1. **CodeQL** - Semantic code analysis
   - Deep semantic analysis of code
   - Security and quality queries
   - Detects complex vulnerabilities
   - Results in GitHub Security tab

2. **Secret scanning** - Credential protection
   - Automatically detects exposed secrets
   - Blocks commits with credentials
   - Alerts on token exposure

#### ✅ **CI/CD Security Scans** (Active)

These run on every push/PR automatically:

1. **gosec** - Static security analysis
   - Detects common security issues
   - Checks for SQL injection, XSS, etc.
   - Results uploaded to GitHub Security tab (SARIF format)

2. **govulncheck** - Vulnerability database
   - Checks against Go vulnerability database
   - Identifies known CVEs in dependencies

3. **golangci-lint** - Code quality
   - 30+ linters including security checks
   - gosec integration
   - Best practices enforcement

4. **Trivy** - Container scanning
   - Scans Docker images for vulnerabilities
   - Checks OS packages and application dependencies
   - Results uploaded to GitHub Security tab

5. **Dependabot** - Dependency updates
   - Automated security updates
   - Weekly dependency checks
   - Pull requests for outdated packages

#### ✅ **Local Security Scanning** (Optional)

We also provide a comprehensive local security scanning script:

```bash
cd backend/auth-service
../../scripts/security-scan.sh
```

This script runs:
- ✅ **gosec** - Security vulnerability scanner
- ✅ **govulncheck** - Known vulnerability checker
- ✅ **staticcheck** - Code quality analyzer
- ✅ **Secret detection** - Hardcoded credentials finder
- ✅ **Dependency auditing** - Outdated package checker

## Security Scanning Workflow

### Local Development

**Before committing:**

```bash
# Run security scan
cd backend/auth-service
../../scripts/security-scan.sh

# Run tests with coverage
go test -race -coverprofile=coverage.out ./...

# Check coverage
go tool cover -func=coverage.out
```

### Continuous Integration

**Automatic on push/PR:**

1. ✅ All tests run with PostgreSQL + Redis
2. ✅ Security scans (gosec, govulncheck)
3. ✅ Code quality checks (golangci-lint)
4. ✅ Coverage threshold check (80%)
5. ✅ Container vulnerability scan (Trivy)

### Manual Security Review

**Run comprehensive security audit:**

```bash
# Go to service directory
cd backend/auth-service

# Install tools
go install github.com/securego/gosec/v2/cmd/gosec@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
go install honnef.co/go/tools/cmd/staticcheck@latest

# Run gosec
gosec -fmt=json -out=gosec-report.json ./...

# Run govulncheck
govulncheck ./...

# Run staticcheck
staticcheck ./...

# Check for secrets
git secrets --scan

# Audit dependencies
go list -m -u all
go mod why -m <module>
```

## Security Best Practices

### Code Security

- ✅ All passwords hashed with bcrypt (cost factor 12)
- ✅ JWT tokens with expiration (15min access, 7day refresh)
- ✅ Input validation on all endpoints
- ✅ Parameterized SQL queries (no SQL injection)
- ✅ Rate limiting (10 requests/minute per IP)
- ✅ CORS properly configured
- ✅ No secrets in code (environment variables)

### Container Security

- ✅ Non-root user (UID 1000)
- ✅ Read-only root filesystem
- ✅ All capabilities dropped
- ✅ Minimal base image (Alpine)
- ✅ Multi-stage builds
- ✅ Regular base image updates (Dependabot)

### Infrastructure Security

- ✅ TLS 1.3 only
- ✅ Secrets in Kubernetes Secrets
- ✅ Pod security policies
- ✅ Network policies
- ✅ Resource limits
- ✅ Health checks

## Reporting Security Vulnerabilities

### Reporting Process

**Do NOT report security vulnerabilities through public GitHub issues.**

Instead:

1. **Email**: security@protobankbankc.example.com
2. **Include**:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if available)

3. **Response Time**: We aim to respond within 48 hours

### Disclosure Policy

- We will acknowledge receipt within 48 hours
- We will provide a detailed response within 7 days
- We will work on a fix and coordinate disclosure timing
- We credit researchers who report responsibly

## Security Checklist

### Before Deployment

- [ ] All tests passing (80%+ coverage)
- [ ] Security scans clean (gosec, govulncheck)
- [ ] No high/critical Trivy vulnerabilities
- [ ] Dependencies up to date
- [ ] Secrets properly managed (not in code)
- [ ] Rate limiting configured
- [ ] Logging enabled
- [ ] Monitoring configured
- [ ] Backup strategy in place
- [ ] Incident response plan ready

### Monthly Security Review

- [ ] Run full security scan
- [ ] Review dependencies for updates
- [ ] Check for new CVEs
- [ ] Review access logs
- [ ] Test incident response
- [ ] Update security documentation

## Security Tools

### Installed in CI/CD

- **gosec** v2.18+ - Go security checker
- **govulncheck** - Go vulnerability database
- **golangci-lint** v1.55+ - Meta-linter with security checks
- **Trivy** - Container vulnerability scanner
- **Dependabot** - Automated dependency updates

### Recommended Local Tools

```bash
# Install security tools
go install github.com/securego/gosec/v2/cmd/gosec@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
go install honnef.co/go/tools/cmd/staticcheck@latest

# Install git-secrets
brew install git-secrets  # macOS
apt install git-secrets    # Ubuntu

# Configure git-secrets
git secrets --install
git secrets --register-aws
```

## Security Resources

### Documentation

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go Security Best Practices](https://go.dev/security/)
- [CWE Top 25](https://cwe.mitre.org/top25/)
- [NIST Cybersecurity Framework](https://www.nist.gov/cyberframework)

### Tools

- [Go Vulnerability Database](https://vuln.go.dev/)
- [GitHub Advisory Database](https://github.com/advisories)
- [Snyk Vulnerability DB](https://snyk.io/vuln/)

## Compliance

This project aims to comply with:

- ✅ **OWASP Top 10** - Web application security
- ✅ **PCI-DSS** - Payment card industry standards
- ✅ **GDPR** - Data protection (EU)
- 🔄 **SOC 2** - Security controls (in progress)
- 🔄 **ISO 27001** - Information security (in progress)

## Contact

- **Security Team**: security@protobankbankc.example.com
- **General Issues**: https://github.com/tafallen-cf/ProtobankBankC/issues
- **Security Policy**: This document

---

**Last Updated**: February 3, 2026
**Next Review**: March 3, 2026
