/**
 * Uni-Steps Web Push Service Worker
 * バックグラウンドで通知を受け取り，表示するためのプログラムである．
 */

self.addEventListener('push', function(event) {
  console.log('[Service Worker] Push event received:', event);
  
  let data = {
    title: 'Uni-Steps',
    body: '新しい通知があります．',
    url: '/'
  };

  if (event.data) {
    try {
      const text = event.data.text();
      console.log('[Service Worker] Raw payload:', text);
      
      // JSON として解析を試みる
      const json = JSON.parse(text);
      data.title = json.title || data.title;
      data.body = json.body || json.message || data.body;
      data.url = json.url || data.url;
    } catch (e) {
      console.warn('[Service Worker] Payload is not JSON, treating as text.');
      data.body = event.data.text();
    }
  }

  const options = {
    body: data.body,
    icon: '/favicon.svg',
    badge: '/favicon.svg',
    vibrate: [100, 50, 100],
    data: {
      url: data.url
    },
    tag: 'uni-steps-notification', // 同じタグの通知は上書きする（通知欄を汚さない）
    renotify: true
  };

  event.waitUntil(
    self.registration.showNotification(data.title, options)
      .then(() => console.log('[Service Worker] Notification shown successfully.'))
      .catch(err => console.error('[Service Worker] Failed to show notification:', err))
  );
});

self.addEventListener('notificationclick', function(event) {
  console.log('[Service Worker] Notification click received.');
  event.notification.close();
  
  event.waitUntil(
    clients.matchAll({ type: 'window' }).then(windowClients => {
      // すでに開いているページがあれば，そこをアクティブにする
      for (let i = 0; i < windowClients.length; i++) {
        const client = windowClients[i];
        if (client.url === event.notification.data.url && 'focus' in client) {
          return client.focus();
        }
      }
      // なければ新しく開く
      if (clients.openWindow) {
        return clients.openWindow(event.notification.data.url);
      }
    })
  );
});
