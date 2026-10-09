import { getToken } from "next-auth/jwt";
import { NextResponse } from "next/server";
import { sessionCookieName } from "../../../../auth";

export async function GET(request: Request) {
  const token = await getToken({
    req: request,
    secret: process.env.AUTH_SECRET,
    cookieName: sessionCookieName,
    secureCookie: process.env.NODE_ENV === "production",
  });
  const accessToken =
    typeof token?.accessToken === "string" ? token.accessToken : undefined;

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
