import React from 'react';

/**
 * ログイン画面のコンポーネントである．
 */
const LoginPage: React.FC = () => {
  const handleGoogleLogin = () => {
    // バックエンドの Google OAuth ログインエンドポイントへリダイレクトする．
    const backendUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
    window.location.href = `${backendUrl}/api/auth/google/login`;
  };

  return (
    <div className="login-page-container">
      {/* 装飾用の背景要素（オプション） */}
      <div className="portal-bg-shapes" style={{display: 'block', opacity: 0.5}}>
        <div className="portal-bg-shape shape-1" style={{background: '#fef3c7', top: '-10%', left: '-10%'}}></div>
        <div className="portal-bg-shape shape-2" style={{background: '#dcfce7', bottom: '-10%', right: '-10%'}}></div>
      </div>

      <div className="login-card">
        <div className="login-brand">
          <h1>Uni-Steps</h1>
          <p>AIと仲間が支える<br />課題管理プラットフォーム</p>
        </div>

        <button onClick={handleGoogleLogin} className="google-login-btn">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M23.49 12.275c0-.825-.075-1.65-.225-2.438H12v4.613h6.45c-.262 1.425-1.088 2.625-2.287 3.45v2.85h3.675c2.175-2.025 3.652-5.025 3.652-8.475z" fill="#4285F4"/>
            <path d="M12 24c3.24 0 5.962-1.088 7.95-2.912l-3.675-2.85c-1.013.675-2.325 1.088-4.275 1.088-3.263 0-6.037-2.213-7.013-5.213H1.275v2.963C3.262 21.037 7.425 24 12 24z" fill="#34A853"/>
            <path d="M4.987 14.112c-.225-.675-.375-1.425-.375-2.112s.15-1.437.375-2.112V6.925H1.275C.45 8.437 0 10.162 0 12s.45 3.563 1.275 5.075l3.712-2.963z" fill="#FBBC05"/>
            <path d="M12 4.763c1.763 0 3.338.6 4.575 1.837l3.413-3.412C17.925 1.125 15.24 0 12 0 7.425 0 3.263 2.962 1.275 7.037l3.713 2.963c.975-3.037 3.75-5.237 7.012-5.237z" fill="#EA4335"/>
          </svg>
          <span>Google でログイン</span>
        </button>

        <div className="login-footer-note">
          <p>ログインすることで，利用規約および<br />プライバシーポリシーに同意したものとみなされます．</p>
          <div style={{marginTop: '1.5rem', fontWeight: 700, color: '#cbd5e0', letterSpacing: '0.05em'}}>© 2026 Uni-Steps</div>
        </div>
      </div>
    </div>
  );
};

export default LoginPage;
