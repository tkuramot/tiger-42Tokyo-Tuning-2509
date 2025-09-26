package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue"
)

func deleteMessageSafe(queueClient *azqueue.QueueClient, messageID, popReceipt *string) {
	if messageID == nil || popReceipt == nil {
		return
	}
	if _, err := queueClient.DeleteMessage(context.TODO(), *messageID, *popReceipt, nil); err != nil {
		log.Printf("キューメッセージの削除に失敗: %v", err)
	} else {
		log.Printf("キューメッセージを削除しました: msgID=%s", *messageID)
	}
}

func processMessage(msg *azqueue.DequeuedMessage) {
	var data QueueMessage
	if err := parseQueueMessage(msg, &data); err != nil {
		return
	}

	skip, err := isDuplicateOrSkip(data)
	if err != nil {
		log.Printf("ロック確認中にエラー: %v", err)
		return
	}
	if skip {
		log.Printf("他のリクエストが既に進行中のためスキップ: team=%s", data.TeamID)
		return
	}

	if err := acquireLock(data); err != nil {
		log.Printf("ロック取得失敗: %v", err)
		return
	}

	score, err := runLoadTest(data.TargetIP)
	if err != nil {
		markFailedAndDrop(msg.MessageID, msg.PopReceipt, data.TeamID, data.LoadTestID, data.TargetIP, err.Error())
		return
	}

	// postScore に統一（baseURL, authToken は main.go のグローバル）
	if err := postScore(baseURL, authToken, data.LoadTestID, data.TeamID, score); err != nil {
		markFailedAndDrop(msg.MessageID, msg.PopReceipt, data.TeamID, data.LoadTestID, data.TargetIP, err.Error())
		return
	}

	_ = releaseLock(data)

	if _, err := queueClient.DeleteMessage(context.TODO(), *msg.MessageID, *msg.PopReceipt, nil); err != nil {
		log.Printf("メッセージ削除失敗: %v", err)
	}
	log.Printf("全行程完了: LoadTestID=%s", data.LoadTestID)
}

func parseQueueMessage(msg *azqueue.DequeuedMessage, data *QueueMessage) error {
	if err := json.Unmarshal([]byte(*msg.MessageText), data); err != nil {
		log.Printf("不正なメッセージを削除: %v", err)
		queueClient.DeleteMessage(context.TODO(), *msg.MessageID, *msg.PopReceipt, nil)
		return err
	}
	log.Printf("タスク受信: LoadTestID=%s, TeamID=%s", data.LoadTestID, data.TeamID)
	// 受信時点で running に遷移させ、SSE の実行中カウントに載せる
	if err := updateRequestStatus(data.TeamID, data.LoadTestID, "running"); err != nil {
		log.Printf("warn: mark running failed: %v", err)
	}
	return nil
}
