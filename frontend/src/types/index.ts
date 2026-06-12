// Uni-Steps フロントエンド型定義ファイルである．

export interface TaskUserProgress {
  task_id: string;
  user_id: string;
  user_name: string;
  is_completed: boolean;
  updated_at: string;
}

export interface RecurrenceSettings {
  type: string;           // none, weekly, biweekly, custom
  custom_dates?: string[]; // ISO8601 形式のリスト
}

export interface Task {
  id: string;              // 課題の一意識別子である．
  group_id: string;        // 所属するグループの ID である．
  source: string;          // 課題の入力元（manual, google_classroom 等）である．
  external_id: string;     // 外部 LMS における課題의 ID である．
  title: string;           // 課題のタイトルである．
  deadline: string;        // 課題の期限である（ISO8601 形式）．
  is_lms_deadline_set: boolean; // 外部 LMS 側で最初から期限があったか
  recurrence: RecurrenceSettings; // 繰り返しの設定（統合オブジェクト）
  user_progress: TaskUserProgress[]; // 各ユーザーの完了状態
}

export interface User {
  id: string;
  name: string;
  email?: string;
  has_push_token?: boolean; // サーバーに Web Push トークンが保存されているか
}

export interface WakeupCheck {
  id: string;
  user_id: string;
  group_id: string;
  target_time: string;
  grace_minutes: number;
  status: 'pending' | 'confirmed' | 'alerted';
  created_at: string;
}

export interface Group {
  id: string;                  // グループの一意識別子である．
  name: string;                // グループの名称である．
  owner_id: string;            // オーナーのユーザー ID である．
  invite_code: string;         // 参加用の招待コードである．
  line_channel_token?: string; // BYOT 用の LINE トークンである．
  line_group_id?: string;      // 通知先の LINE グループ ID である．
  last_synced_at?: string;     // 最終同期時刻である．
  lms_last_updated_at?: string; // LMS 側の最終更新時刻である．
  remind_intervals: number[];   // リマインド通知のタイミングである．
  ai_character: string;        // AI の性格設定である．
  users?: User[];              // 所属メンバー
}
