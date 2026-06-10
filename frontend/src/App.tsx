import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import LoginPage from './pages/LoginPage';
import './App.css';

// アプリケーションのメインコンポーネントである．
// ルーティングの設定を行う．
function App() {
  return (
    <Router>
      <div className="app-container">
        <main>
          <Routes>
            <Route path="/" element={<LoginPage />} />
            <Route path="/login" element={<LoginPage />} />
            {/* 今後ここに Dashboard 等のルートを追加していく */}
          </Routes>
        </main>
      </div>
    </Router>
  );
}

export default App;
