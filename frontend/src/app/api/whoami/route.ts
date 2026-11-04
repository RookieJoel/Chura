import { NextResponse } from "next/server";
import { auth } from "../../../../auth";

export async function GET() {
  const session = await auth();
  const accessToken = session?.accessToken;

  if (!accessToken) {
    return NextResponse.json({ error: "Unauthenticated" }, { status: 401 });
  }

  const apiUrl = process.env.CHURA_API_URL;
  if (!apiUrl) {
    return NextResponse.json(
      { error: "CHURA_API_URL is not configured" },
      { status: 500 },
    );
  }

  const response = await fetch(
    `${apiUrl.replace(/\/$/, "")}/api/v1/whoami`,
    {
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
      cache: "no-store",
    },
  );

  const body = await response.text();
  return new NextResponse(body, {
    status: response.status,
    headers: {
      "content-type": response.headers.get("content-type") ?? "application/json",
    },
  });
}
