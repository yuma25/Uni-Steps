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
    <div className="app-shell" style={{background: 'var(--bg-app)', minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center'}}>
      <main className="app-main" style={{maxWidth: '1000px', padding: '4rem 2rem', display: 'flex', flexDirection: 'column', alignItems: 'center', textAlign: 'center'}}>
        
        {/* 中央寄せのヒーローセクション */}
        <header className="selection-hero" style={{width: '100%', display: 'flex', flexDirection: 'column', alignItems: 'center', padding: '5rem 3rem'}}>
          <h1>Uni-Steps</h1>
          <p>あなたの「一歩」を支える部屋を選びましょう．<br />仲間と一緒に，理想のライフリズムを刻み始めましょう．</p>
        </header>
        
        <div style={{display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))', gap: '2rem', marginBottom: '5rem', width: '100%'}}>
          <button onClick={() => setShowCreateModal(true)} className="task-card" style={{cursor: 'pointer', border: '2px solid transparent', padding: '3rem 2rem', display: 'flex', flexDirection: 'column', alignItems: 'center', textAlign: 'center'}}>
            <div style={{width: '64px', height: '64px', borderRadius: '20px', background: 'var(--text-primary)', color: 'white', display: 'flex', alignItems: 'center', justifyContent: 'center', marginBottom: '1.5rem'}}>
              <Plus size={36} />
            </div>
            <h3 style={{margin: 0, fontSize: '1.5rem', fontWeight: 900}}>新しい部屋を作成</h3>
            <p style={{color: 'var(--text-secondary)', fontSize: '1rem', marginTop: '0.8rem'}}>新しく部屋を作り，友達を招待して管理を始めましょう．</p>
          </button>
          
          <button onClick={() => setShowJoinModal(true)} className="task-card" style={{cursor: 'pointer', border: '2px solid transparent', padding: '3rem 2rem', display: 'flex', flexDirection: 'column', alignItems: 'center', textAlign: 'center'}}>
            <div style={{width: '64px', height: '64px', borderRadius: '20px', background: 'var(--brand-soft)', color: 'var(--brand)', display: 'flex', alignItems: 'center', justifyContent: 'center', marginBottom: '1.5rem'}}>
              <UserPlus size={36} />
            </div>
            <h3 style={{margin: 0, fontSize: '1.5rem', fontWeight: 900}}>招待コードで参加</h3>
            <p style={{color: 'var(--text-secondary)', fontSize: '1rem', marginTop: '0.8rem'}}>共有されたコードを入力し，既存のチームに合流しましょう．</p>
          </button>
        </div>

        <section style={{width: '100%', marginBottom: '6rem'}}>
          <div style={{display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '12px', marginBottom: '2.5rem'}}>
            <Layout size={28} color="var(--brand)" />
            <h2 style={{margin: 0, fontSize: '1.8rem', fontWeight: 950}}>参加中の部屋</h2>
          </div>
          
          {loading ? (
            <div className="loading-state">読み込み中...</div>
          ) : groups.length === 0 ? (
            <div className="empty-state" style={{padding: '5rem'}}>
              <p>まだ所属している部屋がありません．上のボタンから作成または参加してください．</p>
            </div>
          ) : (
            <div style={{display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))', gap: '1.5rem', justifyContent: 'center'}}>
              {groups.map(g => (
                <button key={g.id} onClick={() => selectGroup(g.id)} className="task-card" style={{flexDirection: 'row', alignItems: 'center', gap: '1.5rem', cursor: 'pointer', padding: '1.75rem'}}>
                  <div style={{width: '56px', height: '56px', borderRadius: '16px', background: '#f1f5f9', display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--text-secondary)'}}>
                    <Hash size={28} />
                  </div>
                  <div style={{flex: 1, textAlign: 'left'}}>
                    <div style={{fontWeight: 900, fontSize: '1.25rem', color: 'var(--text-primary)'}}>{g.name}</div>
                    {g.owner_id === userId && <span style={{fontSize: '0.75rem', fontWeight: 900, color: 'var(--brand)', textTransform: 'uppercase', letterSpacing: '0.08em'}}>Room Owner</span>}
                  </div>
                </button>
              ))}
            </div>
          )}
        </section>

        {/* ログアウトボタンを一番下に配置 */}
        <footer style={{marginTop: 'auto', padding: '2rem 0', width: '100%', borderTop: '1px solid #e5e7eb'}}>
          <button onClick={() => navigate('/login')} className="btn btn-ghost" style={{padding: '12px 24px', borderRadius: '14px', border: '1px solid #e5e7eb'}}>
            <LogOut size={20} />
            <span>ログアウト</span>
          </button>
          <p style={{marginTop: '1.5rem', fontSize: '0.85rem', color: 'var(--text-tertiary)'}}>© 2026 Uni-Steps. Your life rhythm partner.</p>
        </footer>

        {/* 部屋作成モーダル */}
        {showCreateModal && (
          <div className="modal-overlay" onClick={() => setShowCreateModal(false)}>
            <div className="modal-content animate-pop" onClick={e => e.stopPropagation()}>
              <button onClick={() => setShowCreateModal(false)} className="btn-ghost" style={{position: 'absolute', top: '1.5rem', right: '1.5rem', padding: '8px', borderRadius: '50%', border: 'none'}}><X size={20} /></button>
              <h2 style={{marginTop: 0, fontWeight: 900}}>新しい部屋を作成</h2>
              <p style={{color: 'var(--text-secondary)', marginBottom: '2.5rem'}}>部屋の名前を決めてください．後から変更も可能です．</p>
              <form onSubmit={handleCreateGroup}>
                <div className="form-group">
                  <label>部屋の名前</label>
                  <input type="text" placeholder="例：ゼミ用，月曜2限" value={newGroupName} onChange={(e) => setNewGroupName(e.target.value)} autoFocus required />
                </div>
                <button type="submit" disabled={loading || !newGroupName.trim()} className="btn btn-primary" style={{width: '100%', padding: '16px', borderRadius: '16px'}}>
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
              <button onClick={() => setShowJoinModal(false)} className="btn-ghost" style={{position: 'absolute', top: '1.5rem', right: '1.5rem', padding: '8px', borderRadius: '50%', border: 'none'}}><X size={20} /></button>
              <h2 style={{marginTop: 0, fontWeight: 900}}>招待コードで参加</h2>
              <p style={{color: 'var(--text-secondary)', marginBottom: '2.5rem'}}>友達から共有された 8 桁のコードを入力してください．</p>
              <form onSubmit={handleJoinGroup}>
                <div className="form-group">
                  <label>招待コード</label>
                  <input type="text" placeholder="例：a1b2c3d4" value={inviteCode} onChange={(e) => setInviteCode(e.target.value)} autoFocus required />
                </div>
                <button type="submit" disabled={loading || !inviteCode.trim()} className="btn btn-primary" style={{width: '100%', padding: '16px', borderRadius: '16px', background: 'var(--success)'}}>
                  {loading ? "参加処理中..." : "部屋に参加する"}
                </button>
              </form>
            </div>
          </div>
        )}
      </main>
    </div>
  );
};

export default GroupSelectionPage;
