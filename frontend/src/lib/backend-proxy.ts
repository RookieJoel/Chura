import { getToken } from "next-auth/jwt";
import { NextResponse } from "next/server";
import { sessionCookieName } from "../../auth";

async function proxyBackendRequest(
  request: Request,
  path: string,
  method: "GET" | "POST",
) {
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
    const headers = new Headers({
      Authorization: `Bearer ${accessToken}`,
    });
    const body =
      method === "POST" ? await request.arrayBuffer() : undefined;
    if (method === "POST") {
      headers.set(
        "content-type",
        request.headers.get("content-type") ?? "application/json",
      );
    }
    const response = await fetch(`${apiUrl.replace(/\/$/, "")}${path}`, {
      method,
      headers,
      body: body && body.byteLength > 0 ? body : undefined,
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

export function proxyBackendGet(request: Request, path: string) {
  return proxyBackendRequest(request, path, "GET");
}

export function proxyBackendPost(request: Request, path: string) {
  return proxyBackendRequest(request, path, "POST");
}
