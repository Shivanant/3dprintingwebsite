"use client";

import Link from "next/link";
import { FormEvent, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { useAuth } from "@/store/auth";

export default function LoginForm() {
  const router = useRouter();
  const params = useSearchParams();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const { login, loading } = useAuth();

  const redirectTo = params.get("redirect") || "/account";

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    try {
      await login(email, password);
      router.push(redirectTo);
    } catch (err: any) {
      setError(err?.response?.data?.error || "Failed to sign in. Check your credentials.");
    }
  }

  return (
    <div className="max-w-md mx-auto bg-white rounded-2xl border shadow-sm p-8 space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Welcome back</h1>
        <p className="text-sm text-gray-500">Sign in to manage your prints, cart, and orders.</p>
      </div>

      {error && (
        <div className="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
          {error}
        </div>
      )}

      <form className="space-y-4" onSubmit={onSubmit}>
        <label className="block text-sm">
          Email
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="mt-1 w-full rounded-lg border px-3 py-2 focus:outline-none focus:ring-2 focus:ring-gray-900/10"
            required
          />
        </label>

        <label className="block text-sm">
          Password
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="mt-1 w-full rounded-lg border px-3 py-2 focus:outline-none focus:ring-2 focus:ring-gray-900/10"
            required
          />
        </label>

        <button
          type="submit"
          disabled={loading}
          className="w-full rounded-lg bg-black text-white py-2.5 font-medium hover:bg-gray-900 transition disabled:opacity-60"
        >
          {loading ? "Signing in..." : "Sign in"}
        </button>
      </form>

      <div className="flex flex-col text-sm text-gray-500 gap-2">
        <Link href="/forgot-password" className="hover:text-gray-800 transition">
          Forgot your password?
        </Link>
        <div>
          No account?{" "}
          <Link href={`/register?redirect=${encodeURIComponent(redirectTo)}`} className="text-gray-900 font-medium">
            Create one
          </Link>
        </div>
      </div>
    </div>
  );
}
