import { test, expect } from "@playwright/test";

test.describe("Auth (smoke)", () => {
  test("ログイン成功 → /product へ遷移", async ({ page }) => {
    await page.goto("/login");
    await page.getByLabel("UserName").fill("user100");
    await page.getByLabel("Password").fill("password");
    await page.getByRole("button", { name: "ログイン" }).click();

    await expect(page).toHaveURL(/\/product$/);

    // 任意: セッションクッキーが乗ったか軽く確認
    const cookies = await page.context().cookies();
    expect(cookies.some((c) => c.name === "session_id")).toBeTruthy();
  });

  test("誤ったパスワード → エラーメッセージ表示＆/loginに留まる", async ({
    page,
  }) => {
    await page.goto("/login");
    await page.getByLabel("UserName").fill("user001");
    await page.getByLabel("Password").fill("wrongs_password");
    await page.getByRole("button", { name: "ログイン" }).click();

    await expect(page).toHaveURL(/\/login/);
    // アプリ側で <div role="alert">失敗...</div> のようにしておくと堅い
    await expect(page.getByTestId("password-error")).toContainText(
      /メールアドレスまたはパスワードが正しくありません。/
    );
  });
});
