"use client";

import Link from "next/link";
import { useSearchParams, useRouter } from "next/navigation";
import { FormEvent, useState } from "react";
import { useAuth } from "@/store/auth";

export default function ResetPasswordForm() {
  const params = useSearchParams();
  const router = useRouter();
  const token = params.get("token") ?? "";
  const { resetPassword } = useAuth();
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    if (!token) {
      setError("Reset token missing. Use the link from your email.");
      return;
    }
    if (password !== confirm) {
      setError("Passwords do not match.");
      return;
    }
    try {
      await resetPassword(token, password);
      setSuccess(true);
      setTimeout(() => router.push("/account"), 1200);
    } catch (err: any) {
      setError(err?.response?.data?.error || "Unable to reset password. The link may have expired.");
    }
  }

  return (
    <div className="max-w-md mx-auto bg-white rounded-2xl border shadow-sm p-8 space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Set a new password</h1>
        <p className="text-sm text-gray-500">Enter a secure password to complete the reset.</p>
      </div>

      {error && (
        <div className="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
          {error}
        </div>
      )}

      {success ? (
        <div className="rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-700">
          Password updated! Redirecting you to your account...
        </div>
      ) : (
        <form className="space-y-4" onSubmit={onSubmit}>
          <label className="block text-sm">
            New password
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              minLength={8}
              required
              className="mt-1 w-full rounded-lg border px-3 py-2 focus:outline-none focus:ring-2 focus:ring-gray-900/10"
            />
          </label>

          <label className="block text-sm">
            Confirm password
            <input
              type="password"
              value={confirm}
              onChange={(e) => setConfirm(e.target.value)}
              minLength={8}
              required
              className="mt-1 w-full rounded-lg border px-3 py-2 focus:outline-none focus:ring-2 focus:ring-gray-900/10"
            />
          </label>

          <button
            type="submit"
            className="w-full rounded-lg bg-black text-white py-2.5 font-medium hover:bg-gray-900 transition"
          >
            Update password
          </button>
        </form>
      )}

      <Link href="/login" className="text-sm text-gray-500 hover:text-gray-800 transition">
        Back to sign in
      </Link>
    </div>
  );
}
