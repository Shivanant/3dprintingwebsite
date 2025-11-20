"use client";

import { useEffect, useState } from "react";
import { useAuth } from "@/store/auth";
import api from "@/lib/apiClient";
import { useRouter } from "next/navigation";

type Order = {
  id: string;
  status: string;
  totalCents: number;
  currency: string;
  createdAt: string;
  items: {
    id: string;
    name: string;
    quantity: number;
    unitPriceCents: number;
  }[];
};

export default function OrdersPage() {
  const { user, initialized, bootstrap } = useAuth();
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const router = useRouter();

  useEffect(() => {
    void bootstrap();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (!initialized) return;
    if (!user) {
      router.push("/login?redirect=/orders");
      return;
    }
    async function loadOrders() {
      try {
        const { data } = await api.get<Order[]>("/orders");
        setOrders(data);
      } catch (err) {
        console.error(err);
      } finally {
        setLoading(false);
      }
    }
    void loadOrders();
  }, [initialized, user, router]);

  if (!initialized || loading) {
    return <div className="text-sm text-gray-500">Loading orders…</div>;
  }

  if (orders.length === 0) {
    return (
      <div className="max-w-lg bg-white border rounded-2xl p-8 space-y-4">
        <h1 className="text-2xl font-semibold">Your orders</h1>
        <p className="text-gray-600">
          No orders yet. Upload a model or explore the shop to get your first print going!
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold">Order history</h1>
      <div className="grid gap-4">
        {orders.map((order) => (
          <div key={order.id} className="rounded-2xl border bg-white p-6 space-y-4">
            <div className="flex flex-wrap justify-between gap-2 text-sm text-gray-500">
              <span>
                Order #{order.id.split("-")[0]} • {new Date(order.createdAt).toLocaleString()}
              </span>
              <span className="font-medium text-gray-900 capitalize">{order.status}</span>
            </div>
            <div className="space-y-2 text-sm">
              {order.items.map((item) => (
                <div key={item.id} className="flex justify-between">
                  <span>
                    {item.quantity}× {item.name}
                  </span>
                  <span>${(item.unitPriceCents * item.quantity / 100).toFixed(2)}</span>
                </div>
              ))}
            </div>
            <div className="flex justify-between text-sm font-semibold text-gray-900">
              <span>Total</span>
              <span>${(order.totalCents / 100).toFixed(2)}</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
