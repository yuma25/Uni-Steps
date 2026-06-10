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

type Task struct {
	ID              string      `json:"id"`               // 課題の一意識別子である．
	GroupID         string      `json:"group_id"`         // 所属するグループの ID である．
	UserID          string      `json:"user_id"`          // 担当するユーザーの ID である．（空の場合はグループ全員向け）
	Source          string      `json:"source"`           // 課題の入力元（manual, ai, google_classroom 等）である．
	ExternalID      string      `json:"external_id"`      // 外部 LMS における課題の ID である（重複登録防止用）．
	RawText         string      `json:"raw_text"`         // ユーザーが入力した生の文章である（AI 解析時のみ）．
	Title           string      `json:"title"`            // 課題のタイトルである．
	Deadline        time.Time   `json:"deadline"`         // 課題の期限（単発または初回）である．
	LMSUpdateTime   time.Time   `json:"lms_update_time"`  // 外部 LMS 側での最終更新日時である．
	Recurrence      string      `json:"recurrence"`       // 繰り返しの設定（none, weekly, biweekly, custom）である．
	CustomDeadlines []time.Time `json:"custom_deadlines"` // 特定の日付を複数選択した場合の期限リストである．
	IsCompleted     bool        `json:"is_completed"`     // 完了したかどうかのフラグである．
	IsCritical      bool        `json:"is_critical"`      // 起床確認が必要な重要課題かどうかのフラグである．
}
