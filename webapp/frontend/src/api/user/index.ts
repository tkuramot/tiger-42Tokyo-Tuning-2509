import axios from "axios";

type User = {
  UserId: number;
  Name: string;
};

export async function login(
  userName: string,
  password: string
): Promise<User | null> {
  const { data: user } = await axios.post<User>("/api/login", {
    user_name: userName,
    password: password,
  });

  console.log("--- ユーザー情報の取得に成功:", user);
  return user;
}
