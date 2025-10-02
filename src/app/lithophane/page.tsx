"use client";
import { useState } from "react";

export default function LithophanePage() {
  const [img, setImg] = useState<string | null>(null);

  return (
    <div className="grid md:grid-cols-2 gap-6">
      <div className="space-y-4">
        <h1 className="text-2xl font-bold">Create Lithophane</h1>
        <label className="block rounded-xl border bg-white p-6 text-center cursor-pointer">
          <input
            type="file"
            accept="image/*"
            className="hidden"
            onChange={(e) => {
              const f = e.target.files?.[0];
              if (f) setImg(URL.createObjectURL(f));
            }}
          />
          <span className="text-gray-600">Upload a photo</span>
        </label>

        <div className="rounded-xl border bg-white p-4">
          <div className="grid grid-cols-2 gap-4 text-sm">
            <label>
              Shape
              <select className="border rounded p-1 w-full">
                <option>Flat</option>
                <option>Curve</option>
                <option>Cylinder</option>
              </select>
            </label>
            <label>
              Size (mm)
              <input className="border rounded p-1 w-full" defaultValue="120x80" />
            </label>
            <label>
              Thickness
              <input className="border rounded p-1 w-full" defaultValue="0.8–3.0" />
            </label>
            <label>
              Material
              <select className="border rounded p-1 w-full">
                <option>PLA White</option>
              </select>
            </label>
          </div>
          <div className="mt-3 font-bold">Est. Price: ₹ 399</div>
        </div>

        <button className="px-4 py-2 rounded bg-black text-white">Add to Cart</button>
      </div>

      <div className="h-80 rounded-xl border bg-white flex items-center justify-center overflow-hidden">
        {img ? (
          <img src={img} alt="preview" className="object-contain h-full" />
        ) : (
          <span className="text-gray-400">Lithophane Preview</span>
        )}
      </div>
    </div>
  );
}
