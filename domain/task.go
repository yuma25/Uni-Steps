package domain

import (
	"time"
)

const (
	SourceManual          = "manual"           // 手動入力による登録である．
	SourceAI              = "ai"               // AI 解析による登録である．
	SourceGoogleClassroom = "google_classroom" // Google Classroom からの取得である．
	SourceWebClass        = "web_class"        // Web Class からの取得である．
)

const (
	RecurrenceNone     = "none"     // 繰り返しの設定なしである．
	RecurrenceWeekly   = "weekly"   // 毎週の繰り返しである．
	RecurrenceBiweekly = "biweekly" // 隔週の繰り返しである．
	RecurrenceCustom   = "custom"   // 特定の日付（複数）を指定する繰り返しである．
)

// Task は課題の基本情報を保持する構造体である．
type Task struct {
	ID              string      `json:"id" gorm:"primaryKey"`                    // 課題の一意識別子である．
	GroupID         string      `json:"group_id"`                                // 所属するグループの ID である．
	Source          string      `json:"source"`                                  // 課題の入力元（manual, ai, google_classroom 等）である．
	ExternalID      string      `json:"external_id" gorm:"uniqueIndex"`          // 外部 LMS における課題の ID である（重複登録防止用）．
	RawText         string      `json:"raw_text"`                                // ユーザーが入力した生の文章である（AI 解析時のみ）．
	Title           string      `json:"title"`                                   // 課題のタイトルである．
	Deadline        time.Time   `json:"deadline"`                                // 課題の期限（単発または初回）である．
	LMSUpdateTime   time.Time   `json:"lms_update_time"`                         // 外部 LMS 側での最終更新日時である．
	Recurrence      string      `json:"recurrence"`                              // 繰り返しの設定（none, weekly, biweekly, custom）である．
	CustomDeadlines []time.Time `json:"custom_deadlines" gorm:"serializer:json"` // 特定の日付を複数選択した場合の期限リストである．JSON 形式で DB に保存される．
	IsCritical      bool        `json:"is_critical"`                             // 起床確認が必要な重要課題かどうかのフラグである．

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
