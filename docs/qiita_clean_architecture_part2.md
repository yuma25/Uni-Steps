# 【第2回】Uni-Stepsで学ぶクリーンアーキテクチャ：Domain と UseCase

Uni-Stepsの構造を題材にしたクリーンアーキテクチャ解説連載の第2回である．

前回は，クリーンアーキテクチャの基本思想と，プロジェクト全体のフォルダ構成を整理した．
今回は，バックエンド（Go）における最深部，すなわちアプリケーションの心臓部である **Domain（ドメイン）レイヤー** と **UseCase（ユースケース）レイヤー** について，実際のコードを交えながら詳しく解説する．

---

## 1. Domainレイヤー：ビジネスのルールと契約

クリーンアーキテクチャの同心円の最も内側にあるのが **Domainレイヤー** である．
このレイヤーには，アプリのビジネスルールそのものを記述する．

### 特徴：外部への依存が「ゼロ」
DomainレイヤーのGoコード（`domain/` 配下）を確認すると，**インポートしているパッケージに外部ライブラリ（GORMやEchoなど）が一切含まれていない** ことに気づく．

標準パッケージ（`context` や `time`）や，UUID生成などのユーティリティ以外は，純粋なGoのコードだけで書かれている．これにより，データベースやWebフレームワークがどう変更されようとも，Domainレイヤーのコードは影響を受けない．

### 実例①：ドメインモデル (domain/task.go)
タスク（課題）を表す構造体定義を示す．

```go
package domain

import (
	"time"
)

type TaskSource string

const (
	SourceManual    TaskSource = "manual"
	SourceClassroom TaskSource = "classroom"
)

// Task は課題を表すドメインモデルである．
type Task struct {
	ID           string             `json:"id" gorm:"primaryKey"`
	GroupID      string             `json:"group_id" gorm:"index:idx_group_external,unique"`
	ExternalID   string             `json:"external_id" gorm:"index:idx_group_external,unique"` // 外部LMSのIDなど
	Title        string             `json:"title"`
	Deadline     time.Time          `json:"deadline"`
	Source       TaskSource         `json:"source"`
	CreatorID    string             `json:"creator_id"`
	UserProgress []*TaskUserProgress `json:"user_progress" gorm:"foreignKey:TaskID"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
}
```
*(※注: GORMのタグ `gorm:"..."` は構造体に付与されているが，これはGoの特性上，モデル定義とマイグレーションの利便性を両立するための妥協案として一般的に許容されるアプローチである．)*

### 実例②：リポジトリのインターフェース (domain/repository.go)
Domainレイヤーの最も重要な役割の一つが，**「インターフェース（契約）」の定義** である．
「データベースからどうやってデータを取得・保存するか」という具体的な方法は記述せず，**「どのような機能が必要か」**というルールだけを定義する．

```go
package domain

import (
	"context"
	"time"
)

// TaskRepository は課題データの永続化に関する約束事（インターフェース）である．
// DDD（ドメイン駆動設計）では，具体的なDBの実装（SQLなど）はここに書かず，
// ビジネスロジックが必要とする「機能」だけを定義する．
type TaskRepository interface {
	Save(ctx context.Context, task *Task) error                            				// 課題の保存・更新
	FindByID(ctx context.Context, id string) (*Task, error)                				// IDによる課題検索
	FindByExternalID(ctx context.Context, externalID string) (*Task, error) 			// 外部LMSのIDによる課題検索
	FindByGroupID(ctx context.Context, groupID string) ([]*Task, error)     			// グループ内の課題一覧取得
	FindApproachingDeadlines(ctx context.Context, until time.Time) ([]*Task, error) 	// 期限が近い未完了課題の取得
	Delete(ctx context.Context, id string) error                           				// 課題の削除
}
```

この `TaskRepository` というインターフェースを定義しておくことで，後述するUseCaseレイヤーは，具体的なデータベースの存在（PostgreSQLなのか，あるいはインメモリのマップなのか）を意識せずにビジネスロジックを実装できるようになる．

---

## 2. UseCaseレイヤー：ビジネスロジックの組み立て

Domainレイヤーの1つ外側にあるのが **UseCaseレイヤー** である．
ユーザーの「手動でタスクを登録したい」「完了状態を切り替えたい」といった具体的な要求（ユースケース）に対して，ビジネスルールをどう適用するかを記述する．

### 特徴：インターフェースに依存する（依存性の逆転）
UseCaseレイヤー（`usecase/` 配下）は，Domainレイヤーで定義されたインターフェース（`domain.TaskRepository` など）を介してデータを操作する．

これにより，UseCaseは**「具象的なDB（GORMなど）」に直接依存するのではなく，「抽象的なインターフェース（Domain）」に依存する**ことになる．これがオブジェクト指向設計における **「依存関係逆転の原則 (Dependency Inversion Principle: DIP)」** の実践である．

### 実例：TaskUsecase (usecase/task_uc.go)
実際のタスク管理ビジネスロジックのコードを示す．

```go
package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yuma25/Uni-Steps/domain"
)

