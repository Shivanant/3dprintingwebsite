"use client";

import Link from "next/link";
import { FormEvent, useState } from "react";
import { useAuth } from "@/store/auth";

export default function ForgotPasswordPage() {
  const { forgotPassword } = useAuth();
  const [email, setEmail] = useState("");
  const [submitted, setSubmitted] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    try {
      await forgotPassword(email);
      setSubmitted(true);
    } catch (err: any) {
      setError(err?.response?.data?.error || "Something went wrong. Try again.");
    }
  }

  return (
    <div className="max-w-md mx-auto bg-white rounded-2xl border shadow-sm p-8 space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Reset your password</h1>
        <p className="text-sm text-gray-500">
          Enter the email linked to your account and we&apos;ll send you a reset link powered by Mailgun.
        </p>
      </div>

      {submitted ? (
        <div className="space-y-3 text-sm text-gray-600">
          <div className="rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-2 text-emerald-700">
            We&apos;ve sent password reset instructions to {email}. Check your inbox and spam folder.
          </div>
          <Link href="/login" className="text-gray-900 font-medium">
            Go back to sign in
          </Link>
        </div>
      ) : (
        <>
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
            <button
              type="submit"
              className="w-full rounded-lg bg-black text-white py-2.5 font-medium hover:bg-gray-900 transition"
            >
              Send reset link
            </button>
          </form>
        </>
      )}
    </div>
  );
}
