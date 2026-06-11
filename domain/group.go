package domain

import "time"

type Group struct {
	ID               string    `json:"id" gorm:"primaryKey"`                // グループの一意識別子である．
	Name             string    `json:"name"`                                // グループの名称である．
	OwnerID          string    `json:"owner_id"`                            // グループの作成者（オーナー）のユーザー ID である．
	LineChannelToken string    `json:"line_channel_token"`                  // オーナーが持ち込んだ LINE Bot のアクセストークンである（BYOT方式）．
	LineGroupID      string    `json:"line_group_id"`                       // 通知を送信する先の LINE グループ ID である．
	LMSCourseID      string    `json:"lms_course_id"`                       // 紐付けられた外部 LMS（Google Classroom等）のコース ID である．
	LastSyncedAt     time.Time `json:"last_synced_at"`                      // 同期処理を最後に実行した時刻である（クールダウン用）．
	LMSLastUpdatedAt time.Time `json:"lms_last_updated_at"`                 // 外部 LMS 側で最後に情報が更新された時刻である（差分検知用）．
	InviteCode       string    `json:"invite_code" gorm:"uniqueIndex"`      // 参加用の招待コードである．
	Users            []*User   `json:"users" gorm:"many2many:user_groups;"` // グループに所属するユーザーのリストである．
}
