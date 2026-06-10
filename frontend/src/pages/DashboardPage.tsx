import React, { useEffect, useState } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { taskApi } from '../api/tasks';
import type { Task } from '../types';
import { Bell, RefreshCw, PlusCircle, X } from 'lucide-react';

/**
 * ダッシュボード画面のコンポーネントである．
 * URL のクエリパラメータからユーザー ID とグループ ID を取得し，そのコンテキストでの課題管理を行う．
 */
const DashboardPage: React.FC = () => {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const userId = searchParams.get('user_id') || '';
  const groupId = searchParams.get('group_id') || '';
  
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showModal, setShowModal] = useState(false);

  // 手動登録用のフォーム状態である．
  const [formData, setFormData] = useState({
    title: '',
    deadline: '',
    is_critical: false,
    recurrence: 'none'
  });

  useEffect(() => {
    if (!userId || !groupId) {
      navigate('/login');
      return;
    }
    fetchTasks();
  }, [userId, groupId]);

  const fetchTasks = async () => {
    try {
      setLoading(true);
      const data = await taskApi.listGroupTasks(groupId);
      setTasks(data);
      setError(null);
    } catch (err) {
      setError("課題の取得に失敗した．");
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleSync = async () => {
    try {
      setLoading(true);
      await taskApi.syncTasks(userId, groupId);
      await fetchTasks();
      alert("同期が完了した．（更新がある場合のみ反映される）");
    } catch (err: any) {
      alert(err.response?.data?.error || "同期に失敗した．");
    } finally {
      setLoading(false);
    }
  };

  const handleCreateTask = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      setLoading(true);
      await taskApi.createManualTask({
        ...formData,
        group_id: groupId,
        user_id: userId,
        deadline: new Date(formData.deadline).toISOString(),
      });
      setShowModal(false);
      setFormData({ title: '', deadline: '', is_critical: false, recurrence: 'none' });
      await fetchTasks();
    } catch (err) {
      alert("課題の登録に失敗した．");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="dashboard-container">
      <header className="dashboard-header">
        <div className="title-group">
          <h1>Uni-Steps</h1>
          <span className="group-badge">Room: {groupId.substring(0, 8)}</span>
        </div>
        <div className="header-actions">
          <button onClick={handleSync} disabled={loading} className="icon-button">
            <RefreshCw className={loading ? "animate-spin" : ""} size={18} />
            同期
          </button>
          <button onClick={() => setShowModal(true)} className="icon-button primary">
            <PlusCircle size={18} />
            課題追加
          </button>
        </div>
      </header>

      <main className="dashboard-content">
        {error && <div className="error-message">{error}</div>}
        
        <section className="task-section">
          <h2>課題タイムライン</h2>
          {loading && tasks.length === 0 ? (
            <p>読み込み中...</p>
          ) : tasks.length === 0 ? (
            <div className="empty-state">
              <Bell size={48} />
              <p>現在，管理されている課題はない．</p>
            </div>
          ) : (
            <div className="task-list">
              {tasks.map(task => (
                <div key={task.id} className={`task-card ${task.is_critical ? 'critical' : ''}`}>
                  <div className="task-info">
                    <h3>{task.title}</h3>
                    <p className="deadline">期限: {new Date(task.deadline).toLocaleString('ja-JP')}</p>
                    <div className="tags">
                      <span className="source-tag">{task.source}</span>
                      {task.recurrence !== 'none' && <span className="recurrence-tag">{task.recurrence}</span>}
                    </div>
                  </div>
                  <div className="task-status">
                    {task.is_completed ? "✅ 完了" : "⏳ 未完了"}
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>
      </main>

      {/* 手動登録用モーダルである． */}
      {showModal && (
        <div className="modal-overlay">
          <div className="modal-content">
            <div className="modal-header">
              <h2>課題を手動で追加する</h2>
              <button onClick={() => setShowModal(false)} className="close-button"><X /></button>
            </div>
            <form onSubmit={handleCreateTask} className="task-form">
              <div className="form-group">
                <label>タイトル</label>
                <input 
                  type="text" 
                  required 
                  value={formData.title}
                  onChange={e => setFormData({...formData, title: e.target.value})}
                  placeholder="例：数学のレポート提出"
                />
              </div>
              <div className="form-group">
                <label>期限</label>
                <input 
                  type="datetime-local" 
                  required 
                  value={formData.deadline}
                  onChange={e => setFormData({...formData, deadline: e.target.value})}
                />
              </div>
              <div className="form-group inline">
                <input 
                  type="checkbox" 
                  id="is_critical"
                  checked={formData.is_critical}
                  onChange={e => setFormData({...formData, is_critical: e.target.checked})}
                />
                <label htmlFor="is_critical">起床確認を必須にする（重要）</label>
              </div>
              <div className="form-group">
                <label>繰り返し</label>
                <select 
                  value={formData.recurrence}
                  onChange={e => setFormData({...formData, recurrence: e.target.value})}
                >
                  <option value="none">なし</option>
                  <option value="weekly">毎週</option>
                  <option value="biweekly">隔週</option>
                </select>
              </div>
              <button type="submit" disabled={loading} className="submit-button">
                {loading ? "登録中..." : "登録する"}
              </button>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default DashboardPage;
