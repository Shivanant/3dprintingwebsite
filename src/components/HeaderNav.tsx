"use client";

import Link from "next/link";
import { useEffect } from "react";
import { useAuth } from "@/store/auth";

const navLinks = [
  { href: "/shop", label: "Shop" },
  { href: "/custom-print", label: "Custom Print" },
  { href: "/lithophane", label: "Lithophane" },
  { href: "/help", label: "Help" },
];

export default function HeaderNav() {
  const { user, logout, initialized, bootstrap } = useAuth();

  useEffect(() => {
    void bootstrap();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <header className="border-b bg-white sticky top-0 z-40">
      <nav className="container mx-auto flex items-center gap-6 p-4">
        <Link href="/" className="font-semibold text-lg tracking-tight">
          3DPrint Hub
        </Link>

        <div className="hidden md:flex items-center gap-4 text-sm text-gray-600">
          {navLinks.map((item) => (
            <Link
              key={item.href}
              href={item.href}
              className="hover:text-gray-900 transition-colors"
            >
              {item.label}
            </Link>
          ))}
          {user?.role === "admin" && (
            <Link href="/admin" className="font-semibold text-amber-600">
              Admin
            </Link>
          )}
        </div>

        <div className="ml-auto flex items-center gap-3 text-sm">
          <Link
            href="/cart"
            className="rounded-full border px-3 py-1.5 hover:bg-gray-50 transition-colors"
          >
            Cart {user ? "" : "🔒"}
          </Link>

          {initialized ? (
            user ? (
              <div className="flex items-center gap-2">
                <Link
                  href="/account"
                  className="hidden md:flex flex-col leading-tight text-right"
                >
                  <span className="text-xs text-gray-500">Signed in as</span>
                  <span className="font-medium text-gray-800">{user.name || user.email}</span>
                </Link>
                <button
                  onClick={logout}
                  className="rounded-full border px-3 py-1.5 hover:bg-gray-50 transition-colors"
                >
                  Log out
                </button>
              </div>
            ) : (
              <div className="flex items-center gap-2">
                <Link
                  href="/login"
                  className="px-3 py-1.5 rounded-full border hover:bg-gray-50 transition-colors"
                >
                  Sign in
                </Link>
                <Link
                  href="/register"
                  className="px-3 py-1.5 rounded-full bg-black text-white hover:bg-gray-800 transition-colors"
                >
                  Create account
                </Link>
              </div>
            )
          ) : (
            <span className="text-xs text-gray-500">Checking session…</span>
          )}
        </div>
      </nav>
    </header>
  );
}
