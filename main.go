package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aptos-labs/aptos-go-sdk"
	"github.com/joho/godotenv"
)

type row struct {
	Name    string
	BaseURL string
	Latency time.Duration
	Octas   uint64
	APT     float64
	Err     error
}

func env(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

// Normalize to what Aptos SDK expects as NodeUrl: base ending in /v1 (NOT /v1/view)
func normalizeNodeURL(u string) string {
	u = strings.TrimSpace(u)
	u = strings.TrimRight(u, "/")
	// If someone accidentally pastes /view, strip it.
	if strings.HasSuffix(u, "/v1/view") {
		u = strings.TrimSuffix(u, "/view")
	} else if strings.HasSuffix(u, "/view") {
		u = strings.TrimSuffix(u, "/view")
	}
	return u
}

func shortURL(u string) string {
	u = strings.TrimSpace(u)
	if u == "" {
		return ""
	}
	if len(u) <= 70 {
		return u
	}
	return u[:67] + "..."
}

func checkOne(name, rpcBaseV1, accountStr string) row {
	r := row{Name: name, BaseURL: normalizeNodeURL(rpcBaseV1)}

	if r.BaseURL == "" {
		r.Err = fmt.Errorf("missing RPC URL (set APTOS_RPC_URL_%s)", strings.TrimPrefix(name, "RPC_"))
		return r
	}

	// SDK client config
	cfg := aptos.NetworkConfig{
		NodeUrl: r.BaseURL,
		// ChainId = 0 => SDK will fetch from chain when needed (fine for this test)
	}

	client, err := aptos.NewClient(cfg)
	if err != nil {
		r.Err = fmt.Errorf("NewClient: %w", err)
		return r
	}

	// Parse account
	var addr aptos.AccountAddress
	if err := addr.ParseStringRelaxed(accountStr); err != nil {
		r.Err = fmt.Errorf("invalid APTOS_ACCOUNT: %w", err)
		return r
	}

	start := time.Now()
	octas, err := client.AccountAPTBalance(addr)
	r.Latency = time.Since(start)

	if err != nil {
		// This is where you’ll see Alchemy errors reproduced
		r.Err = err
		return r
	}

	r.Octas = octas
	r.APT = float64(octas) / 100_000_000.0
	return r
}

func main() {
	_ = godotenv.Load()

	account := env("APTOS_ACCOUNT")
	if account == "" {
		fmt.Fprintln(os.Stderr, "APTOS_ACCOUNT is required")
		os.Exit(1)
	}

	// Read 3 URLs
	rpc1 := env("APTOS_RPC_URL_1")
	rpc2 := env("APTOS_RPC_URL_2")
	rpc3 := env("APTOS_RPC_URL_3")

	// (Optional) backward-compat: allow APTOS_RPC_URL as #1
	if rpc1 == "" {
		rpc1 = env("APTOS_RPC_URL")
	}

	targets := []struct {
		Name string
		URL  string
	}{
		{"RPC_1", rpc1},
		{"RPC_2", rpc2},
		{"RPC_3", rpc3},
	}

	fmt.Println("Aptos APT Balance (via aptos-go-sdk Client.AccountAPTBalance)")
	fmt.Println("--------------------------------------------------------------------------")
	fmt.Printf("Account: %s\n\n", account)
	fmt.Printf("%-6s  %-10s  %-14s  %-12s  %s\n", "NAME", "LATENCY", "OCTAS", "APT", "RPC (base /v1)")
	fmt.Printf("%-6s  %-10s  %-14s  %-12s  %s\n", "------", "----------", "--------------", "------------", "------------------------------")

	for _, t := range targets {
		res := checkOne(t.Name, t.URL, account)
		if res.Err != nil {
			fmt.Printf("%-6s  %-10s  %-14s  %-12s  %s\n", res.Name, "-", "ERROR", "-", shortURL(res.BaseURL))
			fmt.Printf("        ↳ %v\n", res.Err)
			continue
		}
		fmt.Printf("%-6s  %-10s  %-14d  %-12.8f  %s\n",
			res.Name,
			res.Latency.Round(time.Millisecond),
			res.Octas,
			res.APT,
			shortURL(res.BaseURL),
		)
	}

}
