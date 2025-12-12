package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"waverless/pkg/config"
	"waverless/pkg/logger"
)

// FeishuNotifier sends notifications to Feishu (Lark)
type FeishuNotifier struct {
	webhookURL string
	client     *http.Client
}

// NewFeishuNotifier creates a new Feishu notifier
func NewFeishuNotifier() *FeishuNotifier {
	// Priority: config file > environment variable
	var webhookURL string
	if config.GlobalConfig != nil && config.GlobalConfig.Notification.FeishuWebhookURL != "" {
		webhookURL = config.GlobalConfig.Notification.FeishuWebhookURL
		logger.Info("Using Feishu webhook URL from config file")
	} else {
		webhookURL = os.Getenv("FEISHU_WEBHOOK_URL")
		if webhookURL != "" {
			logger.Info("Using Feishu webhook URL from environment variable")
		}
	}

	if webhookURL == "" {
		logger.Warn("Feishu webhook URL not configured (check config file or FEISHU_WEBHOOK_URL env), Feishu notifications will be disabled")
	}

	return &FeishuNotifier{
		webhookURL: webhookURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ImageUpdateNotification represents an image update notification
type ImageUpdateNotification struct {
	Endpoint      string
	CurrentImage  string
	NewImageTag   string
	ImagePrefix   string
	DetectedAt    time.Time
	DetectionType string // "webhook" or "manual"
}

// SendImageUpdateNotification sends image update notification to Feishu
func (f *FeishuNotifier) SendImageUpdateNotification(ctx context.Context, notification *ImageUpdateNotification) error {
	if f.webhookURL == "" {
		logger.WarnCtx(ctx, "Feishu webhook URL not configured, skipping notification")
		return nil
	}

	// Build Feishu message card
	message := f.buildImageUpdateMessage(notification)

	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal Feishu message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", f.webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Feishu notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Feishu API returned status code: %d", resp.StatusCode)
	}

	logger.InfoCtx(ctx, "Feishu notification sent successfully for endpoint: %s", notification.Endpoint)
	return nil
}

// buildImageUpdateMessage builds a Feishu message card for image updates
func (f *FeishuNotifier) buildImageUpdateMessage(notification *ImageUpdateNotification) map[string]interface{} {
	detectionTypeText := "Auto-detected (Webhook)"
	if notification.DetectionType == "manual" {
		detectionTypeText = "Manual Check"
	}

	return map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"header": map[string]interface{}{
				"template": "orange",
				"title": map[string]interface{}{
					"content": "🔔 Image Update Available",
					"tag":     "plain_text",
				},
			},
			"elements": []interface{}{
				map[string]interface{}{
					"tag": "div",
					"text": map[string]interface{}{
						"content": fmt.Sprintf("**Endpoint**: %s\nNew Docker image version detected", notification.Endpoint),
						"tag":     "lark_md",
					},
				},
				map[string]interface{}{
					"tag": "hr",
				},
				map[string]interface{}{
					"tag": "div",
					"fields": []interface{}{
						map[string]interface{}{
							"is_short": true,
							"text": map[string]interface{}{
								"content": fmt.Sprintf("**Current Image**\n%s", notification.CurrentImage),
								"tag":     "lark_md",
							},
						},
						map[string]interface{}{
							"is_short": true,
							"text": map[string]interface{}{
								"content": fmt.Sprintf("**New Image**\n%s", notification.NewImageTag),
								"tag":     "lark_md",
							},
						},
					},
				},
				map[string]interface{}{
					"tag": "div",
					"fields": []interface{}{
						map[string]interface{}{
							"is_short": true,
							"text": map[string]interface{}{
								"content": fmt.Sprintf("**Image Prefix**\n%s", notification.ImagePrefix),
								"tag":     "lark_md",
							},
						},
						map[string]interface{}{
							"is_short": true,
							"text": map[string]interface{}{
								"content": fmt.Sprintf("**Detection Type**\n%s", detectionTypeText),
								"tag":     "lark_md",
							},
						},
					},
				},
				map[string]interface{}{
					"tag": "div",
					"text": map[string]interface{}{
						"content": fmt.Sprintf("**Detection Time**: %s", notification.DetectedAt.Format("2006-01-02 15:04:05")),
						"tag":     "lark_md",
					},
				},
				map[string]interface{}{
					"tag": "hr",
				},
				map[string]interface{}{
					"tag": "note",
					"elements": []interface{}{
						map[string]interface{}{
							"content": "⚠️ Please confirm if you need to update the endpoint's image version. For stability reasons, the system will not update automatically.",
							"tag":     "plain_text",
						},
					},
				},
			},
		},
	}
}
