import React from 'react';
import { ArrowLeft, Users, Share2, BellRing, Plus } from 'lucide-react';
import type { Group } from '../../types';

interface DashboardHeaderProps {
  group: Group | null;
  userId: string;
  onBack: () => void;
  onEnableNotifications: () => void;
  onAddTask: () => void;
  notifPermission: NotificationPermission;
  serverTokenMissing: boolean;
  loading: boolean;
}

const DashboardHeader: React.FC<DashboardHeaderProps> = ({
  group,
  onBack,
  onEnableNotifications,
  onAddTask,
  notifPermission,
  serverTokenMissing,
  loading
}) => {
  const copyInviteCode = () => {
    if (group?.invite_code) {
      navigator.clipboard.writeText(group.invite_code);
      alert("招待コードをコピーしました！");
    }
  };

  // メンバー名をカンマ区切りで生成する（最大 3 名まで表示，それ以上は「ほか X 名」）
  const memberNames = group?.users?.map(u => u.name) || [];
  const displayMembers = memberNames.length > 3 
    ? `${memberNames.slice(0, 3).join(', ')} ほか ${memberNames.length - 3} 名`
    : memberNames.join(', ');

  return (
    <header className="dashboard-header">
      <div className="header-left">
        <button onClick={onBack} className="back-button" title="部屋選択へ戻る">
          <ArrowLeft size={22} />
        </button>
        <div>
          <h1>{group?.name || "Uni-Steps"}</h1>
          <div className="group-info">
            {group?.invite_code && (
              <button onClick={copyInviteCode} className="invite-code-pill">
                <Share2 size={12} /> {group.invite_code}
              </button>
            )}
            {group?.users && (
              <div style={{display: 'flex', alignItems: 'center', gap: '6px'}}>
                <Users size={14} color="var(--text-tertiary)" />
                <span style={{fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-secondary)'}}>{displayMembers}</span>
              </div>
            )}
          </div>
        </div>
      </div>
      <div className="header-actions">
        {!loading && (
          <>
            {(notifPermission !== 'granted' || serverTokenMissing) && (
              <button onClick={onEnableNotifications} className="btn btn-primary" style={{background: 'var(--warning)'}} title="通知を有効化">
                <BellRing size={20} />
                <span className="hide-mobile">通知を有効化</span>
              </button>
            )}
            <button onClick={onAddTask} className="btn btn-primary" title="手動で課題を登録">
              <Plus size={22} />
              <span className="hide-mobile">課題登録</span>
            </button>
          </>
        )}
      </div>
    </header>
  );
};

export default DashboardHeader;
