import React, { useEffect, useState } from 'react';
import { taskApi } from '../api/tasks';
import { Task } from '../types';
import { Bell, RefreshCw, PlusCircle } from 'lucide-react';

/**
 * ダッシュボード画面のコンポーネントである．
 * 課題一覧の表示，同期の実行，手動登録の開始を行う中心的な画面である．
 */
const DashboardPage: React.FC = () => {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // 仮の ID（本来はログインユーザーの情報から取得する）
  const tempGroupId = "default-group-id";
  const tempUserId = "default-user-id";

  // 画面表示時に課題一覧を取得する．
  useEffect(() => {
    fetchTasks();
  }, []);

  const fetchTasks = async () => {
    try {
      setLoading(true);
      const data = await taskApi.listGroupTasks(tempGroupId);
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
      await taskApi.syncTasks(tempUserId, tempGroupId);
      await fetchTasks(); // 同期後に再取得する．
      alert("同期が完了した．（更新がある場合のみ反映される）");
    } catch (err: any) {
      alert(err.response?.data?.error || "同期に失敗した．");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="dashboard-container">
      <header className="dashboard-header">
        <h1>Uni-Steps</h1>
        <div className="header-actions">
          <button onClick={handleSync} disabled={loading} className="icon-button">
            <RefreshCw className={loading ? "animate-spin" : ""} />
            同期
          </button>
          <button className="icon-button primary">
            <PlusCircle />
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
                    <span className="source-tag">{task.source}</span>
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
    </div>
  );
};

export default DashboardPage;
