/**
 * Base64 URL 形式の文字列を Uint8Array に変換するヘルパー関数である．
 * Web Push の VAPID 鍵をブラウザが認識できる形式にするために必要である．
 */
export function urlBase64ToUint8Array(base64String: string) {
  const padding = '='.repeat((4 - base64String.length % 4) % 4);
  const base64 = (base64String + padding)
    .replace(/-/g, '+')
    .replace(/_/g, '/');

  const rawData = window.atob(base64);
  const outputArray = new Uint8Array(rawData.length);

  for (let i = 0; i < rawData.length; ++i) {
    outputArray[i] = rawData.charCodeAt(i);
  }
  return outputArray;
}

/**
 * ISO8601 形式の文字列を日本の日付表示形式に変換する．
 */
export function formatDate(isoString: string): string {
  const date = new Date(isoString);
  if (date.getFullYear() <= 1) return "期限未定";
  return date.toLocaleString('ja-JP', { 
    month: 'short', 
    day: 'numeric', 
    hour: '2-digit', 
    minute: '2-digit' 
  });
}

/**
 * 非同期処理を Go 言語風の [結果, エラー] 形式で処理するためのラッパーである．
 * try-catch のネストを深くせずに，エラーを「値」として扱うことができる．
 */
export async function handle<T>(promise: Promise<T>): Promise<[T, null] | [null, Error]> {
  try {
    const data = await promise;
    return [data, null];
  } catch (err: unknown) {
    if (err instanceof Error) return [null, err];
    return [null, new Error(String(err))];
  }
}
