const DEFAULT_API_BASE = "http://localhost:8080/api/v1";

const apiBaseUrl =
  process.env.API_BASE_URL ||
  process.env.NEXT_PUBLIC_API_URL ||
  DEFAULT_API_BASE;

type EstimateOptions = {
  material?: string;
  quality?: string;
};

export async function requestPricingEstimate(file: File, opts: EstimateOptions = {}) {
  const form = new FormData();
  form.append("file", file, file.name);
  if (opts.material) {
    form.append("material", opts.material);
  }
  if (opts.quality) {
    form.append("quality", opts.quality);
  }

  const res = await fetch(`${apiBaseUrl}/pricing/estimate`, {
    method: "POST",
    body: form,
  });

  const body = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(
      body?.error ??
        `Pricing service returned ${res.status} ${res.statusText}`
    );
  }

  return body;
}
