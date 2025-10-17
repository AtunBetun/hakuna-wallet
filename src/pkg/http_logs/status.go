package http_logs

import (
	"net/http"

	"github.com/atunbetun/hakuna-wallet/pkg/logger"
	"go.uber.org/zap"
)

// LoggingRoundTripper wraps another RoundTripper and logs status codes
type LoggingRoundTripper struct {
	Base http.RoundTripper
}

func (lrt LoggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Use the wrapped RoundTripper
	resp, err := lrt.Base.RoundTrip(req)

	if err != nil {
		logger.Logger.Error("HTTP request failed", zap.Error(err), zap.String("url", req.URL.String()))
		return nil, err
	}

	logger.Logger.Info("HTTP request completed",
		zap.String("method", req.Method),
		zap.String("url", req.URL.String()),
		zap.Int("status", resp.StatusCode),
	)

	return resp, nil
}

// Usage
func NewLoggingClient() *http.Client {
	return &http.Client{
		Transport: LoggingRoundTripper{
			Base: http.DefaultTransport,
		},
	}
}
