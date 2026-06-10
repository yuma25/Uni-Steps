// Uni-Steps フロントエンド型定義ファイルである．

export interface Task {
  id: string;              // 課題の一意識別子である．
  group_id: string;        // 所属するグループの ID である．
  user_id: string;         // 担当するユーザーの ID である．
  source: string;          // 課題の入力元（manual, google_classroom 等）である．
  external_id: string;     // 外部 LMS における課題の ID である．
  title: string;           // 課題のタイトルである．
  deadline: string;        // 課題の期限である（ISO8601 形式）．
  is_completed: boolean;   // 完了したかどうかのフラグである．
  is_critical: boolean;    // 起床確認が必要な重要課題かどうかのフラグである．
}

export interface User {
  id: string;
  name: string;
  email?: string;
}

export interface Group {
  id: string;                  // グループの一意識別子である．
  name: string;                // グループの名称である．
  owner_id: string;            // オーナーのユーザー ID である．
  line_channel_token?: string; // BYOT 用の LINE トークンである．
  line_group_id?: string;      // 通知先の LINE グループ ID である．
}
