# 【第4回】Uni-Stepsで学ぶクリーンアーキテクチャ：フロントエンド編

Uni-Stepsの構造を題材にしたクリーンアーキテクチャ解説連載の第4回である．

これまでバックエンド（Go）におけるレイヤー分離と依存性のコントロールを見てきた．
今回は視点を**フロントエンド（TypeScript / React）**に移す．「フロントエンドにおけるクリーンアーキテクチャ」の解釈と，UIコンポーネントが通信や複雑な状態管理から解放され，スマートに保たれる設計手法について解説する．

---

## 1. フロントエンドにおける「クリーンアーキテクチャ」の定義

バックエンド同様，フロントエンドでも「UIの見た目の変更」と「API通信ロジックの変更」，「データの持ち方（状態管理）」を混ざり合わせないことが，保守性の高いコードを書くための鍵である．

Uni-Stepsのフロントエンド（`frontend/src/`）は，以下のようなレイヤー構造で整理されている．

```
frontend/src
 ├── types/       # 【Domain (実体)】 ビジネスデータ構造（型定義）
 ├── api/         # 【Infrastructure (インフラ)】 APIクライアント（技術的詳細）
 ├── hooks/       # 【UseCase (ロジック)】 カスタムフックによる状態・ビジネスロジック管理
 └── pages/       # 【Presenter (表現)】 画面描画とイベント検知（UIコンポーネント）
```

---

## 2. 各レイヤーのコードと役割

クリーンアーキテクチャの目的は，「何がどこに書いてあるか」を整理し，プログラムの変更に強い構造を作ることである．フロントエンドの各レイヤーにおける役割と，実際のソースコード例を以下に示す．

---

### ① Domain (Types)：データの定義 (`types/`)

**【初学者向け解説】**
Domain（ドメイン）は，アプリケーションにおける**「ルールブック（憲法）」**である．
ここでは，システムが扱うデータそのものの形（TypeScriptの型定義）を定義する．

例えば「タスク（Task）とは何か？」という定義は，画面の見た目（CSS）やAPIの通信ライブラリ（Axios）が何であろうと変わらない最も本質的な情報である．この本質的な定義を最初にカチッと決めておくことで，他のすべてのレイヤーがこのルールに従って安全にプログラムを書くことができる．

```typescript
// frontend/src/types/index.ts (抜粋)

// 課題（タスク）のデータ構造を定義するエンティティ型 (Domain)
export interface Task {
  id: string;                      // 課題を一意に識別するためのID (UUID)
  group_id: string;                // 課題が所属するグループのID
  external_id: string;             // 外部LMS (Classroomなど) と連携するためのID
  title: string;                   // 課題のタイトル
  deadline: string;                // 提出期限日時 (日付文字列)
  source: 'manual' | 'classroom';  // 登録元 (手動作成か，Classroomから自動取得か)
  creator_id: string;              // この課題を作成したユーザーのID
  user_progress: TaskUserProgress[]; // グループ内メンバー全員 of 進捗状況リスト
  created_at: string;              // データベースに登録された日時
  updated_at: string;              // 最後に更新された日時
}

// ユーザーごとの課題進捗（完了しているかどうか）を表す型 (Domain)
export interface TaskUserProgress {
  task_id: string;                 // 対象となる課題のID
  user_id: string;                 // 対象となるユーザーのID
  is_completed: boolean;           // 課題が完了しているかどうかのフラグ
  updated_at: string;              // 進捗が最後に更新された日時
}
```

---

### ② Infrastructure (API)：通信の具象化 (`api/`)

**【初学者向け解説】**
Infrastructure（インフラ）は，外部のサーバーと通信を行うための**「窓口（郵便局）」**である．
ここではAxiosなどのライブラリを使って，実際にWeb APIを呼び出す処理を記述する．

「どのURL（パス）にアクセスしてデータを取ってくるか」や「どのような形式でリクエストを送るか」という**具体的な通信方法（技術的詳細）**をこのレイヤーだけに閉じ込める．これにより，将来的に通信ライブラリをAxiosから標準の `fetch` に切り替えたり，APIのURLが変更されたりしても，他のレイヤー（UIやビジネスロジック）を1行も書き換えることなく，このファイルだけの修正で済むようになる．

