import { useState, useCallback } from 'react';
import { notificationApi } from '../api/notifications';
import { urlBase64ToUint8Array } from '../utils/helpers';

/**
 * Web Push 通知の購読と状態管理を行うカスタムフックである．
 */
export const useWebPush = (userId: string, groupId: string) => {
  const [notifPermission, setNotifPermission] = useState<NotificationPermission>(Notification.permission);

  const handleEnableNotifications = useCallback(async (onSuccess: () => void) => {
    try {
      const permission = await Notification.requestPermission();
      setNotifPermission(permission);
      
      if (permission === 'granted') {
        const registration = await navigator.serviceWorker.register('/sw.js');
        
        // 既存の購読があれば解除
        const existingSub = await registration.pushManager.getSubscription();
        if (existingSub) await existingSub.unsubscribe();

        // 公開鍵の変換
        const vapidPublicKey = 'BDj40a3LAnB-Tyemxggm-wYyuHbE_kadO6CqX6u-Ewyrkqi5ypr-txXJO7jflV_4VGa47paZU7DX_-0OPZy6Bx8';
        const convertedVapidKey = urlBase64ToUint8Array(vapidPublicKey);

        const subscription = await registration.pushManager.subscribe({
          userVisibleOnly: true,
          applicationServerKey: convertedVapidKey
        });
        
        await notificationApi.subscribe(userId, subscription);
        onSuccess();
        alert("通知が有効になりました！AI からのリマインドが届くようになります．");
      }
    } catch (err) {
      console.error("useWebPush: Notification error:", err);
      alert("通知の設定に失敗しました．");
    }
  }, [userId]);

  const handleSendTestNotification = useCallback(async (aiCharacter: string, onTokenMissing: () => void) => {
    try {
      await notificationApi.sendTestNotification(userId, aiCharacter, groupId);
      alert("テスト通知を送信しました．数秒以内に届くはずです．");
    } catch (err: any) {
      if (err.response?.data?.error?.includes("トークンが存在しない")) {
        onTokenMissing();
      }
      alert(`テスト通知の送信に失敗しました：${err.response?.data?.error || err.message}`);
    }
  }, [userId, groupId]);

  return {
    notifPermission,
    handleEnableNotifications,
    handleSendTestNotification
  };
};
