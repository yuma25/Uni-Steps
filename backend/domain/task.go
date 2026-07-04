package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	SourceManual          = "manual"           // 手動入力による登録である．
	SourceGoogleClassroom = "google_classroom" // Google Classroom からの取得である．
	SourceWebClass        = "web_class"        // Web Class からの取得である．
)

// RecurrenceSettings は課題の繰り返しに関する設定を保持する構造体である．
// JSON 形式で一つのカラムに保存される．
type RecurrenceSettings struct {
	Type        string      `json:"type"`         // none, weekly, biweekly, custom
	CustomDates []time.Time `json:"custom_dates"` // 特定の日付（複数）を指定する場合のリスト
}

// Task は課題の基本情報を保持する構造体である．
type Task struct {
	ID               string             `json:"id" gorm:"primaryKey"`                              // 課題の一意識別子である．
	GroupID          string             `json:"group_id" gorm:"uniqueIndex:idx_group_external"`    // 所属するグループの ID である．
	Source           string             `json:"source"`                                            // 課題の入力元（manual, ai, google_classroom 等）である．
	ExternalID       string             `json:"external_id" gorm:"uniqueIndex:idx_group_external"` // 外部 LMS における課題の ID である（重複登録防止用）．
	RawText          string             `json:"raw_text"`                                          // ユーザーが入力した生の文章である（AI 解析時のみ）．
	CreatorID        string             `json:"creator_id"`                                        // 課題の作成者のユーザー ID である（手動課題のみ）．
	Title            string             `json:"title"`                                             // 課題のタイトルである．
	Deadline         time.Time          `json:"deadline"`                                          // 課題の期限（単発または初回）である．
	IsLMSDeadlineSet bool               `json:"is_lms_deadline_set"`                               // 外部 LMS 側で期限があらかじめ設定されていたか．
	LMSUpdateTime    time.Time          `json:"lms_update_time"`                                   // 外部 LMS 側での最終更新日時である．
	Recurrence       RecurrenceSettings `json:"recurrence" gorm:"serializer:json"`                 // 繰り返しの設定（JSON 形式で統合管理）．

	// Has-Many 関係：タスクごとの各ユーザーの進捗状況
	UserProgress []*TaskUserProgress `json:"user_progress" gorm:"foreignKey:TaskID"`
}

// TaskUserProgress はある課題に対する特定のユーザーの完了状態を保持する．
type TaskUserProgress struct {
	TaskID      string    `json:"task_id" gorm:"primaryKey"`
	UserID      string    `json:"user_id" gorm:"primaryKey"`
	UserName    string    `json:"user_name"`    // 画面表示用に保持する．
	IsCompleted bool      `json:"is_completed"` // そのユーザーが完了したか．
	UpdatedAt   time.Time `json:"updated_at"`
}

// SetupManualDefaults は手動登録時の初期値設定を行う．
func (t *Task) SetupManualDefaults() {
	t.Source = SourceManual
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	if t.ExternalID == "" {
		t.ExternalID = t.ID
	}
	for _, up := range t.UserProgress {
		up.TaskID = t.ID
		if up.UpdatedAt.IsZero() {
			up.UpdatedAt = time.Now()
		}
	}
}

// Validate は課題のドメインルールに基づくバリデーションを行う．
func (t *Task) Validate() error {
	if t.Title == "" {
		return fmt.Errorf("タイトルは必須である")
	}
	return nil
}
