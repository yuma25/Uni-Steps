import React, { useEffect, useState } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { groupApi } from '../api/groups';
import type { Group } from '../types';
import { Plus, Layout } from 'lucide-react';

/**
 * グループ（部屋）選択画面のコンポーネントである．
 * ログイン直後に表示され，既存の部屋に入るか，新しい部屋を作成するかを選択する．
 */
const GroupSelectionPage: React.FC = () => {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const userId = searchParams.get('user_id') || '';

  const [groups, setGroups] = useState<Group[]>([]);
  const [newGroupName, setNewGroupName] = useState('');
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
      // プロトタイプ用のモックデータ
      setGroups([{ id: 'default-group-id', name: 'デフォルトグループ', owner_id: 'system' }]);
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
      alert("新しい部屋を作成した．");
    } catch (err) {
      alert("部屋の作成に失敗した．");
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const selectGroup = (groupId: string) => {
    navigate(`/dashboard?user_id=${userId}&group_id=${groupId}`);
  };

  return (
    <div className="selection-container">
      <h1>部屋を選択する</h1>
      
      <section className="group-list-section">
        <h2>参加中の部屋</h2>
        {loading ? (
          <p>読み込み中...</p>
        ) : groups.length === 0 ? (
          <p>所属している部屋はない．</p>
        ) : (
          <div className="group-grid">
            {groups.map(group => (
              <button key={group.id} onClick={() => selectGroup(group.id)} className="group-card">
                <Layout size={32} />
                <span>{group.name}</span>
              </button>
            ))}
          </div>
        )}
      </section>

      <hr />

      <section className="create-group-section">
        <h2>新しく部屋を作る</h2>
        <form onSubmit={handleCreateGroup} className="create-group-form">
          <input 
            type="text" 
            placeholder="部屋の名前（例：大学のゼミ）" 
            value={newGroupName}
            onChange={(e) => setNewGroupName(e.target.value)}
            disabled={loading}
          />
          <button type="submit" disabled={loading || !newGroupName.trim()} className="icon-button primary">
            <Plus size={20} />
            作成
          </button>
        </form>
      </section>
    </div>
  );
};

export default GroupSelectionPage;
