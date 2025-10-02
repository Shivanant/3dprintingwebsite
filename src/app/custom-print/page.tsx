"use client";
import { useEffect, useState } from "react";
import FileDropzone from "@/components/FileDropzone";
import ThreeViewer from "@/components/ThreeViewer";
import axios from "axios";

export default function CustomPrintPage() {
  const [file, setFile] = useState<File | null>(null);
  const [fileUrl, setFileUrl] = useState<string | undefined>(undefined);
  const [estimate, setEstimate] = useState<{ grams: number; timeHours: number; total: number } | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!file) { setFileUrl(undefined); return; }
    const url = URL.createObjectURL(file);
    setFileUrl(url);
    return () => URL.revokeObjectURL(url);
  }, [file]);

  async function handleFiles(files: File[]) {
    const f = files[0];
    setFile(f);
    setLoading(true);
    const form = new FormData();
    form.append("file", f);
    const { data } = await axios.post("/api/estimate", form);
    setEstimate(data);
    setLoading(false);
  }

  return (
    <div className="grid md:grid-cols-2 gap-6">
      <div className="space-y-4">
        <h1 className="text-2xl font-bold">Upload & Estimate</h1>
        <FileDropzone onFiles={handleFiles} />
        {file && <div className="text-sm text-gray-600">Selected: {file.name}</div>}
        {estimate && (
          <div className="rounded-xl border bg-white p-4">
            <div className="font-semibold mb-2">Estimated Cost</div>
            <div className="text-sm">Material: ~{estimate.grams} g</div>
            <div className="text-sm">Print time: ~{estimate.timeHours} h</div>
            <div className="text-lg font-bold mt-2">₹ {estimate.total}</div>
          </div>
        )}
        <button
          disabled={!estimate || loading}
          className="px-4 py-2 rounded bg-black text-white disabled:opacity-50"
        >
          Add to Cart
        </button>
      </div>
      <div>
        <ThreeViewer url={fileUrl} />
      </div>
    </div>
  );
}
