package main

import (
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
)

type QueueMessage struct {
	LoadTestID string `json:"load_test_id"`
	TeamID     string `json:"team_id"`
	TargetIP   string `json:"target_ip"`
}

type ScoreEntity struct {
	aztables.Entity
	TeamID string
	Score  int
}

type RequestEntity struct {
	aztables.Entity
	Status    string    `json:"Status"`
	CreatedAt time.Time `json:"CreatedAt"`
	UpdatedAt time.Time `json:"UpdatedAt"`
}
