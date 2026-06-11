import React, { useEffect, useState } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { taskApi } from '../api/tasks';
import { groupApi } from '../api/groups';
import type { Task, Group } from '../types';
import { Bell, RefreshCw, PlusCircle, X, Settings, Plus, Archive, Edit, ChevronDown, ArrowLeft, Users, Calendar, CheckCircle, Clock, Copy, Share2 } from 'lucide-react';

/**
 * ダッシュボード画面のコンポーネントである．
 */
const DashboardPage: React.FC = () => {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const userId = searchParams.get('user_id') || '';
  const groupId = searchParams.get('group_id') || '';
  
  const [tasks, setTasks] = useState<Task[]>([]);
  const [group, setGroup] = useState<Group | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showTaskModal, setShowTaskModal] = useState(false);
  const [editingTask, setEditingTask] = useState<Task | null>(null);
  const [showArchive, setShowArchive] = useState(false);

  const [taskFormData, setTaskFormData] = useState({
    title: '',
    deadline: '',
    recurrence_type: 'none',
    assignees: [] as string[]
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
      const [taskData, groups] = await Promise.all([
        taskApi.listGroupTasks(groupId),
        groupApi.listMyGroups(userId)
      ]);
      setTasks(taskData || []);
      if (groups && Array.isArray(groups)) {
        const currentGroup = groups.find(g => g.id === groupId);
        if (currentGroup) setGroup(currentGroup);
      }
      setError(null);
    } catch (err) {
      setError("データの取得に失敗した．");
    } finally {
      setLoading(false);
    }
  };

  const handleSync = async () => {
    try {
      setLoading(true);
      const result = await taskApi.syncTasks(userId, groupId);
      if (result && result.tasks && result.tasks.length > 0) {
        await fetchData();
        alert(`同期が完了した．${result.tasks.length} 件の課題を更新した．`);
      } else {
        alert("更新された情報はなかった．");
      }
    } catch (err: any) {
      alert(err.response?.data?.error || "同期に失敗した．");
    } finally {
      setLoading(false);
    }
  };

  const handleOpenCreateModal = () => {
    setEditingTask(null);
    setTaskFormData({
      title: '',
      deadline: '',
      recurrence_type: 'none',
      assignees: [userId]
    });
    setShowTaskModal(true);
  };

  const handleOpenEditModal = (task: Task) => {
    setEditingTask(task);
    const date = new Date(task.deadline);
    const deadlineStr = date.getFullYear() <= 1 
      ? "" 
      : new Date(date.getTime() - date.getTimezoneOffset() * 60000).toISOString().slice(0, 16);
    
    setTaskFormData({
      title: task.title,
      deadline: deadlineStr,
      recurrence_type: task.recurrence?.type || 'none',
      assignees: task.user_progress?.map(p => p.user_id) || []
    });
    setShowTaskModal(true);
  };

  const handleSaveTask = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      setLoading(true);
      const deadline = taskFormData.deadline ? new Date(taskFormData.deadline).toISOString() : "0001-01-01T00:00:00Z";
      
      const userProgress = taskFormData.assignees.map(id => {
        const member = group?.users?.find(u => u.id === id);
        const existingProgress = editingTask?.user_progress?.find(p => p.user_id === id);
        return {
          task_id: editingTask?.id || "",
          user_id: id,
          user_name: member?.name || existingProgress?.user_name || "Unknown",
          is_completed: existingProgress?.is_completed || false,
          updated_at: existingProgress?.updated_at || new Date().toISOString()
        };
      });

      const taskData = {
        title: taskFormData.title,
        deadline: deadline,
        recurrence: { type: taskFormData.recurrence_type, custom_dates: [] },
        user_progress: userProgress as any
      };

      if (editingTask) {
        if (!editingTask.id) throw new Error("課題 ID が見つからないため更新できない．");
        await taskApi.updateTask(editingTask.id, taskData);
      } else {
        await taskApi.createManualTask({
          ...taskData,
          group_id: groupId,
        });
      }
      
      setShowTaskModal(false);
      await fetchData();
    } catch (err: any) {
      const errMsg = err.response?.data?.error || err.message || "不明なエラー";
      alert(`課題の保存に失敗した：${errMsg}`);
    } finally {
      setLoading(false);
    }
  };

  const handleToggleAssignee = (memberId: string) => {
    setTaskFormData(prev => {
      const isSelected = prev.assignees.includes(memberId);
      if (isSelected) {
        return { ...prev, assignees: prev.assignees.filter(id => id !== memberId) };
      } else {
        return { ...prev, assignees: [...prev.assignees, memberId] };
      }
    });
  };

  const handleToggleCompletion = async (taskId: string) => {
    try {
      setLoading(true);
      await taskApi.toggleTaskCompletion(taskId, userId);
      await fetchData();
    } catch (err) {
      alert("状態の更新に失敗した．");
    } finally {
      setLoading(false);
    }
  };

  const copyInviteCode = () => {
    if (group?.invite_code) {
      navigator.clipboard.writeText(group.invite_code);
      alert("招待コードをクリップボードにコピーしました！友達に共有しましょう．");
    }
  };

  const getTaskStatus = (task: Task) => {
    const myProgress = task.user_progress?.find(p => p.user_id === userId);
    const deadlineDate = new Date(task.deadline);
    const isUndetermined = deadlineDate.getFullYear() <= 1;
    const isPast = !isUndetermined && deadlineDate < new Date();
    
    if (myProgress?.is_completed) return { label: "完了", className: "status-completed" };
    if (isPast) return { label: "提出遅れ", className: "status-overdue" };
    return { label: "未完了", className: "status-pending" };
  };

  const activeTasks = tasks.filter(t => {
    const deadlineDate = new Date(t.deadline);
    return deadlineDate.getFullYear() <= 1 || deadlineDate >= new Date();
  });
  const archivedTasks = tasks
    .filter(t => {
      const deadlineDate = new Date(t.deadline);
      return deadlineDate.getFullYear() > 1 && deadlineDate < new Date();
    })
    .sort((a, b) => new Date(b.deadline).getTime() - new Date(a.deadline).getTime());

  const renderTaskCard = (task: Task) => {
    const myStatus = getTaskStatus(task);
    const deadlineDate = new Date(task.deadline);
    const isUndetermined = deadlineDate.getFullYear() <= 1;
    const canEdit = task.source !== 'google_classroom' || isUndetermined;

    return (
      <div key={task.id} className="task-card">
        <div className="task-info">
          <h3>{task.title}</h3>
          <p className="deadline">
            <Calendar size={13} />
            <span>{isUndetermined ? "期限未定" : deadlineDate.toLocaleString('ja-JP', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })}</span>
          </p>
          
          <div className="member-status-list">
            {task.user_progress?.map(p => (
              <span key={p.user_id} className={`member-badge ${p.is_completed ? 'completed' : ''}`}>
                {p.user_name || "User"}: {p.is_completed ? "完了" : "未"}
              </span>
            ))}
          </div>
          
          {task.recurrence?.type && task.recurrence.type !== 'none' && (
            <div className="tags" style={{marginTop: '0.5rem'}}>
              <span className="source-tag">リピート: {task.recurrence.type}</span>
            </div>
          )}
        </div>
        
        <div className="task-status">
          <span className={`status-badge ${myStatus.className}`}>{myStatus.label}</span>
          <div className="card-actions">
            {canEdit && (
              <button onClick={() => handleOpenEditModal(task)} className="edit-button">編集</button>
            )}
            {task.source !== 'google_classroom' && (
              <button onClick={() => handleToggleCompletion(task.id)} className="complete-toggle-button">
                {myStatus.label === "完了" ? "未完了に戻す" : "完了にする"}
              </button>
            )}
          </div>
        </div>
      </div>
    );
  };

  return (
    <div className="dashboard-container">
      <header className="dashboard-header">
        <div className="header-left">
          <button onClick={() => navigate(`/select-group?user_id=${userId}`)} className="back-button">
            <ArrowLeft size={20} />
          </button>
          <div className="title-group">
            <div className="title-row">
              <h1>{group?.name || "Uni-Steps"}</h1>
              {group?.invite_code && (
                <button onClick={copyInviteCode} className="invite-badge" title="招待コードをコピー">
                  <Share2 size={12} />
                  コード: {group.invite_code}
                </button>
              )}
            </div>
            <div className="group-info">
              {group?.users && (
                <div className="member-list-summary">
                  <Users size={12} style={{marginRight: '4px'}} />
                  {group.users.map(u => u.name).join(', ')}
                </div>
              )}
            </div>
          </div>
        </div>
        <div className="header-actions">
          <button onClick={handleSync} disabled={loading} className="icon-button">
            <RefreshCw className={loading ? "animate-spin" : ""} size={16} />
            同期
          </button>
          <button onClick={handleOpenCreateModal} className="icon-button primary">
            <Plus size={16} />
            課題追加
          </button>
        </div>
      </header>

      <main className="dashboard-content">
        {error && <div className="error-message">{error}</div>}
        
        <section className="task-section">
          <h2>現在の課題</h2>
          {activeTasks.length === 0 ? (
            <p className="empty-state">現在取り組むべき課題はありません．</p>
          ) : (
            <div className="task-list">
              {activeTasks.map(renderTaskCard)}
            </div>
          )}
        </section>
        
        {archivedTasks.length > 0 && (
          <div className="archive-container">
            <div className="archive-header" onClick={() => setShowArchive(!showArchive)}>
              <span>過去の課題 ({archivedTasks.length}件)</span>
              <ChevronDown size={16} style={{transform: showArchive ? 'rotate(180deg)' : 'none', transition: '0.2s'}} />
            </div>
            
            {showArchive && (
              <div className="archive-content">
                <div className="task-list">
                  {archivedTasks.map(renderTaskCard)}
                </div>
              </div>
            )}
          </div>
        )}
      </main>

      {showTaskModal && (
        <div className="modal-overlay">
          <div className="modal-content animate-pop">
            <button onClick={() => setShowTaskModal(false)} className="close-button"><X size={20} /></button>
            <div className="modal-header-text">
              <h2>{editingTask ? "課題を編集" : "新しい課題"}</h2>
              <p>{editingTask ? "課題の詳細を変更します．" : "手動で課題を追加します．"}</p>
            </div>
            <form onSubmit={handleSaveTask}>
              <div className="form-group">
                <label>タイトル</label>
                <input type="text" required value={taskFormData.title} onChange={e => setTaskFormData({...taskFormData, title: e.target.value})} placeholder="例：レポート提出" />
              </div>
              <div className="form-group">
                <label>期限</label>
                <input type="datetime-local" value={taskFormData.deadline} onChange={e => setTaskFormData({...taskFormData, deadline: e.target.value})} />
              </div>
              <div className="form-group">
                <label>該当者</label>
                <div className="assignee-selector">
                  {group?.users?.map(member => (
                    <label key={member.id} className="assignee-item">
                      <input type="checkbox" checked={taskFormData.assignees.includes(member.id)} onChange={() => handleToggleAssignee(member.id)} />
                      <span>{member.name}</span>
                    </label>
                  ))}
                </div>
              </div>
              <button type="submit" disabled={loading} className="icon-button primary full-width" style={{marginTop: '1rem', justifyContent: 'center'}}>
                {editingTask ? "更新する" : "登録する"}
              </button>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default DashboardPage;
