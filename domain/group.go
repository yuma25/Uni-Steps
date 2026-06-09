package domain

type Group struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	LineWebhookURL string `json:"line_webhook_url"`
}
