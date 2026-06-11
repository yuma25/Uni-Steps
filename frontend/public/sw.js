/**
 * Uni-Steps Web Push Service Worker
 * バックグラウンドで通知を受け取り，表示するためのプログラムである．
 */

self.addEventListener('push', function(event) {
  console.log('[Service Worker] Push Received.');
  
  let data = {
    title: 'Uni-Steps',
    body: '新しい通知があります．',
    url: '/'
  };

  if (event.data) {
    try {
      // JSON として解析を試みる
      const json = event.data.json();
      data.title = json.title || data.title;
      data.body = json.body || json.message || data.body;
      data.url = json.url || data.url;
    } catch (e) {
      // JSON でなければ，そのままテキストとして扱う
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
    }
  };

  event.waitUntil(
    self.registration.showNotification(data.title, options)
  );
});

self.addEventListener('notificationclick', function(event) {
  console.log('[Service Worker] Notification click Received.');
  event.notification.close();
  event.waitUntil(
    clients.openWindow(event.notification.data.url)
  );
});
