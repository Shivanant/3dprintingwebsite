"use client";

import Link from "next/link";
import { useEffect } from "react";
import { useAuth } from "@/store/auth";
import { useRouter } from "next/navigation";

export default function AccountPage() {
  const { user, initialized, bootstrap } = useAuth();
  const router = useRouter();

  useEffect(() => {
    void bootstrap();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (initialized && !user) {
      router.push("/login?redirect=/account");
    }
  }, [initialized, user, router]);

  if (!initialized) {
    return <div className="text-sm text-gray-500">Loading account…</div>;
  }

  if (!user) {
    return (
      <div className="max-w-md mx-auto text-center space-y-4">
        <h1 className="text-2xl font-semibold">Account required</h1>
        <p className="text-gray-600">
          Sign in to view saved carts, print jobs, and order history.
        </p>
        <div className="flex justify-center gap-3">
          <Link href="/login" className="px-4 py-2 rounded-full border hover:bg-gray-50">
            Sign in
          </Link>
          <Link href="/register" className="px-4 py-2 rounded-full bg-black text-white">
            Create account
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="grid gap-6 md:grid-cols-[2fr,1fr]">
      <section className="rounded-2xl bg-white border p-6 space-y-4">
        <div>
          <h1 className="text-2xl font-semibold">Account overview</h1>
          <p className="text-sm text-gray-500">Manage your preferences and order history.</p>
        </div>
        <div className="grid gap-4 text-sm">
          <div className="rounded-xl border px-4 py-3">
            <div className="text-gray-500">Name</div>
            <div className="text-gray-900 font-medium">{user.name || "Add your name"}</div>
          </div>
          <div className="rounded-xl border px-4 py-3">
            <div className="text-gray-500">Email</div>
            <div className="text-gray-900 font-medium">{user.email}</div>
          </div>
          <div className="rounded-xl border px-4 py-3">
            <div className="text-gray-500">Role</div>
            <div className="text-gray-900 font-medium capitalize">{user.role}</div>
          </div>
        </div>

        <Link
          href="/orders"
          className="inline-flex items-center gap-2 text-sm font-medium text-gray-900 hover:underline"
        >
          View order history →
        </Link>
      </section>

      <aside className="space-y-4">
        <div className="rounded-2xl bg-white border p-5 space-y-2">
          <h2 className="font-semibold">Quick links</h2>
          <div className="flex flex-col text-sm gap-1 text-gray-600">
            <Link href="/cart" className="hover:text-gray-900">
              Saved cart
            </Link>
            <Link href="/custom-print" className="hover:text-gray-900">
              Upload a new model
            </Link>
            <Link href="/help" className="hover:text-gray-900">
              Need support?
            </Link>
          </div>
        </div>
      </aside>
    </div>
  );
}
