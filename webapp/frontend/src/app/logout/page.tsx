"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { clearCookieAction } from "@/actions/auth";

export default function LogoutPage() {
  const [isLoggedOut, setIsLoggedOut] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // このページが最初にクライアントで表示されたときに、一度だけ実行
  useEffect(() => {
    const performLogout = async () => {
      const result = await clearCookieAction();
      if (result.success) {
        setIsLoggedOut(true);
      } else {
        setError("ログアウト処理中にエラーが発生しました。");
      }
    };

    performLogout();
  }, []); // 空の依存配列[]は「マウント時に一度だけ実行」を意味する

  if (error) {
    return (
      <div style={{ textAlign: "center", marginTop: "50px", color: "red" }}>
        <h1>エラー</h1>
        <p>{error}</p>
        <Link href="/">ホームへ戻る</Link>
      </div>
    );
  }

  if (!isLoggedOut) {
    return (
      <div style={{ textAlign: "center", marginTop: "50px" }}>
        <h1>ログアウト処理中...</h1>
      </div>
    );
  }

  return (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        height: "80vh",
      }}
    >
      <h1>ログアウトしました</h1>
      <Link
        href="/login"
        style={{
          display: "inline-block",
          padding: "12px 24px",
          color: "white",
          backgroundColor: "#0070f3",
          borderRadius: "8px",
          textDecoration: "none",
          fontWeight: "bold",
        }}
      >
        ログイン画面へ戻る
      </Link>
    </div>
  );
}
