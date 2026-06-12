import client from './client';
import type { User } from '../types';

/**
 * ユーザー情報に関連する API 通信を担当するモジュールである．
 */
export const userApi = {
  /**
   * 指定した ID のユーザー情報を取得する．
   */
  getMe: async (userId: string): Promise<User> => {
    const resp = await client.get<User>(`/api/users/${userId}`);
    return resp.data;
  },
};
