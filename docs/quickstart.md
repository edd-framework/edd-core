# Quickstart — Your First Contract in 3 Minutes

## 1. Install

```bash
# Homebrew
brew install edd-framework/tap/edd

# Or Go
$ go install github.com/edd-framework/edd-core/cmd/edd@latest

# Or direct download
$ curl -sSL https://edd-framework.github.io/edd-core/install.sh | bash
```

## 2. Initialize EDD in your repo

```bash
$ cd your-project
$ edd init
```

This creates `.edd/contracts/`, `.edd/policy.yaml`, and `.edd/evidence/`.

## 3. Create your first contract

```bash
$ edd new "Add health check endpoint" --profile express
```

## 4. Fill in the contract

Open `.edd/contracts/add-health-check-endpoint.yaml` and replace every TODO.

## 5. Verify

```bash
$ edd verify .edd/contracts/add-health-check-endpoint.yaml
```

## 6. Check status

```bash
$ edd status .edd/contracts/add-health-check-endpoint.yaml
```

Shows claims, gates, and the computed merge/close decision.
