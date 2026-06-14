import { useState, useCallback } from 'react';
import { taskApi } from '../api/tasks';
import { groupApi } from '../api/groups';
import { wakeupApi } from '../api/wakeup';
import { userApi } from '../api/user';
import type { Task, Group, WakeupCheck, NotificationLog } from '../types';

/**
 * ダッシュボードに必要なデータ取得と状態管理を行うカスタムフックである．
 */
export const useDashboardData = (userId: string, groupId: string) => {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [group, setGroup] = useState<Group | null>(null);
  const [activeWakeup, setActiveWakeup] = useState<WakeupCheck | null>(null);
  const [groupWakeups, setGroupWakeups] = useState<WakeupCheck[]>([]);
  const [notificationLogs, setNotificationLogs] = useState<NotificationLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  // 初期状態での UI 崩れを防ぐため，デフォルトは false（トークンあり）としておく
  const [serverTokenMissing, setServerTokenMissing] = useState<boolean>(false);

  const fetchData = useCallback(async () => {
    if (!userId || !groupId) return;
    
    try {
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
        const currentGroup = groupsRes.value?.find(g => g.id === groupId);
        if (currentGroup) setGroup(currentGroup);
      }

      if (wakeupsRes.status === 'fulfilled') {
        const pendingWakeup = wakeupsRes.value?.find(w => w.status === 'pending');
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
    } catch (err: unknown) {
      if (err instanceof Error) {
        console.error("useDashboardData: Fetch error:", err.message);
      }
      setError("データの取得中に予期せぬエラーが発生した．");
    } finally {
      setLoading(false);
    }
  }, [groupId, userId]);

  return {
    tasks,
    group,
    activeWakeup,
    groupWakeups,
    notificationLogs,
    loading,
    error,
    serverTokenMissing,
    setServerTokenMissing,
    fetchData,
    setLoading
  };
};
