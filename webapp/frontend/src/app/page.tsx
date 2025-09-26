import { redirect } from "next/navigation";
import { cookies } from "next/headers";

// サーバーサイドでユーザーの認証状態をチェックする非同期関数
async function checkUserAuthentication() {
  const cookieStore = await cookies();
  const isLoggedIn = cookieStore.has("session_id");
  // TODO：セッションの有効性確認が必要

  return isLoggedIn;
}

export default async function RootPage() {
  const isLoggedIn = await checkUserAuthentication();
  if (isLoggedIn) {
    redirect("/product");
  } else {
    redirect("/login");
  }
}
