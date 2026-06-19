import { useState, useCallback } from 'react';
import { notificationApi } from '../api/notifications';
import { urlBase64ToUint8Array, handle } from '../utils/helpers';
import { AxiosError } from 'axios';

/**
 * Web Push 通知の購読と状態管理を行うカスタムフックである．
 */
export const useWebPush = (userId: string, groupId: string) => {
  // Notification API に非対応のブラウザ（一部のスマホブラウザなど）でのクラッシュを防ぐガード処理
  const [notifPermission, setNotifPermission] = useState<NotificationPermission>(
    typeof Notification !== 'undefined' ? Notification.permission : 'default'
  );

  const handleEnableNotifications = useCallback(async (onSuccess: () => void) => {
    if (typeof Notification === 'undefined') {
      alert("お使いのブラウザはプッシュ通知に対応していません．スマホの場合は「ホーム画面に追加」してからお試しください．");
      return;
    }
    const [permission, pErr] = await handle(Notification.requestPermission());
    if (pErr) {
      console.error("useWebPush: Permission request failed", pErr.message);
      alert("通知の設定に失敗しました．");
      return;
    }

    setNotifPermission(permission);
    
    if (permission === 'granted') {
      const [registration, rErr] = await handle(navigator.serviceWorker.getRegistration());
      if (rErr) {
        console.error("useWebPush: Get registration failed", rErr.message);
        return;
      }

      if (!registration) {
        const [, regErr] = await handle(navigator.serviceWorker.register('/sw.js'));
        if (regErr) {
          console.error("useWebPush: Registration failed", regErr.message);
          return;
        }
      }
      
      const [reg, readyErr] = await handle(navigator.serviceWorker.ready);
      if (readyErr) {
        console.error("useWebPush: Service worker not ready", readyErr.message);
        return;
      }
      
      // 既存の購読があれば解除
      const [existingSub, subErr] = await handle(reg.pushManager.getSubscription());
      if (!subErr && existingSub) {
        await handle(existingSub.unsubscribe());
      }

      // 公開鍵の変換
      const vapidPublicKey = 'BDj40a3LAnB-Tyemxggm-wYyuHbE_kadO6CqX6u-Ewyrkqi5ypr-txXJO7jflV_4VGa47paZU7DX_-0OPZy6Bx8';
      const convertedVapidKey = urlBase64ToUint8Array(vapidPublicKey);

      const [subscription, sErr] = await handle(reg.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: convertedVapidKey
      }));

      if (sErr) {
        console.error("useWebPush: Subscription failed", sErr.message);
        alert("通知の購読に失敗しました．");
        return;
      }
      
      const [, apiErr] = await handle(notificationApi.subscribe(userId, subscription));
      if (apiErr) {
        console.error("useWebPush: API subscribe failed", apiErr.message);
        return;
      }

      onSuccess();
      alert("通知が有効になりました！AI からのリマインドが届くようになります．");
    }
  }, [userId]);

  const handleSilentResubscribe = useCallback(async (onSuccess: () => void) => {
    if (typeof Notification === 'undefined' || Notification.permission !== 'granted') return;

    const [reg, rErr] = await handle(navigator.serviceWorker.ready);
    if (rErr) return;
    
    const [existingSub, subErr] = await handle(reg.pushManager.getSubscription());
    if (!subErr && existingSub) {
      await handle(existingSub.unsubscribe());
    }

    const vapidPublicKey = 'BDj40a3LAnB-Tyemxggm-wYyuHbE_kadO6CqX6u-Ewyrkqi5ypr-txXJO7jflV_4VGa47paZU7DX_-0OPZy6Bx8';
    const convertedVapidKey = urlBase64ToUint8Array(vapidPublicKey);

    const [subscription, sErr] = await handle(reg.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey: convertedVapidKey
    }));

    if (sErr) return;
    
    const [, apiErr] = await handle(notificationApi.subscribe(userId, subscription));
    if (!apiErr) {
      onSuccess();
      console.log("useWebPush: Silent resubscribe successful.");
    }
  }, [userId]);

  const handleSendTestNotification = useCallback(async (aiCharacter: string, onTokenMissing: () => void) => {
    const [, err] = await handle(notificationApi.sendTestNotification(userId, aiCharacter, groupId));
    if (err) {
      const axiosErr = err as AxiosError<{error: string}>; 
      if (axiosErr.response?.data?.error?.includes("トークンが存在しない")) {
        onTokenMissing();
      }
      alert(`テスト通知の送信に失敗しました：${axiosErr.response?.data?.error || err.message}`);
      return;
    }
    alert("テスト通知を送信しました．数秒以内に届くはずです．");
  }, [userId, groupId]);

  return {
    notifPermission,
    handleEnableNotifications,
    handleSilentResubscribe,
    handleSendTestNotification
  };
};
