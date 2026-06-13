import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import LoginPage from './pages/LoginPage';
import DashboardPage from './pages/DashboardPage';
import GroupSelectionPage from './pages/GroupSelectionPage';
import { AuthProvider } from './context/AuthContext';
import './App.css';

// アプリケーションのメインコンポーネントである．
// ルーティングの設定を行う．
function App() {
  return (
    <Router>
      <AuthProvider>
        <Routes>
          <Route path="/" element={<LoginPage />} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="/select-group" element={<GroupSelectionPage />} />
          <Route path="/dashboard" element={<DashboardPage />} />
        </Routes>
      </AuthProvider>
    </Router>
  );
}

export default App;
