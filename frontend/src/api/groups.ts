import client from './client';
import type { Group, NotificationLog } from '../types';

/**
 * グループ（部屋）管理に関連する API 通信を担当するモジュールである．
 */
export const groupApi = {
  /**
   * ユーザーが所属しているグループ一覧を取得する．
   */
  listMyGroups: async (userId: string): Promise<Group[]> => {
    const resp = await client.get<Group[]>(`/api/users/${userId}/groups`);
    return resp.data;
  },

  /**
   * 新しいグループ（部屋）を作成する．
   */
  createGroup: async (name: string, ownerId: string): Promise<Group> => {
    const resp = await client.post<Group>('/api/groups', {
      name: name,
      owner_id: ownerId,
    });
    return resp.data;
  },

  /**
   * 招待コードを用いて既存の部屋に参加する．
   */
  joinGroup: async (inviteCode: string, userId: string): Promise<Group> => {
    const resp = await client.post<Group>('/api/groups/join', {
      invite_code: inviteCode,
      user_id: userId,
    });
    return resp.data;
  },

  /**
   * 外部 LMS (Google Classroom 等) からコース一覧を同期・取得する．
   */
  syncLMSGroups: async (userId: string): Promise<Group[]> => {
    const resp = await client.post<Group[]>(`/api/users/${userId}/groups/sync`);
    return resp.data;
  },
  /**
   * 部屋の設定（名称，リマインド間隔，AI 性格，LINE 連携等）を更新する．
   */
  updateSettings: async (
    groupId: string, 
    name: string,
    intervals: number[], 
    userId: string, 
    aiCharacter: string, 
    lineToken: string, 
    lineGroupId: string,
    morningTime: string,
    eveningTime: string
  ): Promise<void> => {
    await client.patch(`/api/groups/${groupId}/settings`, {
      name: name,
      remind_intervals: intervals,
      user_id: userId,
      ai_character: aiCharacter,
      line_channel_token: lineToken,
      line_group_id: lineGroupId,
      summary_morning_time: morningTime,
      summary_evening_time: eveningTime,
    });
  },

  /**
   * 部屋のオーナー権限を別のユーザーに譲渡する．
   */
  transferOwnership: async (groupId: string, currentOwnerId: string, newOwnerId: string): Promise<void> => {
    await client.put(`/api/groups/${groupId}/owner`, {
      current_owner_id: currentOwnerId,
      new_owner_id: newOwnerId,
    });
  },

  /**
   * 部屋（グループ）を完全に削除する．オーナーのみ可能．
   */
  deleteGroup: async (groupId: string, userId: string): Promise<void> => {
    await client.delete(`/api/groups/${groupId}?user_id=${userId}`);
  },

  /**
   * 指定した部屋から退出する．
   */
  leaveGroup: async (groupId: string, userId: string): Promise<void> => {
    await client.delete(`/api/groups/${groupId}/users/${userId}`);
  },

  /**
   * 指定した部屋の通知履歴を取得する．
   */
  listNotifications: async (groupId: string): Promise<NotificationLog[]> => {
    const resp = await client.get<NotificationLog[]>(`/api/groups/${groupId}/notifications`);
    return resp.data;
  },
};
