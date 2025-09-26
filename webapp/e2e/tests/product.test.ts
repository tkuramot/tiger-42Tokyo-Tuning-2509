// tests/product.test.ts
import { test, expect } from "@playwright/test";
import sampleData from "../tests/sampleData/productsKeywordsHitsCount.json";
import { fail } from "assert";
test.describe("Product page (real backend over Docker network)", () => {
  test("検索のリクエスト/レスポンス内容とUI反映→注文送信(alert)", async ({
    page,
  }) => {
    // 0) sampleData/products.jsonからランダムに1件ピックアップして
    //    その名前の一部を今回のテストの検索語として使う
    const testSampleData =
      sampleData[Math.floor(Math.random() * sampleData.length)];

    // 1) ログイン → /product
    await page.goto("/login");
    await page.getByLabel(/UserName|ユーザー名/i).fill("user001");
    await page.getByLabel(/Password|パスワード/i).fill("password");
    await page.getByRole("button", { name: /ログイン|Sign in/i }).click();
    await expect(page).toHaveURL(/\/product$/);
    await expect(page.getByRole("heading", { name: "商品一覧" })).toBeVisible();

    // 2) 初回ロードは UI で待つ（gridcell が1つ以上出る）
    const gridcells = page.locator('[role="gridcell"]');
    await expect(gridcells).not.toHaveCount(0, { timeout: 15000 });

    const paginationText = await page
      .locator(".MuiTablePagination-displayedRows")
      .textContent(); // paginationText の中身は "1–20 of 1000000"

    if (!paginationText) {
      fail("Pagination text not found");
    }

    const totalCountString = paginationText.split("of ")[1]; // "1000000"
    const totalCount = parseInt(totalCountString, 1000000); // 1000000

    // 3) 検索語を入れて「検索」を押す（クリック前に waitForResponse を仕掛ける）
    await page
      .getByPlaceholder("商品名または説明で検索...")
      .fill(testSampleData.keyword);
    const [searchRes] = await Promise.all([
      page.waitForResponse(
        (r) =>
          r.url().includes("/api/v1/product") &&
          r.request().method() === "POST" &&
          r.status() === 200
      ),
      page.getByRole("button", { name: "検索" }).click(),
    ]);

    // (1) 送信ペイロードに検索語が入っていること
    const req = searchRes.request();
    const sent =
      (req.postDataJSON && req.postDataJSON()) ||
      JSON.parse(req.postData() || "{}");
    expect(String(sent.search || "")).toContain(testSampleData.keyword);

    // (2) 応答JSONの各アイテムが検索語にマッチしていること
    const json = await searchRes.json();
    const items: any[] = Array.isArray(json?.data) ? json.data : [];
    expect(items.length).toBeGreaterThan(0);
    for (const it of items.slice(0, 20)) {
      const hay = `${it?.name ?? ""} ${it?.description ?? ""}`;
      expect(hay).toContain(testSampleData.keyword);
    }

    // (3) UIにも反映：返ってきた先頭アイテムの名前が表示されている
    const firstReturnedName: string = String(items[0]?.name ?? "").trim();
    await expect(
      page
        .locator('[role="gridcell"][data-field="name"]')
        .filter({ hasText: firstReturnedName })
        .first()
    ).toBeVisible();

    // 1. まずはテキスト全体を取得する
    const searchPaginationText = await page
      .locator(".MuiTablePagination-displayedRows")
      .textContent();

    // searchPaginationText の中身は "1–20 of 1000000"
    if (!searchPaginationText) {
      fail("Pagination text not found");
    }

    // 2. テキストを "of " で分割し、最後の部分を取得
    const searchTotalCountString = searchPaginationText.split("of ")[1]; // "1000000"

    // 3. 文字列を数値に変換
    const searchTotalCount = parseInt(searchTotalCountString, 10); // 1000000

    console.log("総件数:", searchTotalCount);

    // 4. 検証する
    expect(searchTotalCount).toBe(testSampleData.local_hits);

    // 5) 数量入力→注文送信（alert 検証 & 200応答）
    await page.locator('input[placeholder="0"]').first().fill("2");

    const dialogPromise = new Promise<void>((resolve) => {
      page.once("dialog", async (d) => {
        expect(d.type()).toBe("alert");
        expect(d.message()).toBe("注文が正常に送信されました。");
        await d.accept();
        resolve();
      });
    });

    const orderResP = page.waitForResponse(
      (r) =>
        r.url().includes("/api/v1/product/post") &&
        r.request().method() === "POST"
    );
    await page.getByRole("button", { name: "注文送信" }).click();
    const orderRes = await orderResP;
    expect(orderRes.status()).toBe(201);
    await dialogPromise;
  });
});
