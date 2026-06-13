import React, { useEffect, useState } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { taskApi } from '../api/tasks';
import { groupApi } from '../api/groups';
import { notificationApi } from '../api/notifications';
import { wakeupApi } from '../api/wakeup';
import { userApi } from '../api/user';
import type { Task, Group, WakeupCheck, NotificationLog } from '../types';
import { Bell, RefreshCw, PlusCircle, X, Settings, Plus, Archive, Edit, ChevronDown, ArrowLeft, Users, Calendar, CheckCircle, Clock, Copy, Share2, Trash2, BellRing, Sparkles, Send, Sunrise, Sun, AlertCircle } from 'lucide-react';

/**
 * Base64 URL 形式の文字列を Uint8Array に変換するヘルパー関数である．
 */
function urlBase64ToUint8Array(base64String: string) {
  const padding = '='.repeat((4 - base64String.length % 4) % 4);
  const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/');
  const rawData = window.atob(base64);
  const outputArray = new Uint8Array(rawData.length);
  for (let i = 0; i < rawData.length; ++i) {
    outputArray[i] = rawData.charCodeAt(i);
  }
  return outputArray;
}

const DashboardPage: React.FC = () => {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const userId = searchParams.get('user_id') || '';
  const groupId = searchParams.get('group_id') || '';
  
  const [tasks, setTasks] = useState<Task[]>([]);
  const [group, setGroup] = useState<Group | null>(null);
  const [activeWakeup, setActiveWakeup] = useState<WakeupCheck | null>(null);
  const [groupWakeups, setGroupWakeups] = useState<WakeupCheck[]>([]);
  const [notificationLogs, setNotificationLogs] = useState<NotificationLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  const [showTaskModal, setShowTaskModal] = useState(false);
  const [showSettingsModal, setShowSettingsModal] = useState(false);
  const [showWakeupModal, setShowWakeupModal] = useState(false);
  const [editingTask, setEditingTask] = useState<Task | null>(null);
  const [showArchive, setShowArchive] = useState(false);
  const [notifPermission, setNotifPermission] = useState<NotificationPermission>(Notification.permission);
  const [serverTokenMissing, setServerTokenMissing] = useState<boolean>(true);

  const [taskFormData, setTaskFormData] = useState({
    title: '',
    deadline: '',
    recurrence_type: 'none',
    assignees: [] as string[]
  });

  const [wakeupFormData, setWakeupFormData] = useState({
    target_time: '',
    grace_minutes: 15
  });

  const [settingsFormData, setSettingsFormData] = useState({
    remind_intervals: [] as number[],
    ai_character: 'default',
    line_channel_token: '',
    line_group_id: ''
  });

  const [newInterval, setNewInterval] = useState<string>('');

  // モーダルが開いた時に最新のグループ設定をフォームに反映する．
  useEffect(() => {
    if (showSettingsModal && group) {
      setSettingsFormData({
        remind_intervals: group.remind_intervals || [1440, 60],
        ai_character: group.ai_character || 'default',
        line_channel_token: group.line_channel_token || '',
        line_group_id: group.line_group_id || ''
      });
    }
  }, [showSettingsModal, group]);

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
      setError(null);
      
      const [taskDataRes, groupsRes, wakeupsRes, userRes, groupWakeupsRes, logsRes] = await Promise.allSettled([
        taskApi.listGroupTasks(groupId),
        groupApi.listMyGroups(userId),
        wakeupApi.getActive(userId),
        userApi.getMe(userId),
        wakeupApi.getActiveByGroup(groupId),
        groupApi.listNotifications(groupId)
      ]);

      if (taskDataRes.status === 'fulfilled') {
        setTasks(taskDataRes.value || []);
      }

      if (groupsRes.status === 'fulfilled') {
        const groups = groupsRes.value;
        if (groups && Array.isArray(groups)) {
          const currentGroup = groups.find(g => g.id === groupId);
          if (currentGroup) {
            setGroup(currentGroup);
          }
        }
      }

      if (wakeupsRes.status === 'fulfilled') {
        const wakeups = wakeupsRes.value;
        const pendingWakeup = wakeups.find(w => w.status === 'pending');
        setActiveWakeup(pendingWakeup || null);
      }

      if (userRes.status === 'fulfilled') {
        setServerTokenMissing(!userRes.value.has_push_token);
      }

      if (groupWakeupsRes.status === 'fulfilled') {
        setGroupWakeups(groupWakeupsRes.value || []);
      }

      if (logsRes.status === 'fulfilled') {
        setNotificationLogs(logsRes.value || []);
      }
    } catch (err: any) {
      console.error("Fetch error:", err);
      setError("データの取得中に予期せぬエラーが発生した．");
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

  const handleEnableNotifications = async () => {
    try {
      const permission = await Notification.requestPermission();
      setNotifPermission(permission);
      if (permission === 'granted') {
        const registration = await navigator.serviceWorker.register('/sw.js');
        const existingSub = await registration.pushManager.getSubscription();
        if (existingSub) await existingSub.unsubscribe();
        const vapidPublicKey = 'BDj40a3LAnB-Tyemxggm-wYyuHbE_kadO6CqX6u-Ewyrkqi5ypr-txXJO7jflV_4VGa47paZU7DX_-0OPZy6Bx8';
        const convertedVapidKey = urlBase64ToUint8Array(vapidPublicKey);
        const subscription = await registration.pushManager.subscribe({
          userVisibleOnly: true,
          applicationServerKey: convertedVapidKey
        });
        await notificationApi.subscribe(userId, subscription);
        setServerTokenMissing(false);
        alert("通知が有効になりました！AI からのリマインドが届くようになります．");
      }
    } catch (err) {
      console.error("Notification error:", err);
      alert("通知の設定に失敗しました．");
    }
  };

  const handleSendTestNotification = async () => {
    try {
      setLoading(true);
      await notificationApi.sendTestNotification(userId, settingsFormData.ai_character, groupId);
      alert("テスト通知を送信しました．数秒以内に届くはずです．");
    } catch (err: any) {
      alert(`テスト通知の送信に失敗しました：${err.response?.data?.error || err.message}`);
    } finally {
      setLoading(false);
    }
  };

  const handleRequestWakeup = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      setLoading(true);
      await wakeupApi.request({
        user_id: userId,
        group_id: groupId,
        target_time: new Date(wakeupFormData.target_time).toISOString(),
        grace_minutes: wakeupFormData.grace_minutes
      });
      setShowWakeupModal(false);
      await fetchData();
      alert("起床見守りを開始しました．明日の朝，忘れずにチェックインしてください！");
    } catch (err: any) {
      alert("見守り予約に失敗しました．");
    } finally {
      setLoading(false);
    }
  };

  const handleCheckIn = async () => {
    try {
      setLoading(true);
      await wakeupApi.checkin(userId);
      await fetchData();
      alert("おはようございます！起床を確認しました．SOS 通知を解除しました．");
    } catch (err) {
      alert("チェックインに失敗しました．");
    } finally {
      setLoading(false);
    }
  };

  const handleCancelWakeup = async () => {
    if (!window.confirm("見守りをキャンセルしますか？")) return;
    try {
      setLoading(true);
      await wakeupApi.cancel(userId);
      await fetchData();
      alert("見守りをキャンセルしました．");
    } catch (err) {
      alert("キャンセルの実行に失敗しました．");
    } finally {
      setLoading(false);
    }
  };

  const handleEditWakeup = () => {
    if (activeWakeup) {
      const date = new Date(activeWakeup.target_time);
      const localStr = new Date(date.getTime() - date.getTimezoneOffset() * 60000).toISOString().slice(0, 16);
      setWakeupFormData({
        target_time: localStr,
        grace_minutes: activeWakeup.grace_minutes
      });
      setShowWakeupModal(true);
    }
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
      const taskData = { title: taskFormData.title, deadline: deadline, recurrence: { type: taskFormData.recurrence_type, custom_dates: [] }, user_progress: userProgress as any };
      if (editingTask) {
        if (!editingTask.id) throw new Error("課題 ID が見つからないため更新できない．");
        await taskApi.updateTask(editingTask.id, taskData);
      } else {
        await taskApi.createManualTask({ ...taskData, group_id: groupId });
      }
      setShowTaskModal(false);
      await fetchData();
    } catch (err: any) {
      alert(`課題の保存に失敗した：${err.response?.data?.error || err.message}`);
    } finally {
      setLoading(false);
    }
  };

  const handleSaveSettings = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      setLoading(true);
      await groupApi.updateSettings(groupId, settingsFormData.remind_intervals, userId, settingsFormData.ai_character, settingsFormData.line_channel_token, settingsFormData.line_group_id);
      setShowSettingsModal(false);
      await fetchData();
      alert("設定を保存しました．");
    } catch (err: any) {
      alert(err.response?.data?.error || "設定の保存に失敗しました．");
    } finally {
      setLoading(false);
    }
  };

  const addInterval = (e: React.MouseEvent) => {
    e.preventDefault();
    const val = parseInt(newInterval);
    if (settingsFormData.remind_intervals.length >= 3) {
      alert("リマインド通知は最大 3 つまで設定できます．");
      return;
    }
    if (!isNaN(val) && val > 0) {
      if (settingsFormData.remind_intervals.includes(val)) {
        alert("その時間は既に設定されています．");
        return;
      }
      setSettingsFormData(prev => ({
        ...prev,
        remind_intervals: [...prev.remind_intervals, val].sort((a, b) => b - a)
      }));
      setNewInterval('');
    }
  };

  const removeInterval = (val: number) => {
    setSettingsFormData(prev => ({
      ...prev,
      remind_intervals: prev.remind_intervals.filter(i => i !== val)
    }));
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
    const myProgress = t.user_progress?.find(p => p.user_id === userId);
    const deadlineDate = new Date(t.deadline);
    return deadlineDate.getFullYear() <= 1 ? !myProgress?.is_completed : deadlineDate >= new Date();
  });
  const archivedTasks = tasks.filter(t => {
    const myProgress = t.user_progress?.find(p => p.user_id === userId);
    const deadlineDate = new Date(t.deadline);
    return deadlineDate.getFullYear() <= 1 ? myProgress?.is_completed : deadlineDate < new Date();
  }).sort((a, b) => new Date(b.deadline).getTime() - new Date(a.deadline).getTime());

  const renderTaskCard = (task: Task) => {
    const myStatus = getTaskStatus(task);
    const deadlineDate = new Date(task.deadline);
    const isUndetermined = deadlineDate.getFullYear() <= 1;
    const canEdit = task.source !== 'google_classroom' || !task.is_lms_deadline_set;
    return (
      <div key={task.id} className="task-card">
        <div className="task-info">
          <h3>{task.title}</h3>
          <p className="deadline"><Calendar size={13} /><span>{isUndetermined ? "期限未定" : deadlineDate.toLocaleString('ja-JP', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })}</span></p>
          <div className="member-status-list">
            {task.user_progress?.map(p => (
              <span key={p.user_id} className={`member-badge ${p.is_completed ? 'completed' : ''}`}>{p.user_name || "User"}: {p.is_completed ? "完了" : "未"}</span>
            ))}
          </div>
        </div>
        <div className="task-status">
          <span className={`status-badge ${myStatus.className}`}>{myStatus.label}</span>
          <div className="card-actions">
            {canEdit && <button onClick={() => {
              setEditingTask(task);
              const date = new Date(task.deadline);
              const deadlineStr = date.getFullYear() <= 1 ? "" : new Date(date.getTime() - date.getTimezoneOffset() * 60000).toISOString().slice(0, 16);
              setTaskFormData({ title: task.title, deadline: deadlineStr, recurrence_type: task.recurrence?.type || 'none', assignees: task.user_progress?.map(p => p.user_id) || [] });
              setShowTaskModal(true);
            }} className="edit-button">編集</button>}
            {task.source !== 'google_classroom' && <button onClick={() => taskApi.toggleTaskCompletion(task.id, userId).then(fetchData)} className="complete-toggle-button">{myStatus.label === "完了" ? "未完了に戻す" : "完了にする"}</button>}
          </div>
        </div>
      </div>
    );
  };

  return (
    <div className="dashboard-container">
      <header className="dashboard-header">
        <div className="header-left">
          <button onClick={() => navigate(`/select-group?user_id=${userId}`)} className="back-button"><ArrowLeft size={20} /></button>
          <div className="title-group">
            <div className="title-row">
              <h1>{group?.name || "Uni-Steps"}</h1>
              {group?.invite_code && <button onClick={() => { navigator.clipboard.writeText(group.invite_code); alert("コピーしました！"); }} className="invite-badge"><Share2 size={12} />コード: {group.invite_code}</button>}
            </div>
            <div className="group-info">{group?.users && <div className="member-list-summary"><Users size={12} style={{marginRight: '4px'}} />{group.users.map(u => u.name).join(', ')}</div>}</div>
          </div>
        </div>
        <div className="header-actions">
          {notifPermission === 'granted' && !serverTokenMissing && (
            <button onClick={handleSendTestNotification} className="icon-button" title="通知テスト"><Send size={16} /><span>テスト</span></button>
          )}
          {(notifPermission !== 'granted' || serverTokenMissing) && (
            <button onClick={handleEnableNotifications} className="icon-button warning-btn"><BellRing size={16} /><span>通知を有効化</span></button>
          )}
          <button onClick={handleSync} disabled={loading} className="icon-button"><RefreshCw className={loading ? "animate-spin" : ""} size={16} />同期</button>
          <button onClick={() => { setEditingTask(null); setTaskFormData({ title: '', deadline: '', recurrence_type: 'none', assignees: [userId] }); setShowTaskModal(true); }} className="icon-button primary"><Plus size={16} />課題追加</button>
          {group?.owner_id === userId && (
            <button onClick={() => setShowSettingsModal(true)} className="icon-button" title="設定"><Settings size={18} /></button>
          )}
        </div>
      </header>

      <main className="dashboard-content">
        {error && <div className="error-message">{error}</div>}

        <section className="wakeup-section">
          {activeWakeup ? (
            <div className="wakeup-card active">
              <div className="wakeup-info">
                <Sunrise size={24} className="wakeup-icon" />
                <div>
                  <h3>起床見守り中</h3>
                  <p>予定時刻: {new Date(activeWakeup.target_time).toLocaleString('ja-JP', { hour: '2-digit', minute: '2-digit' })} (+{activeWakeup.grace_minutes}分猶予)</p>
                </div>
              </div>
              <div className="wakeup-actions" style={{display: 'flex', gap: '0.5rem'}}>
                <button onClick={handleEditWakeup} className="icon-button" title="時間を変更" style={{background: 'white', color: 'var(--text-sub)'}}><Edit size={16} /></button>
                <button onClick={handleCancelWakeup} className="icon-button" title="見守りをやめる" style={{background: 'white', color: 'var(--error)'}}><Trash2 size={16} /></button>
                <button onClick={handleCheckIn} className="checkin-button">
                  <Sun size={18} />
                  <span>起きました！</span>
                </button>
              </div>
            </div>
          ) : (
            <div className="wakeup-card empty" onClick={() => {
               const tomorrow = new Date();
               tomorrow.setDate(tomorrow.getDate() + 1);
               tomorrow.setHours(8, 0, 0, 0);
               setWakeupFormData({ target_time: tomorrow.toISOString().slice(0, 16), grace_minutes: 15 });
               setShowWakeupModal(true);
            }}>
              < Sunrise size={20} />
              <span>明日の起床時間をセットして，仲間に見守ってもらう</span>
            </div>
          )}
        </section>

        {groupWakeups.length > 0 && (
          <section className="group-wakeup-section" style={{marginBottom: '2rem'}}>
            <h3 style={{fontSize: '0.95rem', color: 'var(--text-sub)', marginBottom: '0.8rem'}}>メンバーの起床予定</h3>
            <div className="member-status-list" style={{display: 'flex', flexDirection: 'column', gap: '0.5rem'}}>
              {groupWakeups.map(w => {
                const member = group?.users?.find(u => u.id === w.user_id);
                if (!member) return null;
                const isMe = member.id === userId;
                return (
                  <div key={w.id} style={{
                    display: 'flex', 
                    alignItems: 'center', 
                    gap: '0.5rem', 
                    background: isMe ? 'var(--primary-soft)' : '#f8fafc', 
                    border: isMe ? '1px solid var(--primary-light)' : '1px solid transparent',
                    padding: '0.6rem', 
                    borderRadius: '6px', 
                    fontSize: '0.85rem'
                  }}>
                    <span style={{fontWeight: 600}}>{member.name}{isMe ? ' (自分)' : ''}</span>
                    <span style={{color: 'var(--text-sub)'}}>: {new Date(w.target_time).toLocaleString('ja-JP', { month: 'numeric', day: 'numeric', hour: '2-digit', minute: '2-digit' })} (+{w.grace_minutes}分)</span>
                    {w.status === 'pending' && <span style={{marginLeft: 'auto', color: 'var(--warning)', fontWeight: 600}}>見守り中</span>}
                    {w.status === 'alerted' && <span style={{marginLeft: 'auto', color: 'var(--error)', fontWeight: 600}}>寝坊！</span>}
                  </div>
                );
              })}
            </div>
          </section>
        )}
        
        <section className="task-section">
          <h2>現在の課題</h2>
          {activeTasks.length === 0 ? <p className="empty-state">現在取り組むべき課題はありません．</p> : <div className="task-list">{activeTasks.map(renderTaskCard)}</div>}
        </section>

        {/* 通知履歴（タイムライン）セクション */}
        <section className="timeline-section" style={{marginTop: '3rem', borderTop: '1px solid var(--neutral-200)', paddingTop: '2rem'}}>
          <div style={{display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '1.5rem'}}>
            <Sparkles size={20} color="var(--primary)" />
            <h2 style={{fontSize: '1.1rem', margin: 0}}>AI ログ ＆ 通知履歴</h2>
          </div>
          {notificationLogs.length === 0 ? (
            <p className="empty-state" style={{padding: '1.5rem'}}>まだ通知の履歴はありません．</p>
          ) : (
            <div className="timeline-list" style={{display: 'flex', flexDirection: 'column', gap: '1rem'}}>
              {notificationLogs.map(log => {
                const isSOS = log.type === 'sos';
                return (
                  <div key={log.id} style={{
                    background: isSOS ? '#fff1f2' : 'white',
                    border: isSOS ? '1px solid #fecaca' : '1px solid var(--neutral-200)',
                    padding: '1rem',
                    borderRadius: '12px',
                    position: 'relative'
                  }}>
                    <div style={{display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem'}}>
                      <span style={{
                        fontSize: '0.75rem', 
                        fontWeight: 700, 
                        color: isSOS ? '#e11d48' : 'var(--primary)',
                        textTransform: 'uppercase'
                      }}>
                        {isSOS ? '🚨 SOS ALERT' : '🤖 AI REMIND'}
                      </span>
                      <span style={{fontSize: '0.75rem', color: 'var(--text-sub)'}}>
                        {new Date(log.created_at).toLocaleString('ja-JP', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })}
                      </span>
                    </div>
                    <p style={{margin: 0, fontSize: '0.9rem', lineHeight: 1.5, color: 'var(--text-main)'}}>{log.message}</p>
                  </div>
                );
              })}
            </div>
          )}
        </section>
        
        {archivedTasks.length > 0 && (
          <div className="archive-container">
            <div className="archive-header" onClick={() => setShowArchive(!showArchive)}><span>過去の課題 ({archivedTasks.length}件)</span><ChevronDown size={16} style={{transform: showArchive ? 'rotate(180deg)' : 'none', transition: '0.2s'}} /></div>
            {showArchive && <div className="archive-content"><div className="task-list">{archivedTasks.map(renderTaskCard)}</div></div>}
          </div>
        )}
      </main>

      {/* 課題登録・編集モーダル */}
      {showTaskModal && (
        <div className="modal-overlay">
          <div className="modal-content animate-pop">
            <button onClick={() => setShowTaskModal(false)} className="close-button"><X size={20} /></button>
            <div className="modal-header-text"><h2>{editingTask ? "課題を編集" : "新しい課題"}</h2><p>課題の詳細を入力してください．</p></div>
            <form onSubmit={handleSaveTask}>
              <div className="form-group"><label>タイトル</label><input type="text" required value={taskFormData.title} onChange={e => setTaskFormData({...taskFormData, title: e.target.value})} disabled={editingTask?.source === 'google_classroom'} /></div>
              <div className="form-group"><label>期限</label><input type="datetime-local" value={taskFormData.deadline} onChange={e => setTaskFormData({...taskFormData, deadline: e.target.value})} /></div>
              <div className="form-group"><label>該当者</label>
                <div className="assignee-selector">
                  {group?.users?.map(member => (
                    <label key={member.id} className="assignee-item">
                      <input type="checkbox" checked={taskFormData.assignees.includes(member.id)} onChange={() => {
                        const isSelected = taskFormData.assignees.includes(member.id);
                        setTaskFormData({ ...taskFormData, assignees: isSelected ? taskFormData.assignees.filter(id => id !== member.id) : [...taskFormData.assignees, member.id] });
                      }} disabled={editingTask?.source === 'google_classroom'} />
                      <span>{member.name}</span>
                    </label>
                  ))}
                </div>
              </div>
              <button type="submit" disabled={loading} className="icon-button primary full-width" style={{marginTop: '1rem', justifyContent: 'center'}}>{editingTask ? "更新する" : "登録する"}</button>
            </form>
          </div>
        </div>
      )}

      {/* 起床設定モーダル */}
      {showWakeupModal && (
        <div className="modal-overlay">
          <div className="modal-content animate-pop">
            <button onClick={() => setShowWakeupModal(false)} className="close-button"><X size={20} /></button>
            <div className="modal-header-text"><h2>起床見守りをセット</h2><p>もし起きられなかった場合，仲間に SOS 通知が飛びます．</p></div>
            <form onSubmit={handleRequestWakeup}>
              <div className="form-group"><label>起床予定時刻</label><input type="datetime-local" required value={wakeupFormData.target_time} onChange={e => setWakeupFormData({...wakeupFormData, target_time: e.target.value})} /></div>
              <div className="form-group">
                <label>猶予時間（分）</label>
                <input 
                  type="number" 
                  required 
                  value={isNaN(wakeupFormData.grace_minutes) ? '' : wakeupFormData.grace_minutes} 
                  onChange={e => {
                    const val = parseInt(e.target.value);
                    setWakeupFormData({...wakeupFormData, grace_minutes: isNaN(val) ? 0 : val});
                  }} 
                  placeholder="例：15" 
                />
              </div>
              <div className="alert-info"><AlertCircle size={16} /><span>設定時刻から猶予時間を過ぎてもチェックインがない場合，部屋のメンバー全員に通知が飛びます．</span></div>
              <button type="submit" disabled={loading} className="icon-button primary full-width" style={{marginTop: '1.5rem', background: 'var(--warning)', border: 'none'}}>見守りを開始する</button>
            </form>
          </div>
        </div>
      )}

      {/* 設定モーダル */}
      {showSettingsModal && (
        <div className="modal-overlay">
          <div className="modal-content animate-pop">
            <button onClick={() => setShowSettingsModal(false)} className="close-button"><X size={20} /></button>
            <div className="modal-header-text"><h2>部屋の設定</h2><p>通知タイミングや AI の性格を変更できます．</p></div>
            <form onSubmit={handleSaveSettings}>
              <div className="form-group">
                <label>AI のキャラクター設定</label>
                <select className="full-width" value={settingsFormData.ai_character} onChange={e => setSettingsFormData({...settingsFormData, ai_character: e.target.value})} style={{padding: '0.8rem', borderRadius: '8px', border: '1px solid var(--neutral-300)'}}>
                  <option value="default">標準アシスタント</option>
                  <option value="strict">厳しい教官</option>
                  <option value="kind">心配性な幼馴染</option>
                  <option value="cool">冷徹な執事</option>
                </select>
              </div>
              <div className="form-group">
                <label>LINE Bot 連携（オプション）</label>
                <input type="text" className="full-width" style={{marginBottom: '0.5rem'}} value={settingsFormData.line_channel_token} onChange={e => setSettingsFormData({...settingsFormData, line_channel_token: e.target.value})} placeholder="LINE Channel Token" />
                <input type="text" className="full-width" value={settingsFormData.line_group_id} onChange={e => setSettingsFormData({...settingsFormData, line_group_id: e.target.value})} placeholder="LINE Group ID (例: Cxxxxx...)" />
                <p style={{fontSize: '0.8rem', color: 'var(--text-sub)', margin: '0.5rem 0 0 0'}}>※設定すると，SOS 通知などが指定した LINE グループにも送信されます．</p>
              </div>
              <div className="form-group">
                <label>リマインド通知タイミング（分前 / 最大3つ）</label>
                <div className="interval-list">
                  {settingsFormData.remind_intervals.map(val => (
                    <div key={val} className="interval-tag"><span>{val >= 1440 ? `${val/1440}日` : val >= 60 ? `${val/60}時間` : `${val}分`}前</span><button type="button" onClick={() => removeInterval(val)} className="remove-btn"><X size={12} /></button></div>
                  ))}
                </div>
                <div className="interval-input-group">
                  <input type="number" value={newInterval} onChange={e => setNewInterval(e.target.value)} placeholder="例：30" />
                  <button type="button" onClick={addInterval} className="icon-button">追加</button>
                </div>
              </div>
              <button type="submit" disabled={loading} className="icon-button primary full-width" style={{marginTop: '1rem', justifyContent: 'center'}}>設定を保存する</button>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default DashboardPage;
