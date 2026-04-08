---
description: setup golang
---

================================================================================
  GO CODE QUALITY SETUP — WORKFLOW
  Spendly Backend / Go Projects
  Generated for: VanPhoebe
================================================================================

OVERVIEW
--------
4 layer quality gate yang akan lo setup:
  [1] Static Analysis   → go vet, staticcheck, shadow
  [2] Linting           → golangci-lint (.golangci.yml)
  [3] Testing           → race detector, coverage, pprof, benchmarks
  [4] Automation        → Makefile + pre-commit hook + GitHub Actions

Estimasi waktu setup: ~20-30 menit
Semua tools gratis & open source.

================================================================================
STEP 1 — PREREQUISITES
================================================================================

Pastikan Go sudah terinstall:
  go version
  → minimal Go 1.21+

Pastikan $GOPATH/bin ada di PATH:
  export PATH=$PATH:$(go env GOPATH)/bin

  Tambahkan ke ~/.zshrc atau ~/.bashrc supaya permanen:
  echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
  source ~/.zshrc

================================================================================
STEP 2 — INSTALL TOOLS (Layer 1 & 2)
================================================================================

  go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest
  go install honnef.co/go/tools/cmd/staticcheck@latest
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

Verifikasi:
  shadow --help
  staticcheck --version
  golangci-lint --version

================================================================================
STEP 3 — GOLANGCI-LINT CONFIG
================================================================================

Buat file .golangci.yml di ROOT project lo (sejajar go.mod):

-------- COPY MULAI SINI --------

run:
  timeout: 5m
  tests: true

linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - unused
    - gosec
    - revive
    - gocyclo
    - dupl
    - gocritic
    - exhaustive
    - bodyclose
    - contextcheck
    - noctx
    - sqlclosecheck
    - unparam

linters-settings:
  gocyclo:
    min-complexity: 15
  dupl:
    threshold: 100
  gosec:
    excludes:
      - G104
  revive:
    rules:
      - name: exported
        disabled: false
      - name: var-naming
        disabled: false
      - name: error-return
        disabled: false

issues:
  max-same-issues: 3
  exclude-rules:
    - path: "_test.go"
      linters:
        - dupl
        - gosec

-------- COPY SELESAI SINI --------

Test lint berjalan:
  golangci-lint run ./...

================================================================================
STEP 4 — MAKEFILE
================================================================================

Buat file Makefile di ROOT project (sejajar go.mod):

-------- COPY MULAI SINI --------

.PHONY: check lint test coverage bench clean

# Jalankan semua: lint + test
check: lint test

# Static analysis + linting
lint:
	go vet ./...
	staticcheck ./...
	golangci-lint run ./...

# Test dengan race detector (wajib sebelum push)
test:
	go test -race -count=1 -timeout=30s ./...

# Test + coverage report
coverage:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Lihat summary coverage saja
coverage-summary:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | grep "total:"

# Benchmark semua
bench:
	go test -bench=. -benchmem -count=3 ./...

# CPU profiling
profile-cpu:
	go test -cpuprofile=cpu.prof -bench=. ./...
	go tool pprof cpu.prof

# Memory profiling
profile-mem:
	go test -memprofile=mem.prof -bench=. ./...
	go tool pprof mem.prof

# Cleanup generated files
clean:
	rm -f coverage.out coverage.html cpu.prof mem.prof

-------- COPY SELESAI SINI --------

CATATAN: Indent di Makefile WAJIB menggunakan TAB, bukan spasi.

Test Makefile:
  make lint
  make test
  make coverage

================================================================================
STEP 5 — PRE-COMMIT HOOK
================================================================================

Buat file executable di .git/hooks/pre-commit:

  nano .git/hooks/pre-commit

Isi file:

-------- COPY MULAI SINI --------

#!/bin/sh
# Go Quality Gate — pre-commit hook
# Otomatis berjalan setiap kali `git commit`

echo "==> Running quality checks before commit..."

echo "[1/3] go vet..."
go vet ./...
if [ $? -ne 0 ]; then
  echo "FAILED: go vet menemukan masalah. Commit dibatalkan."
  exit 1
fi

echo "[2/3] golangci-lint..."
golangci-lint run ./...
if [ $? -ne 0 ]; then
  echo "FAILED: Lint error ditemukan. Commit dibatalkan."
  exit 1
fi

echo "[3/3] go test -race..."
go test -race -count=1 -timeout=60s ./...
if [ $? -ne 0 ]; then
  echo "FAILED: Test gagal. Commit dibatalkan."
  exit 1
fi

