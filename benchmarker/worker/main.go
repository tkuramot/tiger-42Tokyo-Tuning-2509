package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue"
)

// === グローバル：Azure クライアント ===
var (
	queueClient         *azqueue.QueueClient
	requestsTableClient *aztables.Client
	dlqClient           *azqueue.QueueClient
	baseURL             string
	authToken           string

	visibilityTimeoutSec int32 = 300 // メッセージの可視性タイムアウト（秒）
	dequeueBatchSize     int32 = 1   // まとめて取る件数（ワーカー単体なら1でOK）
)

func main() {
	initializeClients()

	// SIGINT/SIGTERM で安全に終了
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Println("負荷試験Workerのポーリングを開始します...")
	pollQueue(ctx)
}

func initializeClients() {
	connStr := os.Getenv("AZURE_STORAGE_CONNECTION_STRING")
	if connStr == "" {
		log.Fatal("AZURE_STORAGE_CONNECTION_STRING を設定してください。")
	}
	baseURL = os.Getenv("INTERMEDIATE_BASE_URL")
	if baseURL == "" {
		log.Fatal("INTERMEDIATE_BASE_URL を設定してください。")
	}
	authToken = os.Getenv("WORKER_AUTH_TOKEN")
	if authToken == "" {
		log.Fatal("WORKER_AUTH_TOKEN を設定してください。")
	}

	// 任意：環境変数で調整可能に
	if v := os.Getenv("QUEUE_VISIBILITY_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			visibilityTimeoutSec = int32(d.Seconds())
			if visibilityTimeoutSec < 30 {
				visibilityTimeoutSec = 30
			}
		}
	}
	if v := os.Getenv("DEQUEUE_BATCH_SIZE"); v != "" {
		if n, err := parseInt32(v); err == nil && n > 0 && n <= 32 {
			dequeueBatchSize = n
		}
	}

	queueName := "fttokyo-loadtest-requests"
	dlqName := queueName + "-dlq"
	tableName := "fttokyoloadtestlocks"

	var err error
	queueClient, err = azqueue.NewQueueClientFromConnectionString(connStr, queueName, nil)
	if err != nil {
		log.Fatalf("Queue Client作成失敗: %v", err)
	}

	dlqClient, err = azqueue.NewQueueClientFromConnectionString(connStr, dlqName, nil)
	if err != nil {
		log.Fatalf("DLQ Client作成失敗: %v", err)
	}
	if _, err := dlqClient.Create(context.Background(), nil); err != nil {
		var respErr *azcore.ResponseError
		if !errors.As(err, &respErr) || respErr.ErrorCode != "QueueAlreadyExists" {
			log.Fatalf("DLQ作成失敗: %v", err)
		}
	}

	tableSvcClient, err := aztables.NewServiceClientFromConnectionString(connStr, nil)
	if err != nil {
		log.Fatalf("Table Service Client作成失敗: %v", err)
	}
	requestsTableClient = tableSvcClient.NewClient(tableName)
}

// ------------------------------------------
// 無限ループでQueue監視（キャンセル対応＋バックオフ）
func pollQueue(ctx context.Context) {
	backoff := time.Second
	for {
		select {
		case <-ctx.Done():
			log.Println("シャットダウン: ポーリングを終了します")
			return
		default:
		}

		// Dequeue オプション（可視性タイムアウトと件数）
		resp, err := queueClient.DequeueMessage(ctx, &azqueue.DequeueMessageOptions{
			VisibilityTimeout: &visibilityTimeoutSec,
		})
		if err != nil {
			log.Printf("キューの取得に失敗: %v", err)
			time.Sleep(backoff)
			if backoff < 15*time.Second {
				backoff *= 2
			}
			continue
		}
		backoff = time.Second // 成功したのでリセット

		if resp.Messages == nil || len(resp.Messages) == 0 {
			// 空振り時は軽くスリープ（Storage Queue にはロングポーリングがない）
			time.Sleep(2 * time.Second)
			continue
		}

		// まとめて取った場合も1件ずつ処理
		for _, msg := range resp.Messages {
			// ここでコンテキストを子にしても良い（タイムアウトなど）
			processMessage(msg)
		}
	}
}

// 文字列→int32（環境変数用）
func parseInt32(s string) (int32, error) {
	n, err := time.ParseDuration(s)
	if err == nil {
		return int32(n.Seconds()), nil
	}
	// "300" 形式も許す
	var i int
	_, err = fmt.Sscanf(s, "%d", &i)
	return int32(i), err
}
