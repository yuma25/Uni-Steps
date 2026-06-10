package domain

import "time"

type Group struct {
	ID               string    `json:"id"`                  // グループの一意識別子である．
	Name             string    `json:"name"`                // グループの名称である．
	LineWebhookURL   string    `json:"line_webhook_url"`    // LINE 通知用の Webhook URL である．
	LastSyncedAt     time.Time `json:"last_synced_at"`      // 同期処理を最後に実行した時刻である（クールダウン用）．
	LMSLastUpdatedAt time.Time `json:"lms_last_updated_at"` // 外部 LMS 側で最後に情報が更新された時刻である（差分検知用）．
}
