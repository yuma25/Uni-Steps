# インフラ・ネットワーク構成 (Zero-Cost Infrastructure)

## 1. サービス全体図
無料枠を最大活用し，かつ独自ドメイン風の運用を可能にする構成である．

```mermaid
graph LR
    User((ユーザー))
    Cron[Cron-job.org<br/>目覚まし時計]
    
    subgraph "パブリッククラウド"
        Vercel[Vercel<br/>フロントエンド: React]
        Render[Render<br/>バックエンド: Go API]
    end

    subgraph "外部 API / サービス"
        Supabase[(Supabase<br/>PostgreSQL)]
        Gemini[Gemini API<br/>AI エンジン]
        LINE[LINE API<br/>グループメッセージ]
    end

    User -- "https://uni-steps.vercel.app" --> Vercel
    Vercel -- "API リクエスト" --> Render
    Cron -- "5分おき Ping (/health)" --> Render
    
    Render -- "SQL" --> Supabase
    Render -- "JSON API" --> Gemini
    Render -- "Webhook/Push" --> LINE

    style User fill:#f96,stroke:#333
    style Vercel fill:#eee,stroke:#333
    style Render fill:#eee,stroke:#333
    style Cron fill:#fcf,stroke:#333
```

## 2. インフラ選定詳細

| コンポーネント | サービス名 | 無料枠の活用 | 独自ドメイン/URL |
| :--- | :--- | :--- | :--- |
| **Hosting (Static)** | Vercel | 帯域・ビルド無制限 | `*.vercel.app` (独自ドメイン可) |
| **Hosting (Compute)** | Render | 512MB RAM，CPU共有 | `*.onrender.com` (独自ドメイン可) |
| **Database** | Supabase | 500MB DB，5GB Storage | 内部APIキー接続 |
| **External API** | Gemini AI | 15 RPM (Flashモデル) | APIキー接続 |
| **Keeping Alive** | Cron-job.org | 無制限 | RenderのURLを定期実行 |
