package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"time"

	"github.com/achmadnp/peroot/config"
	"github.com/achmadnp/peroot/model"
)

type WhatsAppService interface {
	SendOrderNotification(ctx context.Context, order *model.OrderRequest) (string, error)
}

type whatsAppService struct {
	cfg    *config.Config
	client *http.Client
}

func NewWhatsAppService(cfg *config.Config) WhatsAppService {
	return &whatsAppService{
		cfg: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *whatsAppService) SendOrderNotification(ctx context.Context, order *model.OrderRequest) (string, error) {
	msg := s.buildOrderMessage(order)

	body, err := json.Marshal(msg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal WhatsApp message: %w", err)
	}

	var lastErr error
	maxRetries := 3

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			slog.Warn("Retrying WhatsApp API call",
				"attempt", attempt,
				"backoff_seconds", backoff.Seconds(),
				"last_error", lastErr,
			)

			select {
			case <-ctx.Done():
				return "", fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			case <-time.After(backoff):
			}
		}

		messageID, err := s.doSend(ctx, body)
		if err == nil {
			if attempt > 0 {
				slog.Info("WhatsApp API call succeeded after retry", "attempt", attempt)
			}
			return messageID, nil
		}

		lastErr = err
		slog.Error("WhatsApp API call failed", "attempt", attempt, "error", err)
	}

	return "", fmt.Errorf("WhatsApp API call failed after %d retries: %w", maxRetries, lastErr)
}

func (s *whatsAppService) doSend(ctx context.Context, body []byte) (string, error) {
	url := fmt.Sprintf("%s/%s/messages", s.cfg.WhatsAppAPIURL, s.cfg.WhatsAppPhoneNumberID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.cfg.WhatsAppAccessToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var waErr model.WAErrorResponse
		if jsonErr := json.Unmarshal(respBody, &waErr); jsonErr == nil && waErr.Error.Message != "" {
			return "", fmt.Errorf("WhatsApp API error (HTTP %d): %s [code: %d, trace: %s]",
				resp.StatusCode,
				waErr.Error.Message,
				waErr.Error.Code,
				waErr.Error.FBTraceID,
			)
		}
		return "", fmt.Errorf("WhatsApp API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var waResp model.WAMessageResponse
	if err := json.Unmarshal(respBody, &waResp); err != nil {
		return "", fmt.Errorf("failed to parse WhatsApp response: %w", err)
	}

	if len(waResp.Messages) == 0 {
		return "", fmt.Errorf("WhatsApp API returned success but no message ID")
	}

	return waResp.Messages[0].ID, nil
}

func (s *whatsAppService) buildOrderMessage(order *model.OrderRequest) *model.WAMessageRequest {
	totalFormatted := fmt.Sprintf("%.2f %s", order.TotalAmount, order.Currency)

	itemSummary := ""
	for i, item := range order.Items {
		if i > 0 {
			itemSummary += ", "
		}
		itemSummary += fmt.Sprintf("%dx %s", item.Quantity, item.Name)
	}

	return &model.WAMessageRequest{
		MessagingProduct: "whatsapp",
		To:               s.cfg.WhatsAppRecipient,
		Type:             "template",
		Template: &model.WATemplate{
			Name: s.cfg.WhatsAppTemplateName,
			Language: &model.WALanguage{
				Code: s.cfg.WhatsAppTemplateLang,
			},
			Components: []model.WAComponent{
				{
					Type: "body",
					Parameters: []model.WAParameter{
						{Type: "text", Text: order.OrderID},
						{Type: "text", Text: order.CustomerName},
						{Type: "text", Text: totalFormatted},
						{Type: "text", Text: itemSummary},
					},
				},
			},
		},
	}
}