echo "==> Semua checks passed. Commit dilanjutkan."

-------- COPY SELESAI SINI --------

Jadikan executable:
  chmod +x .git/hooks/pre-commit

Test hook:
  git add .
  git commit -m "test"
  → Seharusnya menjalankan semua checks dulu sebelum commit terjadi

Bypass hook kalau darurat (jangan dibiasain):
  git commit --no-verify -m "emergency fix"

================================================================================
STEP 6 — GITHUB ACTIONS CI/CD
================================================================================

Buat folder dan file:
  mkdir -p .github/workflows
  nano .github/workflows/quality.yml

Isi file:

-------- COPY MULAI SINI --------

name: Quality Gate

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  quality:
    name: Code Quality Check
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          cache: true

      - name: Download dependencies
        run: go mod download

      - name: Static analysis
        run: |
          go vet ./...
          go install honnef.co/go/tools/cmd/staticcheck@latest
          staticcheck ./...

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest
          args: --timeout=5m

      - name: Run tests with race detector
        run: go test -race -count=1 -timeout=60s ./...

      - name: Check coverage threshold
        run: |
          go test -coverprofile=coverage.out -covermode=atomic ./...
          COVERAGE=$(go tool cover -func=coverage.out | grep "total:" | awk '{print $3}' | tr -d '%')
          echo "Total coverage: ${COVERAGE}%"
          if [ $(echo "$COVERAGE < 60" | bc -l) -eq 1 ]; then
            echo "ERROR: Coverage ${COVERAGE}% di bawah threshold 60%"
            exit 1
          fi
          echo "Coverage OK: ${COVERAGE}%"

      - name: Upload coverage report
        uses: actions/upload-artifact@v4
        if: always()
        with:
          name: coverage-report
          path: coverage.out

-------- COPY SELESAI SINI --------

Commit dan push:
  git add .github/
  git commit -m "ci: add quality gate workflow"
  git push

Cek di: github.com/{username}/{repo}/actions

================================================================================
STEP 7 — CURSOR RULE (opsional tapi recommended)
================================================================================

Buat file .cursor/rules/quality-go.mdc:

-------- COPY MULAI SINI --------

---
description: Go quality enforcement. Auto-attach saat edit file Go.
globs: ["**/*.go"]
---

# Quality Contracts
- Every exported function must have a godoc comment
- Every function returning error must be tested for the error path
- Coverage target: minimum 70% per package
- Cyclomatic complexity max 15 per function
- Benchmarks required for any function in hot path (repository layer, parsers)

# Forbidden Patterns
- _ = someFunc() → always handle errors explicitly
- interface{} or any without justification → use typed interfaces
- init() with side effects → move to explicit constructor
- Global mutable var → inject via struct fields
- goroutine without WaitGroup/errgroup/context cancel

# Test Patterns
- Table-driven tests: var tests = []struct{ name, input, want }{}
- b.ResetTimer() after setup in benchmarks
- t.Parallel() on independent test cases
- testify/assert for cleaner assertions

# Before Suggesting Code
- Check if the function being written has an existing test file
- If no test exists, suggest creating one alongside the implementation
- Flag any error return that is not being checked by the caller

-------- COPY SELESAI SINI --------

================================================================================
STEP 8 — VERIFIKASI AKHIR
================================================================================

Jalankan ini untuk pastikan semua berjalan:

  # 1. Lint
  make lint

  # 2. Test
  make test

  # 3. Coverage (buka coverage.html di browser)
  make coverage

  # 4. Simulasi commit (trigger pre-commit hook)
  git add .
  git commit -m "chore: add quality tooling"

Kalau semua hijau → setup selesai.

================================================================================
CHEATSHEET HARIAN
================================================================================

  make check          → lint + test (jalankan sebelum push)
  make coverage       → lihat coverage per fungsi + HTML report
  make bench          → benchmark semua
  make profile-cpu    → profiling CPU (buka di pprof interactive)
  golangci-lint run   → lint manual kapan saja
  go test -race ./... → test dengan race detector

================================================================================
STRUKTUR FILE YANG DIHASILKAN
================================================================================

project-root/
├── .cursor/
│   └── rules/
│       └── quality-go.mdc         ← Cursor AI rule
├── .github/
│   └── workflows/
│       └── quality.yml            ← GitHub Actions CI
├── .git/
│   └── hooks/
│       └── pre-commit             ← Auto-run sebelum commit
├── .golangci.yml                  ← Linter config
├── Makefile                       ← Command shortcuts
└── [go source files...]

================================================================================
  END OF WORKFLOW
================================================================================