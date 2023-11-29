# paypal

A PayPal Golang SDK.

[![License](https://img.shields.io/github/license/adobaai/paypal)](./LICENSE)
[![GitHubRelease](https://img.shields.io/github/release/adobaai/paypal)](https://github.com/adobaai/paypal/releases)
[![BuildWorkflow](https://github.com/adobaai/paypal/actions/workflows/go.yml/badge.svg)](https://github.com/adobaai/paypal/actions)
[![GoVersion](https://img.shields.io/github/go-mod/go-version/adobaai/paypal)](./go.mod)
[![GoDoc](https://pkg.go.dev/badge/google.golang.org/grpc)](https://pkg.go.dev/github.com/adobaai/paypal)
[![GoReportCard](https://goreportcard.com/badge/adobaai/paypal)](https://goreportcard.com/report/github.com/adobaai/paypal)
[![CodeFactor](https://www.codefactor.io/repository/github/adobaai/paypal/badge)](https://www.codefactor.io/repository/github/adobaai/paypal)
[![Coverage Status](https://codecov.io/gh/adobaai/paypal/branch/main/graph/badge.svg)](https://codecov.io/gh/adobaai/paypal/branch/main)
[![Contributors](https://img.shields.io/github/contributors/adobaai/paypal)](https://github.com/adobaai/paypal/graphs/contributors)

## Usage

```go
package main

import (
	"context"
	"fmt"

	"github.com/adobaai/paypal"
)

func main() {
	ctx := context.Background()
	pc := paypal.NewClient(
		"https://api-m.sandbox.paypal.com",
		os.Getenv("PAYPAL_ID"),
		os.Getenv("PAYPAL_SECRET"),
	)
	order, err := pc.CreateOrder(ctx, &paypal.CreateOrderReq{
		Order: &paypal.Order{
			Intent: paypal.OICapture,
			PurchaseUnits: []*paypal.PurchaseUnit{
				{
					Amount: &paypal.Amount{
						CurrencyCode: "USD",
						Value:        "8.88",
					},
					Description: "Hello order",
				},
			},
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(order)
}
```

## TODO

- [ ] Codecov
- [x] Dependabot
- [x] Unit test in GitHub Action
- [x] Enable secret scanning
- [x] Enable CodeQL

## Other coverage tools

- https://github.com/marketplace/actions/go-coverage-report
- https://coveralls.io/
