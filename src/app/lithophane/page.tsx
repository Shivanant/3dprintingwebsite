"use client";

import { ChangeEvent, useMemo, useState } from "react";

type Shape = "flat" | "arc" | "cylinder";

export default function LithophanePage() {
  const [preview, setPreview] = useState<string | null>(null);
  const [shape, setShape] = useState<Shape>("flat");
  const [width, setWidth] = useState(140);
  const [height, setHeight] = useState(100);
  const [thickness, setThickness] = useState(2);
  const [backlight, setBacklight] = useState(75);
  const [message, setMessage] = useState<string | null>(null);

  const estimatedPrice = useMemo(() => {
    const area = (width * height) / 100; // cm^2 proxy
    const complexityMultiplier = shape === "flat" ? 1 : shape === "arc" ? 1.15 : 1.25;
    const thicknessMultiplier = 1 + (thickness - 2) * 0.15;
    return Math.max(25, Math.round(area * 0.35 * complexityMultiplier * thicknessMultiplier));
  }, [width, height, shape, thickness]);

  function onUpload(e: ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    const url = URL.createObjectURL(file);
    setPreview(url);
  }

  return (
    <div className="relative overflow-hidden rounded-3xl border bg-gray-950 text-white shadow-xl">
      <div className="absolute inset-0 pointer-events-none bg-[radial-gradient(circle_at_top,_rgba(255,255,255,0.18),_transparent_60%)]" />

      <div className="relative grid gap-0 md:grid-cols-[1.25fr,1fr]">
        <section className="p-8 md:p-10 space-y-6">
          <div className="space-y-2">
            <span className="inline-flex items-center rounded-full border border-white/20 px-3 py-1 text-xs uppercase tracking-wide">
              Lithophane Studio
            </span>
            <h1 className="text-3xl md:text-4xl font-semibold tracking-tight">
              Turn memories into glowing lithophanes.
            </h1>
            <p className="text-sm md:text-base text-white/70">
              Upload any photo and watch it transform into a backlit sculpture. Our predictor adjusts pricing based on size, curvature, and material thickness.
            </p>
          </div>

          <label className="group relative block rounded-2xl border border-white/10 bg-white/5 hover:bg-white/10 transition">
            <input type="file" accept="image/*" className="hidden" onChange={onUpload} />
            <div className="p-6 text-center space-y-2">
              <div className="text-sm font-medium text-white">Upload photo</div>
              <div className="text-xs text-white/60">
                PNG, JPG up to 10MB • we auto-light balance for best lithophane contrast.
              </div>
            </div>
          </label>

          <div className="grid gap-4 md:grid-cols-2">
            <div className="rounded-2xl border border-white/10 bg-white/5 p-4 space-y-3">
              <Label title="Shape">
                <div className="flex gap-2">
                  {([
                    { key: "flat", label: "Flat" },
                    { key: "arc", label: "Arc" },
                    { key: "cylinder", label: "Cylinder" },
                  ] as { key: Shape; label: string }[]).map((option) => (
                    <button
                      key={option.key}
                      onClick={() => setShape(option.key)}
                      className={`flex-1 rounded-xl border px-3 py-2 text-sm transition ${
                        shape === option.key ? "border-white bg-white/10" : "border-white/10 text-white/60"
                      }`}
                    >
                      {option.label}
                    </button>
                  ))}
                </div>
              </Label>
              <Label title="Dimensions (mm)">
                <div className="grid grid-cols-2 gap-3">
                  <NumberField label="Width" value={width} setValue={setWidth} min={80} max={220} />
                  <NumberField label="Height" value={height} setValue={setHeight} min={60} max={200} />
                </div>
              </Label>
              <Label title="Thickness (mm)">
                <NumberField label="Shell" value={thickness} setValue={setThickness} min={1} max={4} step={0.5} />
              </Label>
            </div>

            <div className="rounded-2xl border border-white/10 bg-white/5 p-4 space-y-4">
              <Label title="Backlight intensity">
                <input
                  type="range"
                  min={40}
                  max={100}
                  value={backlight}
                  onChange={(e) => setBacklight(Number(e.target.value))}
                  className="w-full accent-white"
                />
                <div className="text-xs text-white/60">{backlight}% brightness preview</div>
              </Label>
              <div className="rounded-xl bg-white/10 p-4 space-y-1">
                <div className="text-xs uppercase tracking-wide text-white/50">Estimated price</div>
                <div className="text-3xl font-semibold">${estimatedPrice}</div>
                <div className="text-xs text-white/60">
                  Includes premium white PLA, fine detail pass, and tea light ready stand.
                </div>
              </div>
              <button
                onClick={() => setMessage("Lithophane builder coming soon. Add to cart via custom print for now!")}
                className="w-full rounded-full bg-white text-gray-900 py-3 text-sm font-medium hover:bg-white/90 transition"
              >
                Add to cart
              </button>
              {message && (
                <div className="rounded-xl border border-white/20 bg-white/5 px-4 py-3 text-xs text-white/80">
                  {message}
                </div>
              )}
            </div>
          </div>
        </section>

        <aside className="relative border-l border-white/10 bg-gradient-to-br from-white/10 to-white/0 p-8 flex flex-col gap-6 justify-center">
          <div className="relative h-80 md:h-[420px] overflow-hidden rounded-3xl border border-white/10 bg-black/40 backdrop-blur">
            {preview ? (
              <img
                src={preview}
                alt="Lithophane preview"
                className="h-full w-full object-cover mix-blend-screen opacity-90"
                style={{ filter: `brightness(${backlight}%) contrast(135%)` }}
              />
            ) : (
              <div className="h-full w-full flex items-center justify-center text-white/30 text-sm uppercase tracking-widest">
                Upload to preview
              </div>
            )}
            <GlowOverlay backlight={backlight} shape={shape} />
          </div>
          <div className="space-y-3 text-sm text-white/70">
            <p>
              Lithophanes are printed vertically to capture detail. We taper the rear shell to keep light diffusion smooth while
              preserving highlights. Expect delivery within 5–7 days.
            </p>
            <p className="text-xs text-white/50">
              Need a lamp base or multi-color panel? Mention it when you checkout and our team will reach out.
            </p>
          </div>
        </aside>
      </div>
    </div>
  );
}

