import React, { useEffect, useState } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { taskApi } from '../api/tasks';
import { groupApi } from '../api/groups';
import type { Task, Group } from '../types';
import { Bell, RefreshCw, PlusCircle, X, Link as LinkIcon, Settings } from 'lucide-react';

/**
 * ダッシュボード画面のコンポーネントである．
 * URL のクエリパラメータからユーザー ID とグループ ID を取得し，そのコンテキストでの課題管理を行う．
 * また，Google Classroom との紐付け（Link）機能を提供する．
 */
const DashboardPage: React.FC = () => {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const userId = searchParams.get('user_id') || '';
  const groupId = searchParams.get('group_id') || '';
  
  const [tasks, setTasks] = useState<Task[]>([]);
  const [group, setGroup] = useState<Group | null>(null);
  const [availableCourses, setAvailableCourses] = useState<Group[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showTaskModal, setShowTaskModal] = useState(false);
  const [showLinkModal, setShowLinkModal] = useState(false);

  const [taskFormData, setTaskFormData] = useState({
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
    fetchData();
  }, [userId, groupId]);

  const fetchData = async () => {
    try {
      setLoading(true);
      // 課題一覧とグループ情報の取得を並行して行う．
      const [taskData, groups] = await Promise.all([
        taskApi.listGroupTasks(groupId),
        groupApi.listMyGroups(userId)
      ]);
      setTasks(taskData);
      const currentGroup = groups.find(g => g.id === groupId);
      if (currentGroup) setGroup(currentGroup);
      setError(null);
    } catch (err) {
      setError("データの取得に失敗した．");
    } finally {
      setLoading(false);
    }
  };

  const handleSync = async () => {
    if (!group?.line_channel_token && !confirm("LINE 連携が未設定である．通知は送信されないが，同期を続行するか？")) return;
    
    try {
      setLoading(true);
      const result = await taskApi.syncTasks(userId, groupId);
      if (result) {
        await fetchData();
        alert("同期が完了した．");
      } else {
        alert("更新された情報はなかった．");
      }
    } catch (err: any) {
      alert(err.response?.data?.error || "同期に失敗した．");
    } finally {
      setLoading(false);
    }
  };

  const handleOpenLinkModal = async () => {
    try {
      setLoading(true);
      const courses = await groupApi.fetchAvailableLMSCourses(userId);
      setAvailableCourses(courses);
      setShowLinkModal(true);
    } catch (err) {
      alert("コース一覧の取得に失敗した．");
    } finally {
      setLoading(false);
    }
  };

  const handleLinkCourse = async (lmsCourseId: string) => {
    try {
      setLoading(true);
      await groupApi.linkLMSCourse(groupId, lmsCourseId);
      setShowLinkModal(false);
      await fetchData();
      alert("LMS コースを紐付けた．");
    } catch (err) {
      alert("紐付けに失敗した．");
    } finally {
      setLoading(false);
    }
  };

  const handleCreateTask = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      setLoading(true);
      await taskApi.createManualTask({
        ...taskFormData,
        group_id: groupId,
        user_id: userId,
        deadline: new Date(taskFormData.deadline).toISOString(),
      });
      setShowTaskModal(false);
      setTaskFormData({ title: '', deadline: '', is_critical: false, recurrence: 'none' });
      await fetchData();
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
          <h1>{group?.name || "Uni-Steps"}</h1>
          <div className="group-info">
            <span className="group-badge">Room ID: {groupId.substring(0, 8)}</span>
            {group?.lms_course_id ? (
              <span className="lms-badge linked">LMS Linked</span>
            ) : (
              <button onClick={handleOpenLinkModal} className="text-button link-action">
                <LinkIcon size={14} /> Classroom 連携
              </button>
            )}
          </div>
        </div>
        <div className="header-actions">
          <button onClick={handleSync} disabled={loading || !group?.lms_course_id} className="icon-button">
            <RefreshCw className={loading ? "animate-spin" : ""} size={18} />
            同期
          </button>
          <button onClick={() => setShowTaskModal(true)} className="icon-button primary">
            <PlusCircle size={18} />
            課題追加
          </button>
          <button className="icon-button" title="設定"><Settings size={18} /></button>
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
                      <span className={`source-tag ${task.source}`}>{task.source}</span>
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

      {/* 課題登録モーダル */}
      {showTaskModal && (
        <div className="modal-overlay">
          <div className="modal-content">
            <div className="modal-header">
              <h2>課題を手動で追加する</h2>
              <button onClick={() => setShowTaskModal(false)} className="close-button"><X /></button>
            </div>
            <form onSubmit={handleCreateTask} className="task-form">
              <div className="form-group">
                <label>タイトル</label>
                <input 
                  type="text" required value={taskFormData.title}
                  onChange={e => setTaskFormData({...taskFormData, title: e.target.value})}
                  placeholder="例：数学のレポート提出"
                />
              </div>
              <div className="form-group">
                <label>期限</label>
                <input 
                  type="datetime-local" required value={taskFormData.deadline}
                  onChange={e => setTaskFormData({...taskFormData, deadline: e.target.value})}
                />
              </div>
              <div className="form-group inline">
                <input 
                  type="checkbox" id="is_critical_modal"
                  checked={taskFormData.is_critical}
                  onChange={e => setTaskFormData({...taskFormData, is_critical: e.target.checked})}
                />
                <label htmlFor="is_critical_modal">起床確認を必須にする（重要）</label>
              </div>
              <button type="submit" className="submit-button">登録する</button>
            </form>
          </div>
        </div>
      )}

      {/* LMS 連携モーダル */}
      {showLinkModal && (
        <div className="modal-overlay">
          <div className="modal-content">
            <div className="modal-header">
              <h2>Google Classroom を連携</h2>
              <button onClick={() => setShowLinkModal(false)} className="close-button"><X /></button>
            </div>
            <p>この部屋に紐付ける Google Classroom の授業（コース）を選択してほしい．</p>
            <div className="course-list">
              {availableCourses.length === 0 ? (
                <p>利用可能なアクティブなコースは見つからなかった．</p>
              ) : (
                availableCourses.map(course => (
                  <button key={course.id} onClick={() => handleLinkCourse(course.id)} className="course-item">
                    <span>{course.name}</span>
                    <Plus size={16} />
                  </button>
                ))
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default DashboardPage;
