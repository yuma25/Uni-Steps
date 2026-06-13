# フロントエンド カスタムフック リファレンス

ビジネスロジックやデータ取得をカプセル化したフックの仕様である．

---

## 🎣 `useAuth()`
アプリケーション全体の認証状態（現在のユーザー）にアクセスするためのフックである．
- **戻り値**: 
  - `userId string`: ログインユーザーの ID．
  - `user User | null`: ユーザー情報の詳細．
  - `loading boolean`: 取得中フラグ．

## 🎣 `useDashboardData(userId, groupId)`
ダッシュボードに必要な全データを一括管理するフックである．
- **管理データ**: `tasks`, `group`, `activeWakeup`, `groupWakeups`, `notificationLogs`．
- **メソッド**: 
  - `fetchData()`: サーバーから最新情報を一括で再取得する．
  - `setServerTokenMissing(bool)`: 通知用トークンの有無を更新する．

## 🎣 `useWebPush(userId, groupId)`
Web Push 通知の購読と確認を担うフックである．
- **機能**: 
  - 通知許可の要求と Service Worker の登録．
  - 410 エラー（購読切れ）時のサーバー側トークン削除依頼．
- **メソッド**: 
  - `handleEnableNotifications()`: 購読プロセスを開始する．
  - `handleSendTestNotification()`: 動作確認用のテスト通知を送信する．

---
*最終更新日: 2026年6月13日*