function Label({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="space-y-2">
      <div className="text-xs uppercase tracking-wide text-white/50">{title}</div>
      {children}
    </div>
  );
}

function NumberField({
  label,
  value,
  setValue,
  min,
  max,
  step = 1,
}: {
  label: string;
  value: number;
  setValue: (num: number) => void;
  min: number;
  max: number;
  step?: number;
}) {
  return (
    <label className="space-y-1 text-xs text-white/60">
      {label}
      <input
        type="number"
        min={min}
        max={max}
        step={step}
        value={value}
        onChange={(e) => setValue(Number(e.target.value))}
        className="w-full rounded-xl border border-white/20 bg-black/30 px-3 py-2 text-sm text-white focus:outline-none focus:ring-2 focus:ring-white/40"
      />
    </label>
  );
}

function GlowOverlay({ backlight, shape }: { backlight: number; shape: Shape }) {
  const gradient =
    shape === "flat"
      ? "radial-gradient(circle at center, rgba(255,255,255,0.4), transparent 55%)"
      : shape === "arc"
      ? "radial-gradient(circle at 30% 50%, rgba(255,255,255,0.45), transparent 60%)"
      : "radial-gradient(circle at 15% 50%, rgba(255,255,255,0.5), transparent 70%)";

  return (
    <div
      className="absolute inset-0 pointer-events-none transition"
      style={{
        backgroundImage: gradient,
        opacity: Math.min(1, backlight / 80),
      }}
    />
  );
}
