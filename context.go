package allymailerwrapper

import (
	"context"
	"net/http"

	allymailer "github.com/allymailer/allymailer-go-sdk"
)

type clientKey struct{}

// ClientMiddleware puts the API client into the context of every request.
func ClientMiddleware(c *allymailer.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(WithClient(r.Context(), c)))
		})
	}
}

// WithClient returns a copy of ctx with the API client.
func WithClient(ctx context.Context, c *allymailer.Client) context.Context {
	return context.WithValue(ctx, clientKey{}, c)
}

// ClientFromContext returns the API client of ClientMiddleware, nil when there is none.
func ClientFromContext(ctx context.Context) *allymailer.Client {
	c, _ := ctx.Value(clientKey{}).(*allymailer.Client)
	return c
}
