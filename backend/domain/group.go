package domain

import "time"

const (
	AICharacterDefault = "default" // 標準的な通知スタイルである．
	AICharacterStrict  = "strict"  // 厳しい指導官スタイルである．
	AICharacterKind    = "kind"    // 心配性な幼馴染スタイルである．
	AICharacterCool    = "cool"    // 冷徹な執事スタイルである．
)

type Group struct {
	ID                 string    `json:"id" gorm:"primaryKey"`                    // グループの一意識別子である．
	Name               string    `json:"name"`                                    // グループの名称である．
	OwnerID            string    `json:"owner_id"`                                // グループの作成者（オーナー）のユーザー ID である．
	LineChannelToken   string    `json:"line_channel_token"`                      // オーナーが持ち込んだ LINE Bot のアクセストークンである（BYOT方式）．
	LineGroupID        string    `json:"line_group_id"`                           // 通知を送信する先の LINE グループ ID である．
	LastSyncedAt       time.Time `json:"last_synced_at"`                          // 同期処理を最後に実行した時刻である（クールダウン用）．
	LMSLastUpdatedAt   time.Time `json:"lms_last_updated_at"`                     // 外部 LMS 側で最後に情報が更新された時刻である（差分検知用）．
	InviteCode         string    `json:"invite_code" gorm:"uniqueIndex"`          // 参加用の招待コードである．
	RemindIntervals    []int     `json:"remind_intervals" gorm:"serializer:json"` // 何分前に通知するか（複数設定可）
	AICharacter        string    `json:"ai_character"`                            // AI の性格設定である．
	SummaryMorningTime string    `json:"summary_morning_time"`                    // 朝のサマリー送信時刻 (HH:mm) である．
	SummaryEveningTime string    `json:"summary_evening_time"`                    // 夜のサマリー送信時刻 (HH:mm) である．
	Users              []*User   `json:"users" gorm:"many2many:user_groups;"`     // グループに所属するユーザーのリストである．
}

const (
	SummaryTypeMorning = "morning"
	SummaryTypeEvening = "evening"
)
