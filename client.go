package paypal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

type Client struct {
	base, id, secret string

	t  *Token
	hc *http.Client
}

func NewClient(base, id, secret string) *Client {
	return &Client{
		base:   base,
		id:     id,
		secret: secret,
		hc: &http.Client{
			Transport: otelhttp.NewTransport(
				http.DefaultTransport,
				otelhttp.WithSpanNameFormatter(formatSpanName),
				otelhttp.WithSpanOptions(
					trace.WithAttributes(semconv.PeerServiceKey.String("paypal")),
				),
			),
		},
	}
}

func formatSpanName(_ string, r *http.Request) string {
	op := GetOperation(r.Context())
	if op == "" {
		// Fallback to the default name
		op = r.Method
	}
	return "PayPal " + op
}

type operationKey struct{}

func WithOperation(ctx context.Context, op string) context.Context {
	return context.WithValue(ctx, operationKey{}, op)
}

func GetOperation(ctx context.Context) string {
	return ctx.Value(operationKey{}).(string)
}

type Token struct {
	Scope       string `json:"scope"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	AppID       string `json:"app_id"`
	Nonce       string `json:"nonce"`
	ExpiresIn   int    `json:"expires_in"`
	expiresAt   time.Time
}

func (t *Token) Valid() bool {
	if t == nil {
		return false
	}
	return t.expiresAt.After(time.Now())
}

// Auth requests a new token from PayPal server.
// See https://developer.paypal.com/api/rest/authentication/.
func (c *Client) Auth(ctx context.Context) (res *Token, err error) {
	ctx = WithOperation(ctx, "Auth")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.base+"/v1/oauth2/token", strings.NewReader("grant_type=client_credentials"))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.SetBasicAuth(c.id, c.secret)

	start := time.Now()
	if res, err = doJSON[Token](ctx, c, req); err != nil {
		return
	}

	// Minus 2 seconds is to prevent expiration due to network latency.
	res.expiresAt = start.Add(time.Duration(res.ExpiresIn-2) * time.Second)
	return
}

func (c *Client) checkToken(ctx context.Context) (err error) {
	if !c.t.Valid() {
		c.t, err = c.Auth(ctx)
	}
	return
}

// JSON performs the request with the data marshaled to JSON format,
// unmarshals the response body into a new R,
// and automatically refreshes the client's access token.
func JSON[R any](ctx context.Context, c *Client, method, path string, data any,
) (res *R, err error) {
	url := c.base + path
	req, err := NewJSONRequest(ctx, method, url, data)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	if err = c.checkToken(ctx); err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+c.t.AccessToken)
	return doJSON[R](ctx, c, req)
}

// JSONNop is similar to [JSON] but with the response body discarded.
func JSONNop(ctx context.Context, c *Client, method, path string, data any) (err error) {
	url := c.base + path
	req, err := NewJSONRequest(ctx, method, url, data)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	if err = c.checkToken(ctx); err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+c.t.AccessToken)
	hres, err := c.hc.Do(req)
	if err != nil {
		return fmt.Errorf("do: %w", err)
	}
	defer hres.Body.Close()
	if hres.StatusCode < 400 {
		return nil
	}
	if hres.ContentLength == 0 {
		return &Error{
			StatusCode: hres.StatusCode,
		}
	}
	e, err := RespJSON[Error](hres)
	if err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}
	e.StatusCode = hres.StatusCode
	return e
}

func doJSON[R any](ctx context.Context, c *Client, req *http.Request) (res *R, err error) {
	hres, err := c.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do: %w", err)
	}

	if hres.StatusCode < 400 {
		return RespJSON[R](hres)
	}
	e, err := RespJSON[Error](hres)
	if err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}
	e.StatusCode = hres.StatusCode
	return nil, e
}

// Error is the PayPal API error response.
// See https://developer.paypal.com/api/rest/responses/.
type Error struct {
	StatusCode int

	Name    string         `json:"name"`
	Message string         `json:"message"`
	DebugID string         `json:"debug_id"`
	Details []*ErrorDetail `json:"details"`
	Links   []*Link        `json:"links"`

	// For identity errors
	Err     string `json:"error"`
	ErrDesc string `json:"error_description"`
}

func (e *Error) Error() string {
	if e.Err != "" {
		return e.Err + ": " + e.ErrDesc
	}
	return fmt.Sprintf("%s: %s (%s)", e.Name, e.Message, e.DebugID)
}

type ErrorDetail struct {
	Field       string `json:"field"`
	Value       string `json:"value"`
	Location    string `json:"location"`
	Issue       string `json:"issue"`
	Description string `json:"description"`
}

// Link is a HATEOAS link.
// See https://developer.paypal.com/api/rest/responses/#link-hateoaslinks.
type Link struct {
	HRef   string `json:"href"`
	Rel    string `json:"rel"`
	Method string `json:"method"`
}

// NewJSONRequest returns a new [http.Request] with the given data marshaled to JSON format.
func NewJSONRequest(ctx context.Context, method, url string, data any,
) (res *http.Request, err error) {
	var r io.Reader
	if data != nil {
		bs, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		r = bytes.NewReader(bs)
	}
	res, err = http.NewRequestWithContext(ctx, method, url, r)
	if err != nil {
		return
	}
	res.Header.Set("Content-Type", "application/json")
	return
}

// RespJSON unmarshals the response body into a new R and closes the body afterward.
func RespJSON[R any](r *http.Response) (res *R, err error) {
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}
	res = new(R)
	err = json.Unmarshal(bs, res)
	return
}
