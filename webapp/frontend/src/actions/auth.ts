"use server";

import { cookies } from "next/headers";

export async function clearCookieAction() {
  try {
    const cookieStore = await cookies();
    cookieStore.set("session_id", "", {
      httpOnly: true,
      secure: process.env.NODE_ENV === "production",
      sameSite: "strict",
      path: "/",
      expires: new Date(0),
    });
    return { success: true };
  } catch (error) {
    console.error("Cookie clear error:", error);
    return { success: false };
  }
}
