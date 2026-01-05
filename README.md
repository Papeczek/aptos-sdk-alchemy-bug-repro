# Aptos Go SDK × Alchemy RPC Repro

This repository demonstrates a compatibility issue between
`github.com/aptos-labs/aptos-go-sdk` and Alchemy’s Aptos RPC.

Specifically, the SDK method `Client.AccountAPTBalance()` fails when used
against Alchemy’s Aptos RPC, while the same call succeeds against other
providers.


---

## What works

Using `Client.AccountAPTBalance()` against:
- Aptos Labs public fullnode
- QuickNode rpc

---

## What fails

Using the same SDK call against:
- Alchemy Aptos RPC

Observed error include:
- `Parse error`

---

## How to reproduce

### 1. Install dependencies
- Go 1.22+
- `go` and `git` installed

## Run

`go run .`
