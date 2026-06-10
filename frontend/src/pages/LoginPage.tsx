import React from 'react';

// ログイン画面のコンポーネントである．
const LoginPage: React.FC = () => {
  const handleGoogleLogin = () => {
    // バックエンドの Google OAuth ログインエンドポイントへリダイレクトする．
    const backendUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
    window.location.href = `${backendUrl}/api/auth/google/login`;
  };

  return (
    <div className="login-container">
      <h1>Uni-Steps</h1>
      <p>AIと仲間が支える課題管理プラットフォーム</p>
      <button onClick={handleGoogleLogin} className="google-login-button">
        Google でログイン
      </button>
    </div>
  );
};

export default LoginPage;
