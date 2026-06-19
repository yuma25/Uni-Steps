import React from 'react';
import { ArrowLeft, Users, Share2, Plus, RefreshCw } from 'lucide-react';
import type { Group } from '../../types';

interface DashboardHeaderProps {
  group: Group | null;
  onBack: () => void;
  onAddTask: () => void;
  onSync: () => void;
  loading: boolean;
  isSyncing?: boolean;
  lastSyncedAt?: string;
}

const DashboardHeader: React.FC<DashboardHeaderProps> = ({
  group,
  onBack,
  onAddTask,
  onSync,
  loading,
  isSyncing = false,
  lastSyncedAt
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
      <div className="header-top-nav">
        <button onClick={onBack} className="back-button" title="部屋選択へ戻る">
          <ArrowLeft size={22} />
        </button>
        <div className="header-actions">
          {!loading && (
            <>
              <button onClick={onSync} className="btn btn-ghost" disabled={isSyncing} title="Google Classroom から課題を同期" style={{display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '2px', padding: '8px 12px', minWidth: '80px'}}>
                <div style={{display: 'flex', alignItems: 'center', gap: '6px'}}>
                  <RefreshCw size={18} className={isSyncing ? "animate-spin" : ""} />
                  <span className="hide-mobile" style={{fontWeight: 800}}>{isSyncing ? "同期中..." : "同期"}</span>
                </div>
                {lastSyncedAt && !isSyncing && (
                  <span style={{fontSize: '0.6rem', color: 'var(--text-tertiary)', fontWeight: 700, whiteSpace: 'nowrap'}}>
                    最終 {new Date(lastSyncedAt).toLocaleTimeString('ja-JP', {hour: '2-digit', minute: '2-digit'})}
                  </span>
                )}
              </button>
              <button onClick={onAddTask} className="btn btn-primary" title="手動で課題を登録">
                <Plus size={22} />
                <span className="hide-mobile">課題登録</span>
              </button>
            </>
          )}
        </div>
      </div>
      
      <div className="header-title-section">
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
    </header>
  );
};

export default DashboardHeader;
