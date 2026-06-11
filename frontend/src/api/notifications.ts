import client from './client';

/**
 * 通知に関連する API 通信を担当するモジュールである．
 */
export const notificationApi = {
  /**
   * Web Push の購読情報をサーバーに保存する．
   */
  subscribe: async (userId: string, subscription: PushSubscription): Promise<void> => {
    await client.post('/api/notifications/subscribe', {
      user_id: userId,
      subscription: JSON.stringify(subscription),
    });
  },
};
