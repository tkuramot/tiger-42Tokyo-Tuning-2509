// internal/score/score.go
package score

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
)

// k6 の JSON から finalScore を作る
func ComputeFinalScoreFromK6(rawPath string) (int, error) {
	f, err := os.Open(rawPath)
	if err != nil {
		// 旧スクリプトは summary が無ければ 0 を返して終了していた
		// ここも互換で 0 を返す
		return 0, fmt.Errorf("open raw json: %w", err)
	}
	defer f.Close()

	raw, err := io.ReadAll(f)
	if err != nil {
		return 0, fmt.Errorf("read raw json: %w", err)
	}

	var data struct {
		Metrics map[string]struct {
			Values map[string]float64 `json:"values"`
		} `json:"metrics"`
		// RootGroup json.RawMessage `json:"root_group,omitempty"` // 未使用
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return 0, fmt.Errorf("unmarshal json: %w", err)
	}

	// 取り出しヘルパ
	get := func(name string) float64 {
		if m, ok := data.Metrics[name]; ok {
			if c, ok := m.Values["count"]; ok {
				return c
			}
		}
		return 0
	}

	// counts
	bench_uj_success_count := get("bench_uj_success_count")
	bench_robot_success_count := get("bench_robot_success_count")

	// 旧パイプラインは整数を期待（Atoi）していたので四捨五入して int に
	return int(math.Round(bench_uj_success_count + bench_robot_success_count)), nil
}

// 互換のための summary.json 出力（不要なら呼び出し削除OK）
func WriteSummaryJSON(path string, score int) error {
	b, _ := json.MarshalIndent(map[string]int{"finalScore": score}, "", "  ")
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return err
	}
	return nil
}
