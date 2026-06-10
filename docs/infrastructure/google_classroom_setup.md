# Google Classroom API 連携設定ガイド

本ドキュメントでは，Uni-Steps と Google Classroom を連携させるために必要な，Google Cloud Console での設定手順を詳細に解説する．

---

## 1．Google Cloud プロジェクトの準備
Google の各種 API を利用するための「プロジェクト」という管理単位を作成する．

1.  [Google Cloud Console](https://console.cloud.google.com/) にログインする．
2.  画面上部のプロジェクト選択メニューから **「新しいプロジェクト」** を作成する．
    *   プロジェクト名： `Uni-Steps`
3.  作成したプロジェクトが選択されていることを確認する．

## 2．API の有効化
プロジェクトに対して，Google Classroom と通信するための機能を許可する．

1.  サイドメニューから **「API とサービス」 > 「ライブラリ」** を開く．
2.  検索ボックスに **「Google Classroom API」** と入力し，該当する API を選択する．
3.  **「有効にする」** ボタンをクリックする．

## 3．OAuth 同意画面の設定
ユーザーが Google ログインを行う際に表示される，権限の確認画面を構成する．

1.  サイドメニューから **「OAuth 同意画面」** を開く．
2.  User Type で **「外部」** を選択し，「作成」をクリックする．
3.  **アプリ登録の編集**:
    *   アプリ名： `Uni-Steps`
    *   ユーザーサポートメール： 自身のメールアドレス
    *   デベロッパーの連絡先： 自身のメールアドレス
4.  **スコープの追加**:
    「スコープを追加または削除」をクリックし，以下の URL 群を「制限付きスコープ」として手動追加する：
    *   `openid` (ユーザーの識別)
    *   `https://www.googleapis.com/auth/userinfo.email` (メールアドレスの取得)
    *   `https://www.googleapis.com/auth/userinfo.profile` (プロフィール情報の取得)
    *   `https://www.googleapis.com/auth/classroom.courses.readonly` (コース一覧の取得)
    *   `https://www.googleapis.com/auth/classroom.coursework.me.readonly` (課題一覧の取得)
5.  「保存して次へ」を繰り返し，設定を完了させる．

## 4．認証情報の作成
アプリケーションが Google のサーバーと安全に通信するための ID と鍵を発行する．

1.  サイドメニューから **「認証情報」** を開く．
2.  **「認証情報を作成」 > 「OAuth クライアント ID」** を選択する．
3.  アプリケーションの種類： **「ウェブ アプリケーション」** を選択する．
4.  名前： `Uni-Steps Web Client` (任意)
5.  **承認済みのリダイレクト URI**:
    「URI を追加」をクリックし，以下の URL を入力する：
    *   `http://localhost:8080/api/auth/google/callback`
    *   ※注意：末尾の `/` の有無まで正確に一致させる必要がある．
6.  「作成」をクリックし，表示された **「クライアント ID」** と **「クライアント シークレット」** を控える．

## 5．環境変数の設定
取得した値をプロジェクトルートの `.env` ファイルに設定する．

```text
GOOGLE_CLIENT_ID="[取得したクライアント ID]"
GOOGLE_CLIENT_SECRET="[取得したクライアント シークレット]"
GOOGLE_REDIRECT_URL="http://localhost:8080/api/auth/google/callback"
```

## 6．補足事項：テストユーザーの登録
OAuth 同意画面が「テスト」フェーズにある場合，許可されたユーザーしかログインできない．
1.  「OAuth 同意画面」の設定ページ下部にある **「Test users」** セクションに，自身の Gmail アドレスを追加しておくこと．
