import React, { useEffect, useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { groupApi } from '../api/groups';
import type { Group } from '../types';
import { Plus, Layout, UserPlus, LogOut, Hash, X, ArrowRight, Sparkles } from 'lucide-react';
import { AxiosError } from 'axios';
import { useAuth } from '../hooks/useAuth';

/**
 * グループ（部屋）選択画面（ポータル）のコンポーネントである．
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

  // モーダル表示時に背景のスクロールを禁止する．
  useEffect(() => {
    if (showCreateModal || showJoinModal) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = 'unset';
    }
    return () => { document.body.style.overflow = 'unset'; };
  }, [showCreateModal, showJoinModal]);

  return (
    <div className="portal-container">
      <div className="portal-content">
        <header className="selection-hero">
          <h1>Uni-Steps</h1>
          <p>
            あなたの「一歩」を支える部屋を選びましょう．<br />
            仲間と一緒に，理想のライフリズムを刻み始めましょう．
          </p>
        </header>

        <div className="action-cards-grid">
          <button onClick={() => setShowCreateModal(true)} className="portal-card">
            <div className="icon-box" style={{background: 'var(--brand-soft)', color: 'var(--brand)'}}>
              <Plus size={32} strokeWidth={3} />
            </div>
            <h3 style={{fontSize: '1.5rem', fontWeight: 900, margin: '0 0 1rem 0'}}>新しい部屋を作成</h3>
            <p style={{color: 'var(--text-secondary)', fontSize: '0.95rem', margin: 0, lineHeight: 1.6}}>
              新しく部屋を作り，友達を招待して<br />共同管理を始めましょう．
            </p>
          </button>

          <button onClick={() => setShowJoinModal(true)} className="portal-card">
            <div className="icon-box" style={{background: '#dcfce7', color: 'var(--success)'}}>
              <UserPlus size={32} strokeWidth={3} />
            </div>
            <h3 style={{fontSize: '1.5rem', fontWeight: 900, margin: '0 0 1rem 0'}}>招待コードで参加</h3>
            <p style={{color: 'var(--text-secondary)', fontSize: '0.95rem', margin: 0, lineHeight: 1.6}}>
              共有されたコードを入力し，<br />既存のチームに合流しましょう．
            </p>
          </button>
        </div>

        <section className="group-list-section">
          <div style={{display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '12px', marginBottom: '3rem'}}>
            <Layout size={24} color="var(--brand)" />
            <h2 style={{margin: 0, fontSize: '1.75rem', fontWeight: 950}}>参加中の部屋</h2>
          </div>

          {loading ? (
            <div style={{textAlign: 'center', padding: '4rem', color: 'var(--text-tertiary)'}}>
              <Sparkles className="loading-logo" style={{marginBottom: '1rem'}} />
              <p>データを読み込んでいます...</p>
            </div>
          ) : groups.length === 0 ? (
            <div style={{textAlign: 'center', padding: '6rem 2rem', background: 'rgba(255,255,255,0.4)', borderRadius: '32px', border: '1px dashed #cbd5e1'}}>
              <p style={{color: 'var(--text-secondary)', fontWeight: 600}}>まだ所属している部屋がありません．</p>
            </div>
          ) : (
            <div style={{display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))', gap: '1.25rem'}}>
              {groups.map(g => (
                <button key={g.id} onClick={() => selectGroup(g.id)} className="group-item-card">
                  <div className="group-item-icon">
                    <Hash size={24} strokeWidth={2.5} />
                  </div>
                  <div style={{flex: 1, textAlign: 'left'}}>
                    <div style={{fontWeight: 900, fontSize: '1.15rem', color: 'var(--text-primary)', marginBottom: '2px'}}>{g.name}</div>
                    {g.owner_id === userId ? (
                      <span style={{fontSize: '0.7rem', fontWeight: 900, color: 'var(--brand)', textTransform: 'uppercase', letterSpacing: '0.05em'}}>Room Owner</span>
                    ) : (
                      <span style={{fontSize: '0.7rem', fontWeight: 700, color: 'var(--text-tertiary)'}}>Member</span>
                    )}
                  </div>
                  <ArrowRight size={18} color="var(--text-tertiary)" />
                </button>
              ))}
            </div>
          )}
        </section>

        <footer className="portal-footer">
          <button onClick={() => navigate('/login')} className="btn btn-ghost" style={{padding: '12px 32px', borderRadius: '16px', fontSize: '0.9rem'}}>
            <LogOut size={18} />
            <span>ログアウト</span>
          </button>
          <div style={{marginTop: '2.5rem', display: 'flex', flexDirection: 'column', gap: '8px'}}>
            <div style={{fontSize: '1.25rem', fontWeight: 950, color: 'var(--text-tertiary)', letterSpacing: '-0.02em'}}>Uni-Steps</div>
            <div style={{fontSize: '0.8rem', color: 'var(--text-tertiary)', opacity: 0.7}}>© 2026 Your life rhythm partner.</div>
          </div>
        </footer>
      </div>

      {/* 部屋作成モーダル */}
      {showCreateModal && (
        <div className="modal-overlay" onClick={() => setShowCreateModal(false)}>
          <div className="modal-content animate-pop" onClick={e => e.stopPropagation()}>
            <button onClick={() => setShowCreateModal(false)} className="btn-ghost" style={{position: 'absolute', top: '1.5rem', right: '1.5rem', padding: '8px', borderRadius: '50%', border: 'none'}}><X size={20} /></button>
            <div className="modal-header-text" style={{marginBottom: '2.5rem'}}>
              <h2 style={{fontWeight: 900}}>新しい部屋を作成</h2>
              <p>部屋の名前を決めてください．後から変更も可能です．</p>
            </div>
            <form onSubmit={handleCreateGroup}>
              <div className="form-group">
                <label><Hash size={14} /> 部屋の名前</label>
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
            <div className="modal-header-text" style={{marginBottom: '2.5rem'}}>
              <h2 style={{fontWeight: 900}}>招待コードで参加</h2>
              <p>友達から共有された 8 桁のコードを入力してください．</p>
            </div>
            <form onSubmit={handleJoinGroup}>
              <div className="form-group">
                <label><UserPlus size={14} /> 招待コード</label>
                <input type="text" placeholder="例：a1b2c3d4" value={inviteCode} onChange={(e) => setInviteCode(e.target.value)} autoFocus required />
              </div>
              <button type="submit" disabled={loading || !inviteCode.trim()} className="btn btn-primary" style={{width: '100%', padding: '16px', borderRadius: '16px', background: 'var(--success)'}}>
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
