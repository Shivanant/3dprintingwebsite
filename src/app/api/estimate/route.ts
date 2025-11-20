import { NextRequest } from "next/server";
import { handlePricingEstimate } from "@/server/handlers/pricingHandler";

export const runtime = "nodejs";

export async function POST(req: NextRequest) {
  return handlePricingEstimate(req);
}