```typescript
// frontend/src/api/tasks.ts (抜粋)

import client from './client'; // 共通のBaseURLや認証設定が済んだAxiosクライアント (Infrastructure)
import type { Task } from '../types';

// 課題に関するAPI通信処理をまとめたオブジェクト (Infrastructure)
export const taskApi = {
  // 指定されたグループに紐づく課題一覧をサーバーから取得する
  listGroupTasks: async (groupId: string): Promise<Task[]> => {
    // GET /api/groups/{groupId}/tasks にリクエストを送信
    const resp = await client.get<Task[]>(`/api/groups/${groupId}/tasks`);
    return resp.data; // サーバーから返ってきた実際のデータ（Task配列）を返却
  },

  // 課題の「完了 / 未完了」の状態を切り替えるリクエストを送信する
  toggleTaskCompletion: async (taskId: string, userId: string): Promise<void> => {
    // PATCH /api/tasks/{taskId}/toggle-completion にリクエストを送信
    await client.patch(`/api/tasks/${taskId}/toggle-completion`, {
      user_id: userId, // リクエストボディにユーザーIDを含めて送信
    });
  }
};
```

---

### ③ UseCase (Hooks)：状態とビジネスプロセスの管理 (`hooks/`)

**【初学者向け解説】**
UseCase（ユースケース）は，画面（UI）と通信（インフラ）を仲介する**「現場監督（オーケストラ指揮者）」**である．Reactでは「カスタムフック（Custom Hooks）」という形で実装する．

「画面を開いたときに，課題一覧とグループ一覧を同時に通信して持ってくる」「通信を待っている間はローディング中という状態にする」「もし通信に失敗したらエラーメッセージを画面に伝える」といった，アプリケーションとしての**「具体的な動作の流れ（ユースケース）」**を管理する．
見た目（UI）のコードからこの複雑なロジックを分離することで，アプリの動きのルールだけをここで集中してテスト・修正できるようになる．

```typescript
// frontend/src/hooks/useDashboardData.ts (抜粋)

import { useState, useCallback } from 'react';
import { taskApi } from '../api/tasks';
import { groupApi } from '../api/groups';
import type { Task, Group } from '../types';
import { handle } from '../utils/helpers'; // 非同期処理のエラーハンドリングを簡潔にするユーティリティ

// ダッシュボードに必要なデータと，その取得ロジックを一元管理するカスタムフック (UseCase)
export const useDashboardData = (userId: string, groupId: string) => {
  const [tasks, setTasks] = useState<Task[]>([]); // 画面に表示する課題一覧の状態管理
  const [group, setGroup] = useState<Group | null>(null); // 現在選択中のグループ情報の状態管理
  const [loading, setLoading] = useState(true); // API通信中のローディング状態を管理するフラグ
  const [error, setError] = useState<string | null>(null); // 通信エラーが発生した際のエラーメッセージ管理

  // APIから最新データをロードする処理定義
  const fetchData = useCallback(async () => {
    if (!userId || !groupId) return; // 必要なIDがない場合は処理をスキップ
    setError(null); // エラー表示を初期化

    // 課題一覧と所属グループ一覧のAPIリクエストを並行実行 (Infrastructureの呼び出し)
    const [results, err] = await handle(Promise.allSettled([
      taskApi.listGroupTasks(groupId),
      groupApi.listMyGroups(userId),
    ]));

    if (err) {
      setError("データの取得中に予期せぬエラーが発生した．");
      setLoading(false);
      return;
    }

    const [taskDataRes, groupsRes] = results;

    // 課題データの取得結果を反映
    if (taskDataRes.status === 'fulfilled') {
      setTasks(taskDataRes.value || []);
    }
    // グループデータの取得結果を反映
    if (groupsRes.status === 'fulfilled') {
      const currentGroup = groupsRes.value?.find(g => g.id === groupId);
      if (currentGroup) setGroup(currentGroup);
    }
    setLoading(false); // 読み込み完了状態に変更
  }, [groupId, userId]);

  // UIコンポーネントが画面表示や操作の検知で必要とするデータ・関数のみを返却する
  return {
    tasks,      // 課題の配列データ
    group,      // グループのオブジェクトデータ
    loading,    // ローディング中かどうかの真偽値
    error,      // エラー文字列（またはnull）
    fetchData,  // 画面更新用の再読み込み関数
  };
};
```

