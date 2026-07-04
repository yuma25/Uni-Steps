# Uni-Steps デプロイメントガイド

本ドキュメントは，Uni-Steps を本番環境（Vercel & Render）へデプロイする手順を解説するものである．

---

## 1. バックエンドのデプロイ (Render)

### 手順
1.  [Render](https://render.com/) にログインし，**"New + Web Service"** を選択する．
2.  Uni-Steps の GitHub リポジトリを接続する．
3.  設定を以下のように入力する：
    *   **Name**: `uni-steps-[ユニークな名前]` (例: `uni-steps-yuma25`)
        *   ※ `uni-steps` は既に使用されている可能性があるため，自分だけの名前を付けてください．
    *   **Runtime**: `Go`
    *   **Root Directory**: `backend` (★重要：backend フォルダをルートとして指定)
    *   **Build Command**: `go build -o main .`
    *   **Start Command**: `./main`
4.  **"Advanced"** をクリックし，以下の環境変数を設定する：
    *   `DATABASE_URL`: Supabase の接続文字列
    *   `GO_ENV`: `production` (重要：本番用設定の読み込みに使用)
    *   `GEMINI_API_KEY`: Google AI Studio のキー
    *   `GEMINI_MODEL`: `models/gemini-1.5-flash` 等
    *   `FRONTEND_URL`: `https://[あなたのアプリ名].vercel.app`
    *   `GOOGLE_CLIENT_ID`: Google Cloud Console から取得
    *   `GOOGLE_CLIENT_SECRET`: 同上
    *   `GOOGLE_REDIRECT_URL`: `https://[RenderのURL]/api/auth/google/callback`
    *   `VAPID_PUBLIC_KEY`: `cd backend && go run cmd/vapid/main.go` で生成
    *   `VAPID_PRIVATE_KEY`: 同上
    *   `VAPID_CONTACT`: `mailto:your-email@example.com`
5.  **"Create Web Service"** をクリックする．

---

## 2. フロントエンドのデプロイ (Vercel)

### 手順
1.  [Vercel](https://vercel.com/) にログインし，リポジトリをインポートする．
2.  **"Root Directory"** を `frontend` に設定する．
3.  **"Environment Variables"** に以下を追加する：
    *   `VITE_API_BASE_URL`: `https://[RenderのURL]` (末尾のスラッシュなし)
4.  **"Deploy"** をクリックする．

---

## 3. 外部サービスの最終設定

### 3.1 Google Cloud Console
1.  Uni-Steps プロジェクトの「認証情報」を開く．
2.  OAuth 2.0 クライアント ID の設定で，**「承認済みのリダイレクト URI」** に以下を追記する：
    *   `https://[RenderのURL]/api/auth/google/callback`

### 3.2 LINE Developers
1.  Messaging API 設定の **「Webhook URL」** を以下に更新する：
    *   `https://[RenderのURL]/api/line/webhook` (使用する場合)

---

## 4. 運用上の注意 (Render 無料枠)
Render の無料枠（Free Tier）を使用する場合，一定時間リクエストがないとインスタンスがスリープ状態になります．
*   **スリープからの復帰**: 初回アクセス時に 30秒〜1分 程度の待ち時間が発生します．
*   **通知の継続**: [Cron-job.org](https://cron-job.org/) 等を使用して，5〜10分おきに `/health` エンドポイントへ Ping を送ることで，スリープを回避し通知機能を維持することが可能です．
