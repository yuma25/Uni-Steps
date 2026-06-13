import React from 'react';
import { ArrowLeft, Users, Share2, Send, BellRing, Plus, Settings } from 'lucide-react';
import type { Group } from '../../types';

interface DashboardHeaderProps {
  group: Group | null;
  userId: string;
  onBack: () => void;
  onSendTestNotification: () => void;
  onEnableNotifications: () => void;
  onAddTask: () => void;
  onOpenSettings: () => void;
  notifPermission: NotificationPermission;
  serverTokenMissing: boolean;
}

const DashboardHeader: React.FC<DashboardHeaderProps> = ({
  group,
  userId,
  onBack,
  onSendTestNotification,
  onEnableNotifications,
  onAddTask,
  onOpenSettings,
  notifPermission,
  serverTokenMissing
}) => {
  const copyInviteCode = () => {
    if (group?.invite_code) {
      navigator.clipboard.writeText(group.invite_code);
      alert("コピーしました！");
    }
  };

  return (
    <header className="dashboard-header">
      <div className="header-left">
        <button onClick={onBack} className="back-button"><ArrowLeft size={20} /></button>
        <div className="title-group">
          <div className="title-row">
            <h1>{group?.name || "Uni-Steps"}</h1>
            {group?.invite_code && (
              <button onClick={copyInviteCode} className="invite-badge">
                <Share2 size={12} />コード: {group.invite_code}
              </button>
            )}
          </div>
          <div className="group-info">
            {group?.users && (
              <div className="member-list-summary">
                <Users size={12} style={{marginRight: '4px'}} />
                {group.users.map(u => u.name).join(', ')}
              </div>
            )}
          </div>
        </div>
      </div>
      <div className="header-actions">
        {notifPermission === 'granted' && !serverTokenMissing && (
          <button onClick={onSendTestNotification} className="icon-button" title="通知テスト">
            <Send size={16} /><span>テスト</span>
          </button>
        )}
        {(notifPermission !== 'granted' || serverTokenMissing) && (
          <button onClick={onEnableNotifications} className="icon-button warning-btn">
            <BellRing size={16} /><span>通知を有効化</span>
          </button>
        )}
        <button onClick={onAddTask} className="icon-button primary">
          <Plus size={16} />課題追加
        </button>
        {group?.owner_id === userId && (
          <button onClick={onOpenSettings} className="icon-button" title="設定">
            <Settings size={18} />
          </button>
        )}
      </div>
    </header>
  );
};

export default DashboardHeader;
