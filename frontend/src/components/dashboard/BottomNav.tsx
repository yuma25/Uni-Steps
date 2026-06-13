import React from 'react';
import { Home, Clock, Sparkles, Settings } from 'lucide-react';

interface BottomNavProps {
  activeTab: string;
  onTabChange: (tab: string) => void;
}

const BottomNav: React.FC<BottomNavProps> = ({ activeTab, onTabChange }) => {
  return (
    <nav className="bottom-nav">
      <div className="sidebar-logo">
        <Sparkles size={28} />
        <span>Uni-Steps</span>
      </div>

      <button 
        onClick={() => onTabChange('tasks')} 
        className={`nav-item ${activeTab === 'tasks' ? 'active' : ''}`}
      >
        <Home size={24} />
        <span>ホーム</span>
      </button>
      <button 
        onClick={() => onTabChange('wakeup')} 
        className={`nav-item ${activeTab === 'wakeup' ? 'active' : ''}`}
      >
        <Clock size={24} />
        <span>起床見守り</span>
      </button>
      <button 
        onClick={() => onTabChange('timeline')} 
        className={`nav-item ${activeTab === 'timeline' ? 'active' : ''}`}
      >
        <Sparkles size={24} />
        <span>活動履歴</span>
      </button>
      <button 
        onClick={() => onTabChange('settings')} 
        className={`nav-item ${activeTab === 'settings' ? 'active' : ''}`}
      >
        <Settings size={24} />
        <span>設定</span>
      </button>
    </nav>
  );
};

export default BottomNav;
