import React, { useEffect, useState, useCallback } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { taskApi } from '../api/tasks';
import { groupApi } from '../api/groups';
import { wakeupApi } from '../api/wakeup';
import type { Task, TaskUserProgress } from '../types';
import { AxiosError } from 'axios';
import { toLocalISOString, handle } from '../utils/helpers';

// Hooks & Context
import { useAuth } from '../hooks/useAuth';
import { useDashboardData } from '../hooks/useDashboardData';
import { useWebPush } from '../hooks/useWebPush';

// Sub-components
import DashboardHeader from '../components/dashboard/DashboardHeader';
import WakeupSection from '../components/dashboard/WakeupSection';
import TaskCard from '../components/dashboard/TaskCard';
import TimelineSection from '../components/dashboard/TimelineSection';
import HomeOverview from '../components/dashboard/HomeOverview';
import AppShell from '../components/layout/AppShell';

// Modals
import TaskModal from '../components/modals/TaskModal';
import WakeupModal from '../components/modals/WakeupModal';

import { ChevronDown, ListTodo, Archive as ArchiveIcon, Sparkles, Hash, Layout, MessageCircle, Clock, Bell, X, Save, Settings, Trash2 } from 'lucide-react';

const DashboardPage: React.FC = () => {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { userId, user } = useAuth();
  const groupId = searchParams.get('group_id') || '';
  
  // Custom Hooks
  const { 
    tasks, group, activeWakeup, groupWakeups, notificationLogs, 
    loading, error, serverTokenMissing, setServerTokenMissing,
    fetchData 
  } = useDashboardData(userId, groupId);

  const { 
    notifPermission, handleEnableNotifications, handleSilentResubscribe 
  } = useWebPush(userId, groupId);

  // 通知の自動復旧ロジック：許可済みだがサーバーにトークンがない場合，サイレントに再登録する．
  useEffect(() => {
    if (serverTokenMissing && notifPermission === 'granted') {
      console.log("DashboardPage: 通知トークンの自動復旧を開始する...");
      handleSilentResubscribe(() => setServerTokenMissing(false));
    }
  }, [serverTokenMissing, notifPermission, handleSilentResubscribe, setServerTokenMissing]);

  // View state
  const [activeTab, setActiveTab] = useState('tasks');

  // Action loading state (To prevent full-page loading during saves)
  const [isProcessing, setIsProcessing] = useState(false);

  // Modal states
  const [showTaskModal, setShowTaskModal] = useState(false);
  const [showWakeupModal, setShowWakeupModal] = useState(false);
  const [editingTask, setEditingTask] = useState<Task | null>(null);
  const [showArchive, setShowArchive] = useState(false);

  const [taskFormData, setTaskFormData] = useState({
    title: '', deadline: '', recurrence_type: 'none', assignees: [] as string[]
  });

  const [wakeupFormData, setWakeupFormData] = useState({
    target_time: '', grace_minutes: 15
  });

  const [settingsFormData, setSettingsFormData] = useState({
    name: '',
    remind_intervals: [] as number[],
    ai_character: 'default',
    line_channel_token: '',
    line_group_id: '',
    summary_morning_time: '08:00',
    summary_evening_time: '21:00'
  });

  const [newInterval, setNewInterval] = useState<string>('');

  const handleSync = useCallback(async (silent: boolean = false) => {
    if (!silent) setIsProcessing(true);
    const [result, err] = await handle(taskApi.syncTasks(userId, groupId));
    if (err) {
      if (!silent) {
        const axiosErr = err as AxiosError<{error: string}>;
        alert(axiosErr.response?.data?.error || "同期に失敗した．");
      }
      console.error("Auto-sync error:", err.message);
      if (!silent) setIsProcessing(false);
      return;
    }

    if (result && result.tasks && result.tasks.length > 0) {
      await fetchData();
      if (!silent) alert(`同期が完了した．${result.tasks.length} 件の課題を更新した．`);
    }
    if (!silent) setIsProcessing(false);
  }, [userId, groupId, fetchData]);

  useEffect(() => {
    if (!userId || !groupId) {
      navigate('/login');
      return;
    }
    const initialize = async () => {
      await fetchData();
      handleSync(true);
    };
    initialize();
  }, [userId, groupId, navigate, fetchData, handleSync]);

  // 設定データが取得できたらフォームに初期値をセットする．
  useEffect(() => {
    if (group) {
      setSettingsFormData({
        name: group.name || '',
        remind_intervals: group.remind_intervals || [1440, 60],
        ai_character: group.ai_character || 'default',
        line_channel_token: group.line_channel_token || '',
        line_group_id: group.line_group_id || '',
        summary_morning_time: group.summary_morning_time || '08:00',
        summary_evening_time: group.summary_evening_time || '21:00'
      });
    }
  }, [group]);

  // モーダル表示時に背景のスクロールを禁止する．
  useEffect(() => {
    if (showTaskModal || showWakeupModal) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = 'unset';
    }
    return () => { document.body.style.overflow = 'unset'; };
  }, [showTaskModal, showWakeupModal]);

  const onEnableNotif = () => handleEnableNotifications(() => setServerTokenMissing(false));

  const handleSaveTask = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsProcessing(true);
    const deadline = taskFormData.deadline ? new Date(taskFormData.deadline).toISOString() : "0001-01-01T00:00:00Z";
    const userProgress: TaskUserProgress[] = taskFormData.assignees.map(id => {
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
      creator_id: editingTask?.creator_id || userId,
      recurrence: { type: taskFormData.recurrence_type, custom_dates: [] }, 
      user_progress: userProgress 
    };
    
    const [, err] = editingTask 
      ? await handle(taskApi.updateTask(editingTask.id, taskData, userId))
      : await handle(taskApi.createManualTask({ ...taskData, group_id: groupId }));

    if (err) {
      const axiosErr = err as AxiosError<{error: string}>;
      alert(`課題の保存に失敗した：${axiosErr.response?.data?.error || axiosErr.message}`);
      setIsProcessing(false);
      return;
    }

    setShowTaskModal(false);
    await fetchData();
    setIsProcessing(false);
  };

  const handleSaveSettings = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsProcessing(true);
    const [, err] = await handle(groupApi.updateSettings(
      groupId, settingsFormData.name, settingsFormData.remind_intervals, userId, 
      settingsFormData.ai_character, settingsFormData.line_channel_token, settingsFormData.line_group_id,
      settingsFormData.summary_morning_time, settingsFormData.summary_evening_time
    ));

    if (err) {
      const axiosErr = err as AxiosError<{error: string}>;
      alert(axiosErr.response?.data?.error || "設定の保存に失敗しました．");
      setIsProcessing(false);
      return;
    }

    await fetchData();
    alert("設定を保存しました．");
    setIsProcessing(false);
  };

  const handleLeaveGroup = async () => {
    const isOwner = group?.owner_id === userId;
    const membersCount = group?.users?.length || 0;

    if (isOwner) {
      if (membersCount === 1) {
        alert("あなたはこの部屋の唯一のメンバーであり，オーナーです．退出する前に部屋を削除することを推奨します．");
        return;
      }
      
      const otherMembers = group?.users?.filter(u => u.id !== userId) || [];
      const memberList = otherMembers.map((m, i) => `${i + 1}: ${m.name}`).join('\n');
      const choice = window.prompt(`オーナーが退出するには，次のオーナーを指名する必要があります．番号を入力してください：\n${memberList}`);
      
      if (choice === null) return;
      const idx = parseInt(choice) - 1;
      if (isNaN(idx) || idx < 0 || idx >= otherMembers.length) {
        alert("無効な番号が選択されました．");
        return;
      }

      const successor = otherMembers[idx];
      if (!window.confirm(`${successor.name} さんを次のオーナーに指名して退出しますか？`)) return;

      setIsProcessing(true);
      const [, tErr] = await handle(groupApi.transferOwnership(groupId, userId, successor.id));
      if (tErr) {
        alert("権限譲渡に失敗しました．" + tErr.message);
        setIsProcessing(false);
        return;
      }
      const [, lErr] = await handle(groupApi.leaveGroup(groupId, userId));
      if (lErr) {
        alert("退出に失敗しました．" + lErr.message);
        setIsProcessing(false);
        return;
      }
      navigate(`/select-group?user_id=${userId}`);
      return;
    }

    if (!window.confirm("この部屋から退出しますか？一度退出すると，再度招待コードが必要になります．")) return;
    setIsProcessing(true);
    const [, err] = await handle(groupApi.leaveGroup(groupId, userId));
    if (err) {
      alert("部屋の退出に失敗しました．" + err.message);
      setIsProcessing(false);
      return;
    }
    navigate(`/select-group?user_id=${userId}`);
  };

  const handleDeleteGroup = async () => {
    if (!window.confirm("【警告】この部屋を完全に削除しますか？この操作は取り消せません．メンバー全員のデータが消去されます．")) return;
    setIsProcessing(true);
    const [, err] = await handle(groupApi.deleteGroup(groupId, userId));
    if (err) {
      const axiosErr = err as AxiosError<{error: string}>;
      alert(axiosErr.response?.data?.error || "部屋の削除に失敗しました．");
      setIsProcessing(false);
      return;
    }
    navigate(`/select-group?user_id=${userId}`);
  };

  const handleDeleteTask = async (taskId: string) => {
    if (!window.confirm("この課題を完全に削除しますか？")) return;
    setIsProcessing(true);
    const [, err] = await handle(taskApi.deleteTask(taskId, userId));
    if (err) {
      alert("課題の削除に失敗しました．" + err.message);
      setIsProcessing(false);
      return;
    }
    await fetchData();
    setIsProcessing(false);
  };

  const addInterval = (e: React.MouseEvent) => {
    e.preventDefault();
    const val = parseInt(newInterval);
    if (settingsFormData.remind_intervals.length >= 3) {
      alert("リマインド通知は最大 3 つまで設定できます．");
      return;
    }
    if (!isNaN(val) && val > 0 && !settingsFormData.remind_intervals.includes(val)) {
      setSettingsFormData({
        ...settingsFormData,
        remind_intervals: [...settingsFormData.remind_intervals, val].sort((a, b) => b - a)
      });
      setNewInterval('');
    }
  };

  const removeInterval = (val: number) => {
    setSettingsFormData({
      ...settingsFormData,
      remind_intervals: settingsFormData.remind_intervals.filter(i => i !== val)
    });
  };

  const handleRequestWakeup = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsProcessing(true);
    const [, err] = await handle(wakeupApi.request({
      user_id: userId, group_id: groupId,
      target_time: new Date(wakeupFormData.target_time).toISOString(),
      grace_minutes: wakeupFormData.grace_minutes
    }));
    if (err) {
      alert("見守り予約に失敗しました．" + err.message);
      setIsProcessing(false);
      return;
    }
    setShowWakeupModal(false);
    await fetchData();
    alert("起床見守りを開始しました．明日の朝，忘れずにチェックインしてください！");
    setIsProcessing(false);
  };

  const handleCheckIn = async () => {
    setIsProcessing(true);
    const [, err] = await handle(wakeupApi.checkin(userId));
    if (err) {
      alert("チェックインに失敗しました．" + err.message);
      setIsProcessing(false);
      return;
    }
    await fetchData();
    alert("おはようございます！起床を確認しました．SOS 通知を解除しました．");
    setIsProcessing(false);
  };

  const handleCancelWakeup = async () => {
    if (!window.confirm("見守りをキャンセルしますか？")) return;
    setIsProcessing(true);
    const [, err] = await handle(wakeupApi.cancel(userId));
    if (err) {
      alert("キャンセルの実行に失敗しました．" + err.message);
      setIsProcessing(false);
      return;
    }
    await fetchData();
    alert("見守りをキャンセルしました．");
    setIsProcessing(false);
  };

  const handleEditWakeup = () => {
    if (activeWakeup) {
      setWakeupFormData({
        target_time: toLocalISOString(new Date(activeWakeup.target_time)),
        grace_minutes: activeWakeup.grace_minutes
      });
      setShowWakeupModal(true);
    }
  };

  const handleEditTask = useCallback((task: Task) => {
    setEditingTask(task);
    setTaskFormData({ 
      title: task.title, 
      deadline: toLocalISOString(new Date(task.deadline)), 
      recurrence_type: task.recurrence?.type || 'none', 
      assignees: task.user_progress?.map(p => p.user_id) || [] 
    });
    setShowTaskModal(true);
  }, []);

  const toggleTaskCompletion = async (taskId: string) => {
    const [, err] = await handle(taskApi.toggleTaskCompletion(taskId, userId));
    if (err) {
      alert("進捗の更新に失敗しました．" + err.message);
      return;
    }
    await fetchData();
  };

  // 課題の分類ロジックである．
  // Active: (まだ完了していない担当者が一人でもいる) かつ (期限が未来 または 期限未定)
  const activeTasks = tasks.filter(t => {
    const deadlineDate = new Date(t.deadline);
    const isUndetermined = deadlineDate.getFullYear() <= 1;
    const isPast = !isUndetermined && deadlineDate < new Date();
    
    // 全員の完了状態を確認する（担当者がいない場合は完了とみなさない）．
    const allCompleted = t.user_progress && t.user_progress.length > 0 && t.user_progress.every(p => p.is_completed);
    
    return !allCompleted && !isPast;
  });

  // Archived: (全員が完了した) または (期限切れ)
  const archivedTasks = tasks.filter(t => {
    const deadlineDate = new Date(t.deadline);
    const isUndetermined = deadlineDate.getFullYear() <= 1;
    const isPast = !isUndetermined && deadlineDate < new Date();
    
    const allCompleted = t.user_progress && t.user_progress.length > 0 && t.user_progress.every(p => p.is_completed);
    
    return allCompleted || isPast;
  }).sort((a, b) => new Date(b.deadline).getTime() - new Date(a.deadline).getTime());

  // 初回ロード時のみ全画面ローディングを出す（group がまだ無い時）．
  if (loading && !group) {
    return (
      <div className="full-page-loading">
        <Sparkles className="loading-logo" size={48} color="var(--brand)" />
        <p style={{color: 'var(--text-secondary)', fontWeight: 600}}>データを準備しています...</p>
      </div>
    );
  }

  const isOwner = group?.owner_id === userId;

  return (
    <AppShell 
      activeTab={activeTab} 
      onTabChange={setActiveTab}
      header={
        <DashboardHeader 
          group={group} userId={userId}
          onBack={() => navigate(`/select-group?user_id=${userId}`)}
          onEnableNotifications={onEnableNotif}
          onAddTask={() => { setEditingTask(null); setTaskFormData({ title: '', deadline: '', recurrence_type: 'none', assignees: [userId] }); setShowTaskModal(true); }}
          notifPermission={notifPermission}
          serverTokenMissing={serverTokenMissing}
          loading={loading}
        />
      }
    >
      {error && <div className="error-message">{error}</div>}

      <div className="dashboard-view">
        {activeTab === 'tasks' && (
          <section>
            <HomeOverview tasks={tasks} group={group} userName={user?.name || ''} userId={userId} />

            <div className="section-header-rich">
              <h2><ListTodo size={22} color="var(--brand)" />取り組むべき課題</h2>
            </div>
            
            {activeTasks.length === 0 ? (
              <div className="empty-state"><p>現在取り組むべき課題はありません．</p></div>
            ) : (
              <div className="task-grid">
                {activeTasks.map(t => (
                  <TaskCard key={t.id} task={t} userId={userId} onToggleCompletion={toggleTaskCompletion} onEdit={handleEditTask} />
                ))}
              </div>
            )}

            {archivedTasks.length > 0 && (
              <div style={{marginTop: '4rem'}}>
                <div className="section-header-rich">
                  <div style={{display: 'flex', alignItems: 'center', gap: '10px'}}>
                    <ArchiveIcon size={20} color="var(--text-tertiary)" />
                    <h3 style={{margin: 0, fontSize: '1.1rem', color: 'var(--text-secondary)'}}>過去の課題 ({archivedTasks.length}件)</h3>
                  </div>
                  <button onClick={() => setShowArchive(!showArchive)} className="btn-ghost" style={{padding: '4px', borderRadius: '50%', border: 'none'}}>
                    <ChevronDown size={22} style={{transform: showArchive ? 'rotate(180deg)' : 'none', transition: '0.3s'}} />
                  </button>
                </div>
                {showArchive && (
                  <div className="task-grid">
                    {archivedTasks.map(t => (
                      <TaskCard key={t.id} task={t} userId={userId} onToggleCompletion={toggleTaskCompletion} onEdit={handleEditTask} />
                    ))}
                  </div>
                )}
              </div>
            )}
          </section>
        )}

        {activeTab === 'wakeup' && (
          <WakeupSection 
            activeWakeup={activeWakeup} groupWakeups={groupWakeups} group={group} userId={userId}
            onSetWakeup={() => {
              const tomorrow = new Date();
              tomorrow.setDate(tomorrow.getDate() + 1);
              tomorrow.setHours(8, 0, 0, 0);
              setWakeupFormData({ target_time: toLocalISOString(tomorrow), grace_minutes: 15 });
              setShowWakeupModal(true);
            }}
            onEditWakeup={handleEditWakeup} onCancelWakeup={handleCancelWakeup} onCheckIn={handleCheckIn}
          />
        )}

        {activeTab === 'timeline' && (
          <TimelineSection logs={notificationLogs} />
        )}

        {activeTab === 'settings' && (
          <section className="animate-pop" style={{maxWidth: '600px', margin: '0 auto'}}>
            <div className="section-header-rich" style={{marginBottom: '3rem'}}>
              <h2><Settings size={22} color="var(--brand)" /> 部屋の設定とカスタマイズ</h2>
            </div>

            {isOwner ? (
              <form onSubmit={handleSaveSettings} style={{background: 'white', padding: '2.5rem', borderRadius: '24px', border: '1px solid #e5e7eb', boxShadow: 'var(--shadow-sm)'}}>
                <div className="form-group">
                  <label><Hash size={14} /> 部屋の名前</label>
                  <input type="text" required value={settingsFormData.name} onChange={e => setSettingsFormData({...settingsFormData, name: e.target.value})} placeholder="例：ゼミ用，月曜2限" />
                </div>

                <div className="form-group">
                  <label><Layout size={14} /> AI の性格</label>
                  <select value={settingsFormData.ai_character} onChange={e => setSettingsFormData({...settingsFormData, ai_character: e.target.value})}>
                    <option value="default">標準アシスタント</option>
                    <option value="strict">軍隊の厳しい教官</option>
                    <option value="kind">心配性な幼馴染</option>
                    <option value="cool">冷徹な仕事人執事</option>
                  </select>
                </div>

                <div className="form-group">
                  <label><MessageCircle size={14} /> LINE 連携設定</label>
                  <div style={{display: 'flex', flexDirection: 'column', gap: '0.75rem'}}>
                    <input type="text" value={settingsFormData.line_channel_token} onChange={e => setSettingsFormData({...settingsFormData, line_channel_token: e.target.value})} placeholder="Channel Access Token" />
                    <input type="text" value={settingsFormData.line_group_id} onChange={e => setSettingsFormData({...settingsFormData, line_group_id: e.target.value})} placeholder="LINE Group ID (C...)" />
                  </div>
                </div>

                <div className="form-group">
                  <label><Clock size={14} /> サマリー配信時刻</label>
                  <div style={{display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem'}}>
                    <div>
                      <span style={{fontSize: '0.75rem', color: 'var(--text-tertiary)', fontWeight: 800}}>朝刊 (今日)</span>
                      <input type="time" value={settingsFormData.summary_morning_time} onChange={e => setSettingsFormData({...settingsFormData, summary_morning_time: e.target.value})} />
                    </div>
                    <div>
                      <span style={{fontSize: '0.75rem', color: 'var(--text-tertiary)', fontWeight: 800}}>夕刊 (明日)</span>
                      <input type="time" value={settingsFormData.summary_evening_time} onChange={e => setSettingsFormData({...settingsFormData, summary_evening_time: e.target.value})} />
                    </div>
                  </div>
                </div>

                <div className="form-group">
                  <label><Bell size={14} /> 通知のタイミング</label>
                  <div style={{display: 'flex', flexWrap: 'wrap', gap: '8px', marginBottom: '1.25rem'}}>
                    {settingsFormData.remind_intervals.map(val => (
                      <div key={val} className="timeline-badge remind" style={{display: 'flex', alignItems: 'center', gap: '8px', padding: '8px 12px', borderRadius: '12px'}}>
                        <span style={{fontWeight: 800}}>{val >= 1440 ? `${val/1440}日` : val >= 60 ? `${val/60}時間` : `${val}分`}前</span>
                        <button type="button" onClick={() => removeInterval(val)} style={{background: 'none', border: 'none', padding: 0, display: 'flex', cursor: 'pointer', color: 'var(--brand)'}}><X size={14} /></button>
                      </div>
                    ))}
                  </div>
                  <div style={{display: 'flex', gap: '10px'}}>
                    <input type="number" value={newInterval} onChange={e => setNewInterval(e.target.value)} placeholder="分前を入力" />
                    <button type="button" onClick={addInterval} className="btn btn-ghost" style={{whiteSpace: 'nowrap', borderRadius: '16px'}}>追加</button>
                  </div>
                </div>

                <button type="submit" disabled={isProcessing} className="btn btn-primary" style={{width: '100%', padding: '16px', marginTop: '1rem', borderRadius: '16px'}}>
                  <Save size={18} />
                  {isProcessing ? "更新中..." : "全ての設定を保存する"}
                </button>
              </form>
            ) : (
              <div style={{background: 'white', padding: '3rem', borderRadius: '24px', border: '1px solid #e5e7eb', textAlign: 'center'}}>
                <Settings size={48} color="var(--text-tertiary)" style={{opacity: 0.3, marginBottom: '1.5rem'}} />
                <p style={{color: 'var(--text-secondary)', fontSize: '1.1rem', fontWeight: 600}}>
                  この部屋の設定はオーナーのみ変更可能です．
                </p>
                <p style={{color: 'var(--text-tertiary)', fontSize: '0.9rem', marginTop: '0.5rem'}}>
                  通知設定や AI の性格を変更したい場合は，オーナーへ依頼してください．
                </p>
              </div>
            )}

            {/* 危険な操作セクション */}
            <div style={{marginTop: '4rem', padding: '2.5rem', background: '#fef2f2', borderRadius: '24px', border: '1px solid #fee2e2'}}>
              <h3 style={{margin: 0, color: 'var(--error)', fontSize: '1.25rem', fontWeight: 900}}>Danger Zone</h3>
              <p style={{fontSize: '0.95rem', color: 'var(--text-secondary)', marginTop: '0.6rem', lineHeight: 1.6}}>
                部屋から退出すると，この部屋の課題や通知を受け取れなくなります．
              </p>
              
              <div style={{display: 'flex', flexDirection: 'column', gap: '1rem', marginTop: '2rem'}}>
                <button 
                  onClick={handleLeaveGroup} 
                  disabled={isProcessing}
                  className="btn btn-ghost" 
                  style={{color: 'var(--error)', borderColor: '#fecaca', width: '100%', padding: '14px'}}
                >
                  部屋を退出する
                </button>

                {isOwner && (
                  <button 
                    onClick={handleDeleteGroup} 
                    disabled={isProcessing}
                    className="btn btn-primary" 
                    style={{background: 'var(--error)', width: '100%', padding: '14px'}}
                  >
                    <Trash2 size={18} />
                    部屋自体を完全に削除する
                  </button>
                )}
              </div>
            </div>
          </section>
        )}
      </div>

      <TaskModal 
        show={showTaskModal} onClose={() => setShowTaskModal(false)}
        editingTask={editingTask} formData={taskFormData} setFormData={setTaskFormData}
        onSave={handleSaveTask} onDelete={handleDeleteTask} group={group} loading={isProcessing}
        operatorId={userId} ownerId={group?.owner_id || ''}
      />

      <WakeupModal 
        show={showWakeupModal} onClose={() => setShowWakeupModal(false)}
        formData={wakeupFormData} setFormData={setWakeupFormData}
        onSave={handleRequestWakeup} loading={isProcessing}
      />
    </AppShell>
  );
};

export default DashboardPage;
