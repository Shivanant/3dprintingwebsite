import { NextRequest, NextResponse } from "next/server";
import { requestPricingEstimate } from "@/server/services/pricingService";

export async function handlePricingEstimate(req: NextRequest) {
  const form = await req.formData();
  const maybeFile = form.get("file");

  if (!(maybeFile instanceof File)) {
    return NextResponse.json(
      { error: "file required" },
      { status: 400 }
    );
  }

  try {
    const material = form.get("material")?.toString();
    const quality = form.get("quality")?.toString();
    const result = await requestPricingEstimate(maybeFile, {
      material: material || undefined,
      quality: quality || undefined,
    });
    return NextResponse.json(result, { status: 200 });
  } catch (error: any) {
    const message =
      typeof error?.message === "string"
        ? error.message
        : "Unable to reach pricing service";
    return NextResponse.json(
      { error: message },
      { status: 502 }
    );
  }
}
