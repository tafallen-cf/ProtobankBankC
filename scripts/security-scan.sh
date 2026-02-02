#!/bin/bash

# Security Scanning Script for Local Development
# Run this before pushing code to catch security issues early

set -e

echo "🔒 Running Security Scans for ProtobankBankC"
echo "=============================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo -e "${RED}❌ Error: Must be run from a Go module directory${NC}"
    echo "   cd backend/auth-service first"
    exit 1
fi

echo "1️⃣  Installing security tools..."
echo "-----------------------------------"

# Install gosec if not present
if ! command -v gosec &> /dev/null; then
    echo "Installing gosec..."
    go install github.com/securego/gosec/v2/cmd/gosec@latest
fi

# Install govulncheck if not present
if ! command -v govulncheck &> /dev/null; then
    echo "Installing govulncheck..."
    go install golang.org/x/vuln/cmd/govulncheck@latest
fi

# Install staticcheck if not present
if ! command -v staticcheck &> /dev/null; then
    echo "Installing staticcheck..."
    go install honnef.co/go/tools/cmd/staticcheck@latest
fi

echo -e "${GREEN}✅ Tools installed${NC}"
echo ""

echo "2️⃣  Running gosec (Security Scanner)..."
echo "-----------------------------------"
gosec -fmt=json -out=gosec-report.json ./... 2>/dev/null || true
gosec ./...
GOSEC_EXIT=$?

if [ $GOSEC_EXIT -eq 0 ]; then
    echo -e "${GREEN}✅ No security issues found by gosec${NC}"
else
    echo -e "${YELLOW}⚠️  Security issues found - check output above${NC}"
fi
echo ""

echo "3️⃣  Running govulncheck (Vulnerability Scanner)..."
echo "-----------------------------------"
govulncheck ./...
GOVULN_EXIT=$?

if [ $GOVULN_EXIT -eq 0 ]; then
    echo -e "${GREEN}✅ No known vulnerabilities found${NC}"
else
    echo -e "${RED}❌ Vulnerabilities found - please update dependencies${NC}"
fi
echo ""

echo "4️⃣  Running staticcheck (Code Quality)..."
echo "-----------------------------------"
staticcheck ./...
STATIC_EXIT=$?

if [ $STATIC_EXIT -eq 0 ]; then
    echo -e "${GREEN}✅ No issues found by staticcheck${NC}"
else
    echo -e "${YELLOW}⚠️  Code quality issues found${NC}"
fi
echo ""

echo "5️⃣  Checking for sensitive data..."
echo "-----------------------------------"
# Check for potential secrets in code
SECRETS_FOUND=0

# Check for hardcoded passwords
if grep -r "password.*=.*\"" --include="*.go" . | grep -v "_test.go" | grep -v "Password string" > /dev/null 2>&1; then
    echo -e "${RED}⚠️  Potential hardcoded passwords found${NC}"
    grep -r "password.*=.*\"" --include="*.go" . | grep -v "_test.go" | grep -v "Password string" || true
    SECRETS_FOUND=1
fi

# Check for hardcoded API keys
if grep -r "api[_-]key.*=.*\"" --include="*.go" . | grep -v "_test.go" > /dev/null 2>&1; then
    echo -e "${RED}⚠️  Potential hardcoded API keys found${NC}"
    grep -r "api[_-]key.*=.*\"" --include="*.go" . | grep -v "_test.go" || true
    SECRETS_FOUND=1
fi

# Check for TODO SECURITY comments
if grep -r "TODO.*SECURITY" --include="*.go" . > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  Security TODOs found:${NC}"
    grep -r "TODO.*SECURITY" --include="*.go" . || true
fi

if [ $SECRETS_FOUND -eq 0 ]; then
    echo -e "${GREEN}✅ No obvious secrets found in code${NC}"
fi
echo ""

echo "6️⃣  Checking dependencies..."
echo "-----------------------------------"
# Check for outdated dependencies
go list -m -u all | grep "\[" || echo -e "${GREEN}✅ All dependencies up to date${NC}"
echo ""

# Summary
echo "=============================================="
echo "📊 Security Scan Summary"
echo "=============================================="

FAILED=0
if [ $GOSEC_EXIT -ne 0 ]; then
    echo -e "${YELLOW}⚠️  gosec: Issues found${NC}"
    FAILED=$((FAILED + 1))
else
    echo -e "${GREEN}✅ gosec: Clean${NC}"
fi

if [ $GOVULN_EXIT -ne 0 ]; then
    echo -e "${RED}❌ govulncheck: Vulnerabilities found${NC}"
    FAILED=$((FAILED + 1))
else
    echo -e "${GREEN}✅ govulncheck: Clean${NC}"
fi

if [ $STATIC_EXIT -ne 0 ]; then
    echo -e "${YELLOW}⚠️  staticcheck: Issues found${NC}"
    FAILED=$((FAILED + 1))
else
    echo -e "${GREEN}✅ staticcheck: Clean${NC}"
fi

if [ $SECRETS_FOUND -ne 0 ]; then
    echo -e "${RED}❌ Secrets: Potential secrets found${NC}"
    FAILED=$((FAILED + 1))
else
    echo -e "${GREEN}✅ Secrets: Clean${NC}"
fi

echo ""
echo "Reports generated:"
echo "  - gosec-report.json"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}🎉 All security checks passed!${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠️  $FAILED check(s) found issues${NC}"
    echo "   Review the output above and fix issues before committing"
    exit 1
fi
