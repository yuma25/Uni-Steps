import React, { useEffect, useState } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { groupApi } from '../api/groups';
import type { Group } from '../types';
import { Plus, Layout, RefreshCw } from 'lucide-react';

/**
 * グループ（部屋）選択画面のコンポーネントである．
 * ログイン直後に表示され，既存の部屋に入るか，新しい部屋を作成するかを選択する．
 * Google Classroom からの自動インポート機能も備えている．
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
      // バックエンドの DB に保存されている参加済みグループを取得する．
      const data = await groupApi.listMyGroups(userId);
      setGroups(data);
    } catch (err) {
      console.error("グループ一覧の取得に失敗した：", err);
    } finally {
      setLoading(false);
    }
  };

  const handleSyncLMS = async () => {
    try {
      setLoading(true);
      const data = await groupApi.syncLMSGroups(userId);
      setGroups(data);
      alert("Google Classroom から授業一覧を取得した．");
    } catch (err) {
      console.error("LMS 同期に失敗した：", err);
      alert("同期に失敗した．");
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
      <div className="header-with-action">
        <h1>部屋を選択する</h1>
        <button onClick={handleSyncLMS} disabled={loading} className="icon-button">
          <RefreshCw className={loading ? "animate-spin" : ""} size={16} />
          LMS同期
        </button>
      </div>
      
      <section className="group-list-section">
        <h2>参加中の部屋</h2>
        {loading ? (
          <p>読み込み中...</p>
        ) : groups.length === 0 ? (
          <div className="empty-state">
            <p>所属している部屋はない．上の「LMS同期」から Google Classroom の授業を取り込むか，新しい部屋を作ってほしい．</p>
          </div>
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
