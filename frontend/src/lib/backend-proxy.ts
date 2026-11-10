import { getToken } from "next-auth/jwt";
import { NextResponse } from "next/server";
import { sessionCookieName } from "../../auth";

const NULL_BODY_STATUSES = new Set([204, 205, 304]);

async function proxyBackendRequest(
  request: Request,
  path: string,
  method: "GET" | "POST" | "PUT" | "PATCH" | "DELETE",
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
      method === "GET" ? undefined : await request.arrayBuffer();
    if (method !== "GET") {
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
    // Null-body statuses (204 No Content, 205, 304) must not carry a body:
    // the Response constructor throws on one, even an empty string.
    if (NULL_BODY_STATUSES.has(response.status)) {
      return new NextResponse(null, { status: response.status });
    }
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

export function proxyBackendPut(request: Request, path: string) {
  return proxyBackendRequest(request, path, "PUT");
}

export function proxyBackendPatch(request: Request, path: string) {
  return proxyBackendRequest(request, path, "PATCH");
}

export function proxyBackendDelete(request: Request, path: string) {
  return proxyBackendRequest(request, path, "DELETE");
}
