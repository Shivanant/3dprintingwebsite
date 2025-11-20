"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import api from "@/lib/apiClient";
import { useAuth } from "@/store/auth";
import { useRouter } from "next/navigation";

type CartItem = {
  id: string;
  sku: string;
  displayName: string;
  quantity: number;
  unitPriceCents: number;
  metadata: Record<string, any>;
};

type CartResponse = {
  id: string;
  items: CartItem[];
  subtotal: number;
};

export default function CartPage() {
  const { user, initialized, bootstrap } = useAuth();
  const router = useRouter();
  const [cart, setCart] = useState<CartResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [note, setNote] = useState("");

  useEffect(() => {
    void bootstrap();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (!initialized) return;
    if (!user) {
      router.push("/login?redirect=/cart");
      return;
    }
    async function fetchCart() {
      try {
        const { data } = await api.get<CartResponse>("/cart");
        setCart(data);
      } catch (err) {
        setError("Unable to load your cart. Please try again.");
      } finally {
        setLoading(false);
      }
    }
    void fetchCart();
  }, [initialized, user, router]);

  async function removeItem(id: string) {
    try {
      const { data } = await api.delete<CartResponse>(`/cart/items/${id}`);
      setCart(data);
    } catch (err) {
      setError("Failed to remove item.");
    }
  }

  async function checkout() {
    try {
      await api.post("/orders/checkout", { notes: note });
      router.push("/orders");
    } catch (err) {
      setError("Checkout failed. Please try again.");
    }
  }

  if (!initialized || loading) {
    return <div className="text-sm text-gray-500">Loading cart…</div>;
  }

  if (!user) {
    return null;
  }

  if (!cart || cart.items.length === 0) {
    return (
      <div className="max-w-lg mx-auto bg-white border rounded-2xl p-8 space-y-4 text-center">
        <h1 className="text-2xl font-semibold">Your cart is empty</h1>
        <p className="text-gray-500">Upload a model to see instant pricing and add it to your cart.</p>
        <div className="flex justify-center gap-3">
          <Link href="/custom-print" className="px-4 py-2 rounded-full bg-black text-white hover:bg-gray-900">
            Start a custom print
          </Link>
          <Link href="/shop" className="px-4 py-2 rounded-full border hover:bg-gray-50">
            Browse shop
          </Link>
        </div>
      </div>
    );
  }

  const subtotal = cart.items.reduce((sum, item) => sum + item.unitPriceCents * item.quantity, 0);
  const tax = Math.round(subtotal * 0.08);
  const total = subtotal + tax;

  return (
    <div className="grid gap-6 md:grid-cols-[2fr,1fr]">
      <section className="space-y-4">
        <h1 className="text-2xl font-semibold">Cart</h1>

        {error && (
          <div className="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
            {error}
          </div>
        )}

        <div className="space-y-4">
          {cart.items.map((item) => (
            <div key={item.id} className="rounded-2xl border bg-white p-5 flex flex-col gap-2">
              <div className="flex justify-between items-center">
                <div>
                  <div className="font-medium text-gray-900">{item.displayName}</div>
                  <div className="text-sm text-gray-500">
                    Quantity: {item.quantity} • ${(item.unitPriceCents / 100).toFixed(2)} ea
                  </div>
                </div>
                <button
                  onClick={() => removeItem(item.id)}
                  className="text-sm text-gray-500 hover:text-gray-900 transition"
                >
                  Remove
                </button>
              </div>
              {item.metadata?.estimatedGrams && (
                <div className="text-xs text-gray-500">
                  ~{Math.round(item.metadata.estimatedGrams)} g &bull; ~{item.metadata.estimatedHours} h print &bull;{" "}
                  {item.metadata.material ?? "PLA"}
                </div>
              )}
            </div>
          ))}
        </div>

        <label className="block text-sm">
          Order notes (optional)
          <textarea
            value={note}
            onChange={(e) => setNote(e.target.value)}
            className="mt-1 w-full rounded-xl border px-3 py-2 focus:outline-none focus:ring-2 focus:ring-gray-900/10 resize-none"
            rows={3}
            placeholder="Share surface finish preferences, timelines, or other instructions."
          />
        </label>
      </section>

      <aside className="rounded-2xl border bg-white p-6 space-y-4">
        <h2 className="text-lg font-semibold">Summary</h2>
        <div className="space-y-2 text-sm">
          <div className="flex justify-between">
            <span>Subtotal</span>
            <span>${(subtotal / 100).toFixed(2)}</span>
          </div>
          <div className="flex justify-between">
            <span>Tax (8%)</span>
            <span>${(tax / 100).toFixed(2)}</span>
          </div>
          <div className="flex justify-between font-semibold text-gray-900">
            <span>Total</span>
            <span>${(total / 100).toFixed(2)}</span>
          </div>
        </div>

        <button
          onClick={checkout}
          className="w-full rounded-full bg-black text-white py-2.5 font-medium hover:bg-gray-900 transition"
        >
          Place order
        </button>
      </aside>
    </div>
  );
}
