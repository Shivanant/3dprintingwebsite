import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";
import {
  AuthPayload,
  User,
  login as apiLogin,
  register as apiRegister,
  refresh as apiRefresh,
  me as apiMe,
  resetPassword as apiResetPassword,
  forgotPassword as apiForgotPassword,
} from "@/lib/authClient";

type AuthState = {
  user?: User;
  accessToken?: string;
  refreshToken?: string;
  accessExpiresAt?: string;
  refreshExpiresAt?: string;
  loading: boolean;
  initialized: boolean;
  login(email: string, password: string): Promise<void>;
  register(name: string, email: string, password: string): Promise<void>;
  refresh(): Promise<void>;
  logout(): void;
  bootstrap(): Promise<void>;
  forgotPassword(email: string): Promise<void>;
  resetPassword(token: string, newPassword: string): Promise<void>;
};

const tokenStorage = {
  set(payload: AuthPayload) {
    if (typeof window === "undefined") return;
    window.localStorage.setItem("accessToken", payload.accessToken);
    window.localStorage.setItem("refreshToken", payload.refreshToken);
    window.localStorage.setItem("userId", payload.user.id);
  },
  clear() {
    if (typeof window === "undefined") return;
    window.localStorage.removeItem("accessToken");
    window.localStorage.removeItem("refreshToken");
    window.localStorage.removeItem("userId");
  },
  get() {
    if (typeof window === "undefined") return { access: undefined, refresh: undefined, userId: undefined };
    return {
      access: window.localStorage.getItem("accessToken") ?? undefined,
      refresh: window.localStorage.getItem("refreshToken") ?? undefined,
      userId: window.localStorage.getItem("userId") ?? undefined,
    };
  },
};

const memoryStorage = {
  getItem: () => null,
  setItem: () => {},
  removeItem: () => {},
};

export const useAuth = create<AuthState>()(
  persist(
    (set, get) => ({
      user: undefined,
      accessToken: undefined,
      refreshToken: undefined,
      accessExpiresAt: undefined,
      refreshExpiresAt: undefined,
      loading: false,
      initialized: false,
      async login(email, password) {
        set({ loading: true });
        try {
          const payload = await apiLogin({ email, password });
          setAuthState(set, payload);
        } finally {
          set({ loading: false });
        }
      },
      async register(name, email, password) {
        set({ loading: true });
        try {
          const payload = await apiRegister({ name, email, password });
          setAuthState(set, payload);
        } finally {
          set({ loading: false });
        }
      },
      async refresh() {
        const { user, refreshToken } = get();
        if (!user || !refreshToken) return;
        const payload = await apiRefresh(user.id, refreshToken);
        setAuthState(set, payload);
      },
      logout() {
        tokenStorage.clear();
        set({
          user: undefined,
          accessToken: undefined,
          refreshToken: undefined,
          accessExpiresAt: undefined,
          refreshExpiresAt: undefined,
        });
      },
      async bootstrap() {
        if (get().initialized) return;
        const tokens = tokenStorage.get();
        if (!tokens.access || !tokens.refresh || !tokens.userId) {
          set({ initialized: true });
          return;
        }
        try {
          const payload = await apiMe();
          set({
            user: payload,
            accessToken: tokens.access,
            refreshToken: tokens.refresh,
            initialized: true,
          });
        } catch (err) {
          tokenStorage.clear();
          set({ initialized: true });
        }
      },
      async forgotPassword(email: string) {
        await apiForgotPassword(email);
      },
      async resetPassword(token: string, newPassword: string) {
        const payload = await apiResetPassword(token, newPassword);
        setAuthState(set, payload);
      },
    }),
    {
      name: "auth-store",
      storage: createJSONStorage(() => (typeof window === "undefined" ? memoryStorage : window.localStorage)),
      partialize: (state) => ({
        user: state.user,
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
        accessExpiresAt: state.accessExpiresAt,
        refreshExpiresAt: state.refreshExpiresAt,
      }),
    }
  )
);

function setAuthState(
  set: (update: Partial<AuthState> | ((state: AuthState) => Partial<AuthState>)) => void,
  payload: AuthPayload
) {
  tokenStorage.set(payload);
  set({
    user: payload.user,
    accessToken: payload.accessToken,
    refreshToken: payload.refreshToken,
    accessExpiresAt: payload.accessExpiresAt,
    refreshExpiresAt: payload.refreshExpiresAt,
    initialized: true,
  });
}
