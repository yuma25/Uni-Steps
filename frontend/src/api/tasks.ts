import client from './client';
import type { Task } from '../types';

/**
 * 課題管理に関連する API 通信を担当するモジュールである．
 */
export const taskApi = {
  /**
   * グループに紐づく課題一覧を取得する．
   */
  listGroupTasks: async (groupId: string): Promise<Task[]> => {
    const resp = await client.get<Task[]>(`/api/groups/${groupId}/tasks`);
    return resp.data;
  },

  /**
   * 手動で新しい課題を登録する．
   */
  createManualTask: async (task: Partial<Task>): Promise<Task> => {
    const resp = await client.post<Task>('/api/tasks/manual', task);
    return resp.data;
  },

  /**
   * 外部 LMS (Google Classroom等) から課題を同期する．
   */
  syncTasks: async (userId: string, groupId: string): Promise<{ message: string; tasks: Task[] }> => {
    const resp = await client.post('/api/tasks/sync', {
      user_id: userId,
      group_id: groupId,
    });
    return resp.data;
  },

  /**
   * 課題の完了状態を切り替える．
   */
  toggleTaskCompletion: async (taskId: string, userId: string): Promise<void> => {
    await client.patch(`/api/tasks/${taskId}/toggle-completion`, {
      user_id: userId,
    });
  },
  /**
   * 課題の情報を更新する．
   */
  updateTask: async (taskId: string, task: Partial<Task>): Promise<Task> => {
    const resp = await client.put<Task>(`/api/tasks/${taskId}`, task);
    return resp.data;
  },
};
