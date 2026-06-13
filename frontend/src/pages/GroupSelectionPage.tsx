import React, { useEffect, useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { groupApi } from '../api/groups';
import type { Group } from '../types';
import { Plus, Layout, UserPlus, LogOut, Hash, X } from 'lucide-react';
import { AxiosError } from 'axios';
import { useAuth } from '../hooks/useAuth';

/**
 * グループ（部屋）選択画面のコンポーネントである．
 */
const GroupSelectionPage: React.FC = () => {
  const navigate = useNavigate();
  const { userId } = useAuth();

  const [groups, setGroups] = useState<Group[]>([]);
  const [newGroupName, setNewGroupName] = useState('');
  const [inviteCode, setInviteCode] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showJoinModal, setShowJoinModal] = useState(false);
  const [loading, setLoading] = useState(true);

  const fetchGroups = useCallback(async () => {
    try {
      // 初期状態が loading: true のため，ここでは同期的な設定を避けて警告を回避する
      const data = await groupApi.listMyGroups(userId);
      setGroups(data || []);
    } catch (err) {
      console.error("GroupSelectionPage: fetchGroups failed", err);
    } finally {
      setLoading(false);
    }
  }, [userId]);

  useEffect(() => {
    if (!userId) {
      navigate('/login');
      return;
    }
    fetchGroups();
  }, [userId, navigate, fetchGroups]);

  const handleCreateGroup = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newGroupName.trim()) return;
    try {
      setLoading(true);
      const newGroup = await groupApi.createGroup(newGroupName, userId);
      setGroups([...groups, newGroup]);
      setShowCreateModal(false);
      navigate(`/dashboard?user_id=${userId}&group_id=${newGroup.id}`);
    } catch (err: unknown) {
      const axiosErr = err as AxiosError<{error: string}>;
      alert(`エラー：${axiosErr.response?.data?.error || "部屋の作成に失敗した．"}`);
    } finally {
      setLoading(false);
    }
  };

  const handleJoinGroup = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!inviteCode.trim()) return;
    try {
      setLoading(true);
      const joinedGroup = await groupApi.joinGroup(inviteCode, userId);
      setShowJoinModal(false);
      navigate(`/dashboard?user_id=${userId}&group_id=${joinedGroup.id}`);
    } catch (err: unknown) {
      const axiosErr = err as AxiosError<{error: string}>;
      alert(`エラー：${axiosErr.response?.data?.error || "部屋への参加に失敗した．招待コードを確認してほしい．"}`);
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
        
        <button onClick={() => setShowJoinModal(true)} className="action-card join">
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
            <p>まだ所属している部屋がありません．</p>
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
            <button onClick={() => setShowCreateModal(false)} className="close-button"><X size={24} /></button>
            <div className="modal-header-text">
              <h2>新しい部屋を作成</h2>
              <p>部屋の名前を決めてください．</p>
            </div>
            <form onSubmit={handleCreateGroup} className="create-group-form">
              <div className="form-group">
                <label>部屋の名前</label>
                <input type="text" placeholder="例：ゼミ用，月曜2限" value={newGroupName} onChange={(e) => setNewGroupName(e.target.value)} autoFocus required />
              </div>
              <button type="submit" disabled={loading || !newGroupName.trim()} className="icon-button primary full-width">
                {loading ? "作成中..." : "部屋を作成して移動する"}
              </button>
            </form>
          </div>
        </div>
      )}

      {/* 部屋参加モーダル */}
      {showJoinModal && (
        <div className="modal-overlay" onClick={() => setShowJoinModal(false)}>
          <div className="modal-content animate-pop" onClick={e => e.stopPropagation()}>
            <button onClick={() => setShowJoinModal(false)} className="close-button"><X size={24} /></button>
            <div className="modal-header-text">
              <h2>招待コードで参加</h2>
              <p>友達から共有された 8 桁のコードを入力してください．</p>
            </div>
            <form onSubmit={handleJoinGroup} className="create-group-form">
              <div className="form-group">
                <label>招待コード</label>
                <input type="text" placeholder="例：a1b2c3d4" value={inviteCode} onChange={(e) => setInviteCode(e.target.value)} autoFocus required />
              </div>
              <button type="submit" disabled={loading || !inviteCode.trim()} className="icon-button primary full-width" style={{background: 'var(--success)'}}>
                {loading ? "参加処理中..." : "部屋に参加する"}
              </button>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default GroupSelectionPage;