// TaskUsecase は課題管理に関するビジネスロジックを担当する構造体である．
type TaskUsecase struct {
	taskRepo  domain.TaskRepository   // domainのインターフェースに依存
	groupRepo domain.GroupRepository  // domainのインターフェースに依存
	aiService domain.AIService        // AIサービスのインターフェース
	scheduler domain.SchedulerService // スケジューラのインターフェース
}

// NewTaskUsecase は TaskUsecase の新しいインスタンスを生成する（コンストラクタによるDI）．
func NewTaskUsecase(tr domain.TaskRepository, gr domain.GroupRepository, ai domain.AIService, sch domain.SchedulerService) *TaskUsecase {
	return &TaskUsecase{
		taskRepo:  tr,
		groupRepo: gr,
		aiService: ai,
		scheduler: sch,
	}
}

// RegisterManualTask は UI から直接入力された情報に基づいて課題を登録するユースケースである．
func (uc *TaskUsecase) RegisterManualTask(ctx context.Context, task *domain.Task) (*domain.Task, error) {
	// 1. ビジネスルールの検証
	task.Source = domain.SourceManual
	if task.Title == "" {
		return nil, fmt.Errorf("タイトルは必須である")
	}
	if task.ID == "" {
		task.ID = uuid.New().String()
	}
	if task.ExternalID == "" {
		task.ExternalID = task.ID
	}

	// 2. データの保存（インターフェースを介した呼び出し）
	if err := uc.taskRepo.Save(ctx, task); err != nil {
		return nil, fmt.Errorf("手動タスクの保存に失敗した： %w", err)
	}

	// 3. 関連するビジネスロジック（リマインドの自動予約）の実行
	group, _ := uc.groupRepo.FindByID(ctx, task.GroupID)
	if group != nil && !task.Deadline.IsZero() && task.Deadline.After(time.Now()) {
		for _, up := range task.UserProgress {
			if !up.IsCompleted {
				for _, interval := range group.RemindIntervals {
					// スケジューラ（インターフェース）を通じてリマインド予約
					_ = uc.scheduler.ScheduleTaskRemind(ctx, task, up.UserID, interval, group.AICharacter, task.Deadline.Add(-time.Duration(interval)*time.Minute))
				}
			}
		}
	}
	return task, nil
}
```

### コラム：ここで return された値はどこへ行くのか？
現段階（第2回）では，UseCaseを呼び出す側のコードが登場していないため，この `return nil, err` や `return task, nil` がどこに戻っていくのか疑問に思うかもしれない．
この関数の呼び出し元は，1つ外側のレイヤーである **Interfaces（Handler）レイヤー** である．
Handlerは，このUseCaseを呼び出して結果（データやエラー）を受け取り，HTTPレスポンス（成功なら `201 Created`，エラーなら `500 Internal Server Error` などのJSON）に変換してフロントエンドへ返却する役割を持つ．これについては，次回の第3回で詳しく解説する．

### なぜこの設計が優れているのか？

1.  **ユニットテスト（単体テスト）が書きやすい**:
    `domain.TaskRepository` をモック（Mock）に差し替えるだけで，本物のデータベースを起動することなく `RegisterManualTask` のビジネスロジック（タイトルが空のときにエラーになるか，リマインドが正しく計算されるかなど）を100%テストできる．
2.  **ビジネスロジックの可読性が高い**:
    「データベースへのトランザクション開始」や「SQLの組み立て」といったコードが混ざらないため，「タスクを登録して，グループ情報を取得し，期限前ならリマインドを予約する」という**ビジネスの流れそのもの**がコードから一目で理解できる．

---

## 3. まとめ

第2回では，クリーンアーキテクチャの「本質」である最内周の2つのレイヤーを学んだ．

*   **Domain**: ビジネスの「データ構造（モデル）」と「操作の約束事（インターフェース）」を定義する．外部への依存はゼロ．
*   **UseCase**: Domainのインターフェースを組み合わせて，「ビジネスロジックの流れ」を組み立てる．依存関係逆転の原則（DIP）により，データベースの具象実装には直接依存しない．

次回は，この「本質」を動かすための「詳細」にあたる部分を解説する．
**「第3回：バックエンド編②：Interfaces と Infrastructure 〜フレームワークやDBと対話する外界のレイヤー〜」**として，Echoを使ったHTTPリクエストの受け口と，GORMを使った実際のDBアクセスのコードを紐解いていく．
