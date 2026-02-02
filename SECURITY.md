# Security Policy

## Enabling GitHub Code Scanning

### For Private Repositories

GitHub Code Scanning with CodeQL requires **GitHub Advanced Security**, which is available:

1. **GitHub Enterprise Cloud** with Advanced Security
2. **GitHub Enterprise Server** 3.0+
3. **Public repositories** (free)

#### Option 1: Enable Advanced Security (Enterprise/Paid Plans)

If you have GitHub Advanced Security available:

1. Go to repository **Settings** → **Code security and analysis**
2. Click **Enable** for:
   - **Dependency graph**
   - **Dependabot alerts**
   - **Dependabot security updates**
   - **Code scanning** (requires Advanced Security)
   - **Secret scanning** (requires Advanced Security)

3. The CodeQL workflow (`.github/workflows/codeql.yml`) will automatically start running

#### Option 2: Make Repository Public

If you want free code scanning:

```bash
gh repo edit tafallen-cf/ProtobankBankC --visibility public
```

**⚠️ Warning**: This makes all code publicly visible. Only do this for open-source projects.

#### Option 3: Use Local Security Scanning

We provide a comprehensive local security scanning script that works without GitHub Advanced Security:

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

### Current Security Scanning Setup

Even without GitHub Advanced Security enabled, our repository includes:

#### ✅ **CI/CD Security Scans** (Already Working)

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

#### 🔄 **To Be Enabled** (Requires Advanced Security)

- **CodeQL** - Advanced semantic analysis
- **Secret scanning** - Automatic credential detection
- **Code scanning alerts** - GitHub Security tab integration

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

**Last Updated**: February 2, 2026
**Next Review**: March 2, 2026
