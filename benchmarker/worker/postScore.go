package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// 中間サーバにスコア登録（環境変数でTLS検証を制御）
func postScore(baseURL, token, loadTestID, teamID string, score int) error {
	payload := map[string]interface{}{
		"load_test_id": loadTestID,
		"team_id":      teamID,
		"score":        score,
	}
	b, _ := json.Marshal(payload)

	insecure := os.Getenv("TLS_INSECURE_SKIP_VERIFY") == "1"
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
	}
	client := &http.Client{Transport: tr, Timeout: 15 * time.Second}

	req, err := http.NewRequest(
		"POST",
		strings.TrimRight(baseURL, "/")+"/internal/register-score",
		bytes.NewReader(b),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return nil
}

// 旧 submitScore は、baseURL/token のハードコード＆常時 Insecure なので削除しました。
// 呼び出し側は postScore(baseURL, authToken, ...) を使ってください。
