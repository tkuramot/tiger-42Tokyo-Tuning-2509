package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"
)

// Item は商品の情報を保持する構造体
// `json:"..."` はJSONに変換する際のキー名を指定します
type Item struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

// Order は注文情報を保持する構造体
type Order struct {
	Items []Item `json:"items"`
}

func main2() {
	// 乱数のシード（種）を設定。毎回違う結果を得るために必要です。
	rand.Seed(time.Now().UnixNano())

	// --- 生成ルールの設定 ---
	totalOrders := 100 // 生成する注文の総数

	minItemsPerOrder := 1 // 注文ごとの最小アイテム数
	maxItemsPerOrder := 5 // 注文ごとの最大アイテム数

	minProductID := 1    // 最小商品ID
	maxProductID := 1000 // 最大商品ID

	minQuantity := 1  // 最小数量
	maxQuantity := 10 // 最大数量
	// -------------------------

	// 生成した全注文を格納するスライス
	orderList := []Order{}

	// 指定した総数だけループして、注文データを作成
	for i := 0; i < totalOrders; i++ {
		// この注文に含めるアイテム数をランダムに決定
		numItems := rand.Intn(maxItemsPerOrder-minItemsPerOrder+1) + minItemsPerOrder

		// 現在の注文のアイテムを格納するスライス
		currentItems := []Item{}

		// 決まったアイテム数だけループして、ランダムな商品を生成
		for j := 0; j < numItems; j++ {
			item := Item{
				ProductID: rand.Intn(maxProductID-minProductID+1) + minProductID,
				Quantity:  rand.Intn(maxQuantity-minQuantity+1) + minQuantity,
			}
			currentItems = append(currentItems, item)
		}

		// 完成した商品リストで注文オブジェクトを作成し、最終リストに追加
		order := Order{Items: currentItems}
		orderList = append(orderList, order)
	}

	// 作成したデータ構造を、見やすい形式(インデント付き)のJSONバイト配列に変換
	jsonData, err := json.MarshalIndent(orderList, "", "  ")
	if err != nil {
		// エラーハンドリング
		log.Fatalf("JSONの変換に失敗しました: %s", err)
	}

	// バイト配列を文字列に変換して出力
	fmt.Println(string(jsonData))
}
