import { NextResponse } from "next/server";
import { auth } from "../../../../auth";

export async function GET() {
  const session = await auth();
  if (!session?.accessToken) {
    return NextResponse.json({ error: "Unauthenticated" }, { status: 401 });
  }
  return NextResponse.json({ access_token: session.accessToken });
}
