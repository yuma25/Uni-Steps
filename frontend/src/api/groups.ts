import client from './client';
import type { Group } from '../types';

/**
 * グループ（部屋）管理に関連する API 通信を担当するモジュールである．
 */
export const groupApi = {
  /**
   * ユーザーが所属しているグループ一覧を取得する．
   */
  listMyGroups: async (userId: string): Promise<Group[]> => {
    // 本来はバックエンドに専用のエンドポイントが必要であるが，
    // 現時点ではプロトタイプ用として全取得などの代替手段を想定する．
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
   * 既存のグループに招待コード等で参加する（将来的な拡張用）．
   */
  joinGroup: async (groupId: string, userId: string): Promise<void> => {
    await client.post(`/api/groups/${groupId}/join`, {
      user_id: userId,
    });
  },

  /**
   * 外部 LMS (Google Classroom 等) からコース一覧を同期・取得する．
   */
  syncLMSGroups: async (userId: string): Promise<Group[]> => {
    const resp = await client.post<Group[]>(`/api/users/${userId}/groups/sync`);
    return resp.data;
  },

  /**
   * 外部 LMS から利用可能な（アーカイブされていない）コース一覧を取得する．
   */
  fetchAvailableLMSCourses: async (userId: string): Promise<Group[]> => {
    const resp = await client.get<Group[]>(`/api/users/${userId}/lms/courses`);
    return resp.data;
  },

  /**
   * 特定の部屋に LMS コースを紐付ける．
   */
  linkLMSCourse: async (groupId: string, lmsCourseId: string): Promise<void> => {
    await client.patch(`/api/groups/${groupId}/link-lms`, {
      lms_course_id: lmsCourseId,
    });
  },
};
