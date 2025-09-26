package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
	"worker/score"
)

func runLoadTest(targetIP string) (int, error) {
	log.Printf("負荷試験開始: TargetIP=%s", targetIP)
	ts := time.Now().Format("20060102150405")

	rawPath := filepath.Join("/app", "scores", fmt.Sprintf("raw-data-%s.json", targetIP+ts))
	// 互換用（必要なら）summary.json も書く
	summaryPath := filepath.Join("/app", "scores", fmt.Sprintf("summary-%s.json", targetIP+ts))

	if err := os.MkdirAll(filepath.Dir(rawPath), 0o755); err != nil {
		return 0, fmt.Errorf("mkdir scores: %w", err)
	}

	cmd := exec.Command("./run_k6.sh", targetIP+ts)
	cmd.Dir = "/app"
	cmd.Env = append(os.Environ(),
		"TARGET_IP="+targetIP,
	)
	cmd.Stderr = os.Stderr

	// 出力は読まない（スコアはGoで計算する）
	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("run_k6.sh failed: %w", err)
	}

	result, err := score.ComputeFinalScoreFromK6(rawPath)
	if err != nil {
		// 旧挙動に合わせ、取れなければ 0 を返す（致命ではない）
		log.Printf("warn: failed to compute score: %v", err)
		return 0, nil
	}

	// 互換のために summary.json も出しておく（不要なら削除OK）
	if err := score.WriteSummaryJSON(summaryPath, result); err != nil {
		log.Printf("warn: write summary: %v", err)
	}
	return result, nil
}
