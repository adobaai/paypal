package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

const docBaseURL = "https://developer.paypal.com"
const eventNamesPage = docBaseURL + "/api/rest/webhooks/event-names/"
const UA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36 Edg/118.0.2088.69"

var webhookEnum = flag.String("webhook-enum", "", "Path of generated Golang PayPal enum file")
var payPalHTML = flag.String("paypal-html", "", "The paypal html file for code generation, use online html if empty")
var goPacakge = flag.String("go-pkg", "paypal", "Package name for generated go files")

func main() {
	flag.Parse()
	if err := do(context.Background()); err != nil {
		fmt.Println("ERR:", err)
	}
}

func do(ctx context.Context) error {
	if *webhookEnum != "" {
		return genPayPal(ctx)
	}

	return fmt.Errorf("nothing to do")
}

func genPayPal(ctx context.Context) (err error) {
	var bs []byte
	if *payPalHTML == "" {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, eventNamesPage, nil)
		if err != nil {
			return fmt.Errorf("new request: %w", err)
		}
		req.Header.Add("User-Agent", UA)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("get: %w", err)
		}
		defer res.Body.Close()
		bs, err = io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}
	} else {
		bs, err = os.ReadFile(*payPalHTML)
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}
	}

	wgs, err := parsePayPal(bs)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	g := Generator{
		Package: *goPacakge,
	}
	return g.WriteFile(wgs, *webhookEnum)
}
