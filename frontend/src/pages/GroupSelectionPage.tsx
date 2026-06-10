import React, { useEffect, useState } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { groupApi } from '../api/groups';
import type { Group } from '../types';
import { Plus, Layout, UserPlus, LogOut, Hash, X } from 'lucide-react';

/**
 * グループ（部屋）選択画面のコンポーネントである．
 * UIUXを大幅に改善し，シンプルで迷わない操作感を実現した．
 */
const GroupSelectionPage: React.FC = () => {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const userId = searchParams.get('user_id') || '';

  const [groups, setGroups] = useState<Group[]>([]);
  const [newGroupName, setNewGroupName] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!userId) {
      navigate('/login');
      return;
    }
    fetchGroups();
  }, [userId]);

  const fetchGroups = async () => {
    try {
      setLoading(true);
      const data = await groupApi.listMyGroups(userId);
      setGroups(data || []);
    } catch (err) {
      console.error("グループ一覧の取得に失敗した：", err);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateGroup = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newGroupName.trim()) return;

    try {
      setLoading(true);
      const newGroup = await groupApi.createGroup(newGroupName, userId);
      setGroups([...groups, newGroup]);
      setNewGroupName('');
      setShowCreateModal(false);
      // 作成後，自動的にその部屋へ遷移する
      navigate(`/dashboard?user_id=${userId}&group_id=${newGroup.id}`);
    } catch (err: any) {
      const errMsg = err.response?.data?.error || "部屋の作成に失敗した．";
      alert(`エラー：${errMsg}`);
    } finally {
      setLoading(false);
    }
  };

  const selectGroup = (groupId: string) => {
    navigate(`/dashboard?user_id=${userId}&group_id=${groupId}`);
  };

  return (
    <div className="selection-container">
      <header className="selection-header">
        <div className="welcome-text">
          <h1>Uni-Steps</h1>
          <p className="subtitle">あなたの「一歩」を支える部屋を選びましょう．</p>
        </div>
        <button onClick={() => navigate('/login')} className="logout-button" title="ログアウト">
          <LogOut size={18} />
          <span>ログアウト</span>
        </button>
      </header>
      
      <div className="main-actions">
        <button onClick={() => setShowCreateModal(true)} className="action-card create">
          <div className="action-icon-circle">
            <Plus size={32} />
          </div>
          <div className="action-content">
            <h3>新しい部屋を作成</h3>
            <p>友達を招待して，一緒に課題を管理し始めましょう．</p>
          </div>
        </button>
        
        <button onClick={() => alert("参加機能は準備中である．招待コードを入力する予定である．")} className="action-card join">
          <div className="action-icon-circle secondary">
            <UserPlus size={32} />
          </div>
          <div className="action-content">
            <h3>招待コードで参加</h3>
            <p>既存の部屋に参加して，仲間のサポートを受けましょう．</p>
          </div>
        </button>
      </div>

      <section className="group-list-section">
        <div className="section-title">
          <Layout size={20} />
          <h2>参加中の部屋</h2>
        </div>
        
        {loading ? (
          <div className="loading-state">読み込み中...</div>
        ) : groups.length === 0 ? (
          <div className="empty-state-card">
            <p>まだ所属している部屋がありません．<br />「新しい部屋を作成」から最初の一歩を踏み出しましょう．</p>
          </div>
        ) : (
          <div className="group-grid">
            {groups.map(group => (
              <button key={group.id} onClick={() => selectGroup(group.id)} className="group-card">
                <div className="group-info-main">
                  <div className="group-avatar">
                    <Hash size={24} />
                  </div>
                  <span className="group-name">{group.name}</span>
                </div>
                {group.owner_id === userId && <span className="owner-badge">OWNER</span>}
              </button>
            ))}
          </div>
        )}
      </section>

      {/* 部屋作成モーダル */}
      {showCreateModal && (
        <div className="modal-overlay" onClick={() => setShowCreateModal(false)}>
          <div className="modal-content animate-pop" onClick={e => e.stopPropagation()}>
            <button onClick={() => setShowCreateModal(false)} className="close-button">
              <X size={24} />
            </button>
            <div className="modal-header-text">
              <h2>新しい部屋を作成</h2>
              <p>部屋の名前を決めてください．後から変更も可能です．</p>
            </div>
            <form onSubmit={handleCreateGroup} className="create-group-form">
              <div className="form-group">
                <label>部屋の名前</label>
                <input 
                  type="text" 
                  placeholder="例：ゼミ用，サークル，月曜2限" 
                  value={newGroupName}
                  onChange={(e) => setNewGroupName(e.target.value)}
                  autoFocus
                  required
                />
              </div>
              <button type="submit" disabled={loading || !newGroupName.trim()} className="icon-button primary full-width">
                {loading ? "作成中..." : "部屋を作成して移動する"}
              </button>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default GroupSelectionPage;
