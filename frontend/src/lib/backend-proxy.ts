import { getToken } from "next-auth/jwt";
import { NextResponse } from "next/server";
import { sessionCookieName } from "../../auth";

export async function proxyBackendGet(request: Request, path: string) {
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

  try {
    const response = await fetch(`${apiUrl.replace(/\/$/, "")}${path}`, {
      headers: { Authorization: `Bearer ${accessToken}` },
      cache: "no-store",
    });
    return new NextResponse(await response.text(), {
      status: response.status,
      headers: {
        "content-type":
          response.headers.get("content-type") ?? "application/json",
      },
    });
  } catch {
    return NextResponse.json(
      { error: "Backend API unavailable" },
      { status: 502 },
    );
  }
}
