import {NextRequest, NextResponse} from "next/server";
import { quickEstimate } from "@/lib/pricing";

export const runtime = "nodejs";

export async function POST(req: NextRequest) {
  const form = await req.formData();
  const file = form.get("file") as File | null;
  if (!file) return NextResponse.json({error:"file required"}, {status:400});
  const buf = Buffer.from(await file.arrayBuffer());
  const est = quickEstimate(buf.byteLength);
  return NextResponse.json(est);
}
