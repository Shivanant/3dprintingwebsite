import api from "./apiClient";

export type User = {
  id: string;
  email: string;
  name: string;
  role: string;
  avatarUrl?: string | null;
};

export type AuthPayload = {
  user: User;
  accessToken: string;
  accessExpiresAt: string;
  refreshToken: string;
  refreshExpiresAt: string;
};

export async function register(payload: { email: string; password: string; name: string }) {
  const { data } = await api.post<AuthPayload>("/auth/register", payload);
  return data;
}

export async function login(payload: { email: string; password: string }) {
  const { data } = await api.post<AuthPayload>("/auth/login", payload);
  return data;
}

export async function refresh(userId: string, refreshToken: string) {
  const { data } = await api.post<AuthPayload>("/auth/refresh", {
    userId,
    refreshToken,
  });
  return data;
}

export async function forgotPassword(email: string) {
  await api.post("/auth/forgot-password", { email });
}

export async function resetPassword(token: string, newPassword: string) {
  const { data } = await api.post<AuthPayload>("/auth/reset-password", {
    token,
    newPassword,
  });
  return data;
}

export async function me() {
  const { data } = await api.get<User>("/auth/me");
  return data;
}
