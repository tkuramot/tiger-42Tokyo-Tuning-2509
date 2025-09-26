// cmd/runner/main.go
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"worker/score"
)

func main() {
	target := flag.String("target-ip", "", "Target IP or hostname (e.g., 127.0.0.1)")
	// k6 が --summary-export で書く先（入力）
	sum := flag.String("summary", "", "Path where k6 --summary-export writes (input)")
	raw := flag.String("raw", "", "Path where k6 --out json writes (input)")
	// finalScore だけを書き出す先（出力）※別ファイルにする
	final := flag.String("final", "", "Path to final-only JSON (output)")
	flag.Parse()

	// TARGET_IP: フラグ優先 → 環境変数 → デフォルト
	tgt := *target
	if tgt == "" {
		tgt = os.Getenv("TARGET_IP")
	}
	if tgt == "" {
		tgt = "127.0.0.1"
	}

	ts := time.Now().Format("20060102150405")
	// ファイル名用に安全化（: / \ 空白 を避ける）
	slug := strings.NewReplacer(":", "-", "/", "-", "\\", "-", " ", "_").Replace(tgt)

	// デフォルトの出力先（/app 固定）
	if *sum == "" {
		*sum = filepath.Join("/app", "scores", fmt.Sprintf("summary-%s-%s.json", slug, ts))
	}
	if *raw == "" {
		*raw = filepath.Join("/app", "scores", fmt.Sprintf("raw-data-%s-%s.json", slug, ts))
	}
	if *final == "" {
		*final = filepath.Join("/app", "scores", fmt.Sprintf("final-%s-%s.json", slug, ts))
	}
	if err := os.MkdirAll(filepath.Dir(*sum), 0o755); err != nil {
		log.Fatalf("mkdir (summary dir): %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(*final), 0o755); err != nil {
		log.Fatalf("mkdir (final dir): %v", err)
	}

	// run_k6.sh は --summary-export "$SUMMARY_FILE_PATH" に吐く
	cmd := exec.Command("./run_k6.sh", slug+"-"+ts)
	cmd.Dir = "/app"
	cmd.Env = append(os.Environ(),
		"TARGET_IP="+tgt,
	)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("run_k6.sh failed: %v", err)
	}

	// 解析対象は k6 の summary（counts入り）
	scoreVal, err := score.ComputeFinalScoreFromK6(*raw)
	if err != nil {
		scoreVal = 0
	}

	// counts を壊さないよう、finalScore は別ファイルへ
	if err := score.WriteSummaryJSON(*final, scoreVal); err != nil {
		log.Printf("warn: write final summary: %v", err)
	}

	// 標準出力に最終スコア（run.sh で拾いやすい）
	fmt.Println(scoreVal)
}
