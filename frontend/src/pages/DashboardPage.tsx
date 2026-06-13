import React, { useEffect, useState, useCallback } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { taskApi } from '../api/tasks';
import { groupApi } from '../api/groups';
import { wakeupApi } from '../api/wakeup';
import type { Task, TaskUserProgress } from '../types';
import { AxiosError } from 'axios';
import { toLocalISOString } from '../utils/helpers';

// Hooks & Context
import { useAuth } from '../hooks/useAuth';
import { useDashboardData } from '../hooks/useDashboardData';
import { useWebPush } from '../hooks/useWebPush';

// Sub-components
import DashboardHeader from '../components/dashboard/DashboardHeader';
import WakeupSection from '../components/dashboard/WakeupSection';
import TaskCard from '../components/dashboard/TaskCard';
import TimelineSection from '../components/dashboard/TimelineSection';

// Modals
import TaskModal from '../components/modals/TaskModal';
import SettingsModal from '../components/modals/SettingsModal';
import WakeupModal from '../components/modals/WakeupModal';

import { ChevronDown } from 'lucide-react';

const DashboardPage: React.FC = () => {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { userId } = useAuth();
  const groupId = searchParams.get('group_id') || '';
  
  // Custom Hooks
  const { 
    tasks, group, activeWakeup, groupWakeups, notificationLogs, 
    loading, error, serverTokenMissing, setServerTokenMissing,
    fetchData, setLoading 
  } = useDashboardData(userId, groupId);

  const { 
    notifPermission, handleEnableNotifications, handleSendTestNotification 
  } = useWebPush(userId, groupId);

  // Modal states
  const [showTaskModal, setShowTaskModal] = useState(false);
  const [showSettingsModal, setShowSettingsModal] = useState(false);
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
    remind_intervals: [] as number[],
    ai_character: 'default',
    line_channel_token: '',
    line_group_id: '',
    summary_morning_time: '08:00',
    summary_evening_time: '21:00'
  });

  const handleSync = useCallback(async (silent: boolean = false) => {
    try {
      if (!silent) setLoading(true);
      const result = await taskApi.syncTasks(userId, groupId);
      if (result && result.tasks && result.tasks.length > 0) {
        await fetchData();
        if (!silent) alert(`同期が完了した．${result.tasks.length} 件の課題を更新した．`);
      }
    } catch (err: unknown) {
      if (!silent) {
        const axiosErr = err as AxiosError<{error: string}>;
        alert(axiosErr.response?.data?.error || "同期に失敗した．");
      }
      console.error("Auto-sync error:", err);
    } finally {
      if (!silent) setLoading(false);
    }
  }, [userId, groupId, fetchData, setLoading]);

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

  const handleOpenSettings = () => {
    if (group) {
      setSettingsFormData({
        remind_intervals: group.remind_intervals || [1440, 60],
        ai_character: group.ai_character || 'default',
        line_channel_token: group.line_channel_token || '',
        line_group_id: group.line_group_id || '',
        summary_morning_time: group.summary_morning_time || '08:00',
        summary_evening_time: group.summary_evening_time || '21:00'
      });
      setShowSettingsModal(true);
    }
  };

  const onEnableNotif = () => handleEnableNotifications(() => setServerTokenMissing(false));
  const onTestNotif = () => handleSendTestNotification(settingsFormData.ai_character, () => setServerTokenMissing(true));

  const handleSaveTask = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      setLoading(true);
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
      const taskData = { title: taskFormData.title, deadline: deadline, recurrence: { type: taskFormData.recurrence_type, custom_dates: [] }, user_progress: userProgress };
      
      if (editingTask) {
        await taskApi.updateTask(editingTask.id, taskData);
      } else {
        await taskApi.createManualTask({ ...taskData, group_id: groupId });
      }
      setShowTaskModal(false);
      await fetchData();
    } catch (err: unknown) {
      const axiosErr = err as AxiosError<{error: string}>;
      alert(`課題の保存に失敗した：${axiosErr.response?.data?.error || axiosErr.message}`);
    } finally {
      setLoading(false);
    }
  };

  const handleSaveSettings = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      setLoading(true);
      await groupApi.updateSettings(
        groupId, settingsFormData.remind_intervals, userId, 
        settingsFormData.ai_character, settingsFormData.line_channel_token, settingsFormData.line_group_id,
        settingsFormData.summary_morning_time, settingsFormData.summary_evening_time
      );
      setShowSettingsModal(false);
      await fetchData();
      alert("設定を保存しました．");
    } catch (err: unknown) {
      const axiosErr = err as AxiosError<{error: string}>;
      alert(axiosErr.response?.data?.error || "設定の保存に失敗しました．");
    } finally {
      setLoading(false);
    }
  };

  const handleRequestWakeup = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      setLoading(true);
      await wakeupApi.request({
        user_id: userId, group_id: groupId,
        target_time: new Date(wakeupFormData.target_time).toISOString(),
        grace_minutes: wakeupFormData.grace_minutes
      });
      setShowWakeupModal(false);
      await fetchData();
      alert("起床見守りを開始しました．明日の朝，忘れずにチェックインしてください！");
    } catch (err) {
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
      setWakeupFormData({
        target_time: toLocalISOString(new Date(activeWakeup.target_time)),
        grace_minutes: activeWakeup.grace_minutes
      });
      setShowWakeupModal(true);
    }
  };

  const toggleTaskCompletion = async (taskId: string) => {
    try {
      await taskApi.toggleTaskCompletion(taskId, userId);
      await fetchData();
    } catch (err) {
      alert("進捗の更新に失敗しました．");
    }
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

  return (
    <div className="dashboard-container">
      <DashboardHeader 
        group={group} userId={userId}
        onBack={() => navigate(`/select-group?user_id=${userId}`)}
        onSendTestNotification={onTestNotif}
        onEnableNotifications={onEnableNotif}
        onAddTask={() => { setEditingTask(null); setTaskFormData({ title: '', deadline: '', recurrence_type: 'none', assignees: [userId] }); setShowTaskModal(true); }}
        onOpenSettings={handleOpenSettings}
        notifPermission={notifPermission}
        serverTokenMissing={serverTokenMissing}
      />

      <main className="dashboard-content">
        {error && <div className="error-message">{error}</div>}

        <WakeupSection 
          activeWakeup={activeWakeup} groupWakeups={groupWakeups}
          group={group} userId={userId}
          onSetWakeup={() => {
            const tomorrow = new Date();
            tomorrow.setDate(tomorrow.getDate() + 1);
            tomorrow.setHours(8, 0, 0, 0);
            setWakeupFormData({ target_time: toLocalISOString(tomorrow), grace_minutes: 15 });
            setShowWakeupModal(true);
          }}
          onEditWakeup={handleEditWakeup} onCancelWakeup={handleCancelWakeup} onCheckIn={handleCheckIn}
        />
        
        <section className="task-section">
          <h2>現在の課題</h2>
          {activeTasks.length === 0 ? (
            <p className="empty-state">現在取り組むべき課題はありません．</p>
          ) : (
            <div className="task-list">
              {activeTasks.map(t => (
                <TaskCard key={t.id} task={t} userId={userId} onToggleCompletion={toggleTaskCompletion} onEdit={(task) => {
                  setEditingTask(task);
                  setTaskFormData({ 
                    title: task.title, 
                    deadline: toLocalISOString(new Date(task.deadline)), 
                    recurrence_type: task.recurrence?.type || 'none', 
                    assignees: task.user_progress?.map(p => p.user_id) || [] 
                  });
                  setShowTaskModal(true);
                }} />
              ))}
            </div>
          )}
        </section>

        <TimelineSection logs={notificationLogs} />
        
        {archivedTasks.length > 0 && (
          <div className="archive-container">
            <div className="archive-header" onClick={() => setShowArchive(!showArchive)}>
              <span>過去の課題 ({archivedTasks.length}件)</span>
              <ChevronDown size={16} style={{transform: showArchive ? 'rotate(180deg)' : 'none', transition: '0.2s'}} />
            </div>
            {showArchive && (
              <div className="archive-content">
                <div className="task-list">
                  {archivedTasks.map(t => (
                    <TaskCard key={t.id} task={t} userId={userId} onToggleCompletion={toggleTaskCompletion} onEdit={() => {}} />
                  ))}
                </div>
              </div>
            )}
          </div>
        )}
      </main>

      <TaskModal 
        show={showTaskModal} onClose={() => setShowTaskModal(false)}
        editingTask={editingTask} formData={taskFormData} setFormData={setTaskFormData}
        onSave={handleSaveTask} group={group} loading={loading}
      />

      <WakeupModal 
        show={showWakeupModal} onClose={() => setShowWakeupModal(false)}
        formData={wakeupFormData} setFormData={setWakeupFormData}
        onSave={handleRequestWakeup} loading={loading}
      />

      <SettingsModal 
        show={showSettingsModal} onClose={() => setShowSettingsModal(false)}
        formData={settingsFormData} setFormData={setSettingsFormData}
        onSave={handleSaveSettings} loading={loading}
      />
    </div>
  );
};

export default DashboardPage;
