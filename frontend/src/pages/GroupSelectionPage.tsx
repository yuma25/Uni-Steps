import React, { useEffect, useState } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { groupApi } from '../api/groups';
import type { Group } from '../types';
import { Plus, Layout,  UserPlus } from 'lucide-react';

/**
 * グループ（部屋）選択画面のコンポーネントである．
 * ログイン直後に表示され，友達と共有する「Uni-Steps 部屋」を作成するか，参加するかを選択する．
 */
const GroupSelectionPage: React.FC = () => {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const userId = searchParams.get('user_id') || '';

  const [groups, setGroups] = useState<Group[]>([]);
  const [newGroupName, setNewGroupName] = useState('');
  const [showCreateForm, setShowCreateForm] = useState(false);
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
      setGroups(data);
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
      setShowCreateForm(false);
      alert("新しい部屋を作成した．");
    } catch (err: any) {
      const errMsg = err.response?.data?.error || "部屋の作成に失敗した．";
      alert(`エラー：${errMsg}`);
      console.error("部屋の作成に失敗した：", err);
    } finally {
      setLoading(false);
    }
  };

  const selectGroup = (groupId: string) => {
    navigate(`/dashboard?user_id=${userId}&group_id=${groupId}`);
  };

  return (
    <div className="selection-container">
      <h1>Uni-Steps へようこそ</h1>
      <p className="subtitle">友達と一緒に課題を管理するための「部屋」を選んでほしい．</p>
      
      <div className="main-actions">
        <button onClick={() => setShowCreateForm(!showCreateForm)} className="action-card create">
          <Plus size={40} />
          <span>部屋を作成する</span>
        </button>
        <button onClick={() => alert("参加機能は準備中である．招待コードを入力する予定である．")} className="action-card join">
          <UserPlus size={40} />
          <span>部屋に参加する</span>
        </button>
      </div>

      {showCreateForm && (
        <form onSubmit={handleCreateGroup} className="create-group-inline-form">
          <input 
            type="text" 
            placeholder="新しい部屋の名前（例：サークル用）" 
            value={newGroupName}
            onChange={(e) => setNewGroupName(e.target.value)}
            autoFocus
          />
          <button type="submit" disabled={loading || !newGroupName.trim()}>作成</button>
        </form>
      )}

      <section className="group-list-section">
        <h2>参加中の部屋</h2>
        {loading ? (
          <p>読み込み中...</p>
        ) : groups.length === 0 ? (
          <div className="empty-state">
            <p>まだ所属している部屋はない．左のボタンから最初の部屋を作ってみてほしい．</p>
          </div>
        ) : (
          <div className="group-grid">
            {groups.map(group => (
              <button key={group.id} onClick={() => selectGroup(group.id)} className="group-card">
                <Layout size={32} />
                <span className="group-name">{group.name}</span>
                {group.owner_id === userId && <span className="owner-badge">Owner</span>}
              </button>
            ))}
          </div>
        )}
      </section>
    </div>
  );
};

export default GroupSelectionPage;
