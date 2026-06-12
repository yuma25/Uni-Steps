import client from './client';
import type { WakeupCheck } from '../types';

/**
 * 起床確認に関連する API 通信を担当するモジュールである．
 */
export const wakeupApi = {
  /**
   * 新しい起床確認を予約する．
   */
  request: async (params: { user_id: string; group_id: string; target_time: string; grace_minutes: number }): Promise<WakeupCheck> => {
    const resp = await client.post<WakeupCheck>('/api/wakeup/request', params);
    return resp.data;
  },

  /**
   * 起床を報告（チェックイン）する．
   */
  checkin: async (userId: string): Promise<void> => {
    await client.post('/api/wakeup/checkin', { user_id: userId });
  },

  /**
   * ユーザーの現在の（進行中の）起床確認を取得する．
   */
  getActive: async (userId: string): Promise<WakeupCheck[]> => {
    const resp = await client.get<WakeupCheck[]>(`/api/wakeup/active?user_id=${userId}`);
    return resp.data;
  },

  /**
   * グループ内の全メンバーの進行中の起床確認を取得する．
   */
  getActiveByGroup: async (groupId: string): Promise<WakeupCheck[]> => {
    const resp = await client.get<WakeupCheck[]>(`/api/groups/${groupId}/wakeups/active`);
    return resp.data;
  },

  /**
   * 進行中の起床見守りをキャンセルする．
   */
  cancel: async (userId: string): Promise<void> => {
    await client.delete(`/api/wakeup/cancel?user_id=${userId}`);
  },
};
