"use client";

import { useEffect, useState } from "react";
import api from "@/lib/apiClient";
import { useAuth } from "@/store/auth";
import { useRouter } from "next/navigation";

type AdminOrder = {
  id: string;
  status: string;
  totalCents: number;
  subtotalCents: number;
  taxCents: number;
  createdAt: string;
  userId: string;
};

const STATUSES = ["pending", "paid", "in_progress", "shipped", "cancelled"] as const;

export default function AdminPage() {
  const { user, initialized, bootstrap } = useAuth();
  const router = useRouter();
  const [orders, setOrders] = useState<AdminOrder[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    void bootstrap();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (!initialized) return;
    if (!user || user.role !== "admin") {
      router.push("/account");
      return;
    }
    async function load() {
      try {
        const { data } = await api.get<AdminOrder[]>("/admin/orders");
        setOrders(data);
      } catch (err: any) {
        setError(err?.response?.data?.error || "Failed to load admin data.");
      } finally {
        setLoading(false);
      }
    }
    void load();
  }, [initialized, user, router]);

  async function updateStatus(orderId: string, status: string) {
    try {
      await api.patch(`/admin/orders/${orderId}/status`, { status });
      setOrders((prev) =>
        prev.map((order) => (order.id === orderId ? { ...order, status } : order))
      );
    } catch (err) {
      setError("Failed to update order status.");
    }
  }

  if (!initialized || loading) {
    return <div className="text-sm text-gray-500">Loading admin dashboard…</div>;
  }

  if (error) {
    return (
      <div className="rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
        {error}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Admin control center</h1>
        <p className="text-sm text-gray-500">Review orders, update statuses, and keep the print farm flowing.</p>
      </div>
      <div className="rounded-2xl border bg-white p-6 space-y-4">
        <div className="flex justify-between text-sm text-gray-500">
          <span>{orders.length} orders in system</span>
          <span>Statuses update instantly</span>
        </div>

        <div className="overflow-x-auto">
          <table className="min-w-full text-sm">
            <thead className="text-left text-gray-500 uppercase tracking-wide text-xs">
              <tr className="border-b">
                <th className="py-2 pr-6">Order</th>
                <th className="py-2 pr-6">User</th>
                <th className="py-2 pr-6">Placed</th>
                <th className="py-2 pr-6">Total</th>
                <th className="py-2 pr-6">Status</th>
                <th className="py-2">Action</th>
              </tr>
            </thead>
            <tbody className="align-top">
              {orders.map((order) => (
                <tr key={order.id} className="border-b last:border-0">
                  <td className="py-3 pr-6 font-medium text-gray-900">{order.id.slice(0, 8)}</td>
                  <td className="py-3 pr-6 text-gray-500 text-xs">{order.userId}</td>
                  <td className="py-3 pr-6">{new Date(order.createdAt).toLocaleString()}</td>
                  <td className="py-3 pr-6 font-semibold">${(order.totalCents / 100).toFixed(2)}</td>
                  <td className="py-3 pr-6 capitalize">{order.status.replaceAll("_", " ")}</td>
                  <td className="py-3">
                    <select
                      value={order.status}
                      onChange={(e) => updateStatus(order.id, e.target.value)}
                      className="rounded-lg border px-2 py-1 text-xs"
                    >
                      {STATUSES.map((status) => (
                        <option key={status} value={status}>
                          {status.replaceAll("_", " ")}
                        </option>
                      ))}
                    </select>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
