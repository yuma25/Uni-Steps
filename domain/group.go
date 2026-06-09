package domain

type Group struct {
	ID             string `json:"id"`               // グループの一意識別子である．
	Name           string `json:"name"`             // グループの名称である．
	LineWebhookURL string `json:"line_webhook_url"` // LINE 通知用の Webhook URL である．
}
