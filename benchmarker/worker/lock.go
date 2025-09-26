package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
)

// OData用の簡易エスケープ：シングルクォートを2つに
func escapeODataString(s string) string {
	if s == "" {
		return s
	}
	return strings.ReplaceAll(s, "'", "''")
}

// 同一 Team に 1時間以内の waiting/running が1件でもあれば true。
// 自分自身（同一 LoadTestID）のみなら false。
func isDuplicateOrSkip(data QueueMessage) (bool, error) {
	hourAgo := time.Now().UTC().Add(-1 * time.Hour)
	filter := fmt.Sprintf(
		"PartitionKey eq '%s' and (Status eq 'waiting' or Status eq 'running') and UpdatedAt ge datetime'%s'",
		escapeODataString(data.TeamID),
		hourAgo.Format(time.RFC3339),
	)
	top := int32(1)
	selectStr := "RowKey,Status,UpdatedAt"

	pager := requestsTableClient.NewListEntitiesPager(&aztables.ListEntitiesOptions{
		Filter: &filter,
		Top:    &top,
		Select: &selectStr, // SDK版によっては []string 型。合わなければこの行を削除でもOK
	})

	for pager.More() {
		page, err := pager.NextPage(context.TODO())
		if err != nil {
			log.Printf("ページ取得失敗: %v", err)
			return false, err
		}
		for _, entity := range page.Entities {
			var existing RequestEntity
			// SDKの型差異対策：[]byte でも map[string]any でも処理
			switch v := any(entity).(type) {
			case []byte:
				if err := json.Unmarshal(v, &existing); err != nil {
					continue
				}
			case map[string]any:
				if b, _ := json.Marshal(v); b != nil {
					_ = json.Unmarshal(b, &existing)
				}
			default:
				continue
			}
			if existing.RowKey != data.LoadTestID {
				return true, nil
			}
		}
	}
	return false, nil
}

// running で作成/更新（Upsert）
func acquireLock(data QueueMessage) error {
	now := time.Now().UTC()
	e := RequestEntity{
		Entity: aztables.Entity{
			PartitionKey: data.TeamID,
			RowKey:       data.LoadTestID,
		},
		Status:    "running",
		UpdatedAt: now,
	}
	b, _ := json.Marshal(e)
	if _, err := requestsTableClient.UpsertEntity(context.TODO(), b, nil); err != nil {
		return fmt.Errorf("テーブル更新失敗: %w", err)
	}
	return nil
}

func releaseLock(data QueueMessage) error {
	return updateRequestStatus(data.TeamID, data.LoadTestID, "finished")
}

func updateRequestStatus(teamID, rowKey, status string) error {
	now := time.Now().UTC()
	e := RequestEntity{
		Entity: aztables.Entity{
			PartitionKey: teamID,
			RowKey:       rowKey,
		},
		Status:    status,
		UpdatedAt: now,
	}
	b, _ := json.Marshal(e)
	_, err := requestsTableClient.UpsertEntity(context.TODO(), b, nil)
	return err
}

func markFailedAndDrop(messageID, popReceipt *string, teamID, loadTestID, targetIP, errMsg string) {
	log.Printf("[FAIL-DROP] team=%s loadTestID=%s targetIP=%s err=%s", teamID, loadTestID, targetIP, errMsg)

	// DLQへ退避
	if dlqClient != nil {
		payload := map[string]string{
			"load_test_id": loadTestID,
			"team_id":      teamID,
			"target_ip":    targetIP,
			"error":        errMsg,
		}
		if b, err := json.Marshal(payload); err == nil {
			if _, err := dlqClient.EnqueueMessage(context.TODO(), string(b), nil); err != nil {
				log.Printf("[FAIL-DROP] DLQ送信失敗: %v", err)
			}
		}
	}

	// ステータス更新
	if err := updateRequestStatus(teamID, loadTestID, "failed"); err != nil {
		log.Printf("[FAIL-DROP] ステータス更新失敗: %v", err)
	}

	// メッセージ削除
	if messageID != nil && popReceipt != nil {
		if _, err := queueClient.DeleteMessage(context.TODO(), *messageID, *popReceipt, nil); err != nil {
			log.Printf("[FAIL-DROP] DeleteMessage失敗: %v", err)
		}
	}
}
