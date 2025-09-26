package main

import (
	"context"
	"encoding/json"
	"log"
)

func enqueueToDLQ(dlqPayload map[string]string) {
	if dlqClient == nil {
		return
	}
	b, _ := json.Marshal(dlqPayload)
	if _, err := dlqClient.EnqueueMessage(context.TODO(), string(b), nil); err != nil {
		log.Printf("[FAIL-DROP] DLQへの保存に失敗: %v", err)
	} else {
		log.Printf("[FAIL-DROP] DLQに保存しました: team=%s loadTestID=%s", dlqPayload["team_id"], dlqPayload["load_test_id"])
	}
}