---

### ④ Presenter (Pages/Components)：UIとユーザー操作の検知 (`pages/`)

**【初学者向け解説】**
Presenter（プレゼンター）は，ユーザーの目に直接触れる**「看板（UIコンポーネント）」**である．
Reactにおける各画面やパーツのコンポーネントファイルがこれに該当する．

このレイヤーの唯一の役割は，**「現場監督（UseCase）から渡されたデータを画面に綺麗に表示すること」**と**「ボタンクリックなどのユーザー操作を検知して現場監督に伝えること」**だけである．
ここにはAPIのURLも，通信エラーの判定ロジックも一切書かない．そうすることで，コンポーネントはHTMLとCSS（またはデザインシステム）の調整だけに集中することができ，シンプルで読みやすいコードが維持される．

```tsx
// frontend/src/pages/Dashboard.tsx (イメージ例)

import React, { useEffect } from 'react';
import { useDashboardData } from '../hooks/useDashboardData';

// ダッシュボード画面を描画するコンポーネント (Presenter)
export const Dashboard: React.FC<{ userId: string; groupId: string }> = ({ userId, groupId }) => {
  // UseCaseレイヤー（カスタムフック）を呼び出し，描画に必要なデータや関数を取り出す
  const { tasks, group, loading, error, fetchData } = useDashboardData(userId, groupId);

  // コンポーネントが画面に表示されたタイミング（マウント時）にデータを読み込む
  useEffect(() => {
    fetchData();
  }, [fetchData]);

  // 通信中または通信エラー発生時の画面表示制御（ガード節による早期描画分け）
  if (loading) return <div>読み込み中...</div>;
  if (error) return <div>エラー: {error}</div>;

  return (
    <div>
      {/* 取得したグループ名を表示 */}
      <h1>{group?.name} のダッシュボード</h1>
      <ul>
        {/* 課題一覧のデータをループ処理で1件ずつリスト表示する */}
        {tasks.map(task => (
          <li key={task.id}>
            <span>{task.title}</span>
            {/* ボタン押下時にトグル処理を実行（UseCaseへイベント通知） */}
            <button onClick={() => /* 完了切り替えのUseCase（フック）を実行する */ {}}>
              トグル
            </button>
          </li>
        ))}
      </ul>
    </div>
  );
};
```

---

## 3. フロントエンドで分離するメリット

1.  **コンポーネントがシンプルになる**:
    JSX（HTML）の中に `axios.get` や `try-catch` といった非同期通信処理が一切現れないため，デザイン（見た目）のコーディングに集中できる．
2.  **モックを使った表示確認が容易**:
    `api/` の中身を一時的にローカルのダミーデータ（Mock）を返すように書き換えるだけで，バックエンドが未完成でも画面のデザイン確認や状態のテストが行える．
3.  **状態管理ライブラリの載せ替えに強い**:
    もし将来，状態管理を useState から Redux や Zustand，Jotai に載せ替えることになっても，変更するのは `hooks/` レイヤーだけであり，UIコンポーネントの大部分はそのまま維持できる．

---

## 4. まとめ

第4回では，フロントエンドにおけるクリーンアーキテクチャの適用について解説した．

*   **Domain (Types)**: ビジネスモデルの型定義．
*   **Infrastructure (API)**: 通信仕様（URLパスやパラメータ）を閉じたモジュール．
*   **UseCase (Hooks)**: UI状態の管理と通信を組み合わせる「現場監督」．
*   **Presenter (Pages/Components)**: レンダリングとユーザーイベントの検知に徹するUI．

次回は最終回である．
**「第5回：連動編：データの一生（エンドツーエンドで追うクリーンアーキテクチャの連動）」**として，フロントエンドからバックエンドのデータベースまで，1つのリクエストがすべてのレイヤーをどのように通過していくのか，その全体像をシーケンス図と共にトレースする．
