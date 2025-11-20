"use client";

import { useEffect, useMemo, useState } from "react";
import FileDropzone from "@/components/FileDropzone";
import ThreeViewer from "@/components/ThreeViewer";
import api from "@/lib/apiClient";
import { useAuth } from "@/store/auth";
import { useRouter } from "next/navigation";

type Estimate = {
  id: string;
  material: string;
  estimatedGrams: number;
  estimatedHours: number;
  estimatedPrice: number;
  boundingBoxMm: { min: [number, number, number]; max: [number, number, number] };
  triangleCount: number;
  recommendedInfill: number;
  warnings?: string[];
};

export default function CustomPrintPage() {
  const [file, setFile] = useState<File | null>(null);
  const [fileUrl, setFileUrl] = useState<string>();
  const [estimate, setEstimate] = useState<Estimate | null>(null);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const { user, initialized, bootstrap } = useAuth();
  const router = useRouter();

  useEffect(() => {
    void bootstrap();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (!file) {
      setFileUrl(undefined);
      return;
    }
    const url = URL.createObjectURL(file);
    setFileUrl(url);
    return () => URL.revokeObjectURL(url);
  }, [file]);

  async function handleFiles(files: File[]) {
    const picked = files[0];
    setError(null);
    setMessage(null);
    setFile(picked);
    setLoading(true);
    try {
      const form = new FormData();
      form.append("file", picked);
      form.append("material", "PLA");
      form.append("quality", "standard");
      const { data } = await api.post<Estimate>("/pricing/estimate", form, {
        headers: { "Content-Type": "multipart/form-data" },
      });
      setEstimate(data);
    } catch (err: any) {
      setError(err?.response?.data?.error || "Could not evaluate this model. Try another file.");
    } finally {
      setLoading(false);
    }
  }

  async function addToCart() {
    if (!estimate || !file) return;
    if (initialized && !user) {
      router.push("/login?redirect=/custom-print");
      return;
    }
    try {
      const displayName = file.name.replace(/\.[^.]+$/, "");
      await api.post("/cart/items", {
        sku: estimate.id,
        displayName,
        quantity: 1,
        unitPriceCents: Math.round(estimate.estimatedPrice * 100),
        metadata: {
          estimatedGrams: estimate.estimatedGrams,
          estimatedHours: estimate.estimatedHours,
          material: estimate.material,
          triangleCount: estimate.triangleCount,
          boundingBox: estimate.boundingBoxMm,
        },
      });
      setMessage("Added to your cart! You can review it from the header.");
    } catch (err: any) {
      setError(err?.response?.data?.error || "Failed to add to cart. Try again.");
    }
  }

  const boundingBox = useMemo(() => {
    if (!estimate) return null;
    const [minX, minY, minZ] = estimate.boundingBoxMm.min;
    const [maxX, maxY, maxZ] = estimate.boundingBoxMm.max;
    return {
      x: Math.abs(maxX - minX).toFixed(1),
      y: Math.abs(maxY - minY).toFixed(1),
      z: Math.abs(maxZ - minZ).toFixed(1),
    };
  }, [estimate]);

  return (
    <div className="grid gap-6 xl:grid-cols-[1.1fr,0.9fr]">
      <div className="space-y-4">
        <div className="rounded-3xl border bg-white p-6 shadow-sm space-y-3">
          <h1 className="text-3xl font-semibold tracking-tight">Upload a model</h1>
          <p className="text-sm text-gray-600">
            Drop STL or OBJ files to get material usage, print time, and pricing with our Go-powered estimator.
          </p>
          <FileDropzone onFiles={handleFiles} />
          {file && (
            <div className="text-xs text-gray-500">
              Selected: <span className="font-medium text-gray-900">{file.name}</span>
            </div>
          )}
          {error && (
            <div className="rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{error}</div>
          )}
          {message && (
            <div className="rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">
              {message}{" "}
              <button onClick={() => router.push("/cart")} className="underline">
                Go to cart
              </button>
            </div>
          )}
        </div>

        {estimate && (
          <div className="rounded-3xl border bg-white p-6 shadow-sm space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold">Estimated cost</h2>
              <span className="text-2xl font-bold">${estimate.estimatedPrice.toFixed(2)}</span>
            </div>
            <div className="grid gap-3 md:grid-cols-3 text-sm">
              <Metric label="Material" value={`${Math.round(estimate.estimatedGrams)} g PLA`} />
              <Metric label="Print time" value={`${estimate.estimatedHours.toFixed(1)} h`} />
              <Metric label="Infill" value={`${estimate.recommendedInfill}%`} />
            </div>
            {boundingBox && (
              <div className="rounded-2xl bg-slate-50 px-4 py-3 text-xs text-gray-600">
                Bounding box: {boundingBox.x} x {boundingBox.y} x {boundingBox.z} mm • Triangles:{" "}
                {estimate.triangleCount.toLocaleString()}
              </div>
            )}
            {estimate.warnings?.length ? (
              <ul className="text-xs text-amber-600 border border-amber-200 rounded-xl bg-amber-50 px-3 py-2 space-y-1">
                {estimate.warnings.map((warn) => (
                  <li key={warn}>! {warn}</li>
                ))}
              </ul>
            ) : null}
            <button
              onClick={addToCart}
              disabled={loading}
              className="w-full rounded-full bg-black text-white py-3 font-medium hover:bg-gray-900 transition disabled:opacity-60"
            >
              {user ? "Add to cart" : "Sign in to add to cart"}
            </button>
          </div>
        )}
      </div>

      <div className="space-y-4">
        <ThreeViewer url={fileUrl} />
        <div className="rounded-3xl border bg-white p-6 shadow-sm space-y-3 text-sm text-gray-600">
          <h2 className="text-lg font-semibold text-gray-900">Tips for best results</h2>
          <ul className="space-y-2">
            <li>- Orient models so the largest face meets the build plate for faster print times.</li>
            <li>- We double-check wall thickness and supports before the print starts.</li>
            <li>- Want premium finishes or multi-material prints? Mention it in your order notes.</li>
          </ul>
        </div>
      </div>
    </div>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl border px-4 py-3 bg-slate-50">
      <div className="text-xs uppercase tracking-wide text-gray-500">{label}</div>
      <div className="text-sm font-semibold text-gray-900">{value}</div>
    </div>
  );
}
