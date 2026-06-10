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
};
