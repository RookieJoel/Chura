import { NextResponse } from "next/server";
import { auth } from "../../../../../auth";

async function proxy(request: Request, context: RouteContext<"/api/backend/[...path]">) {
  const session = await auth();
  const accessToken = session?.accessToken;
  if (!accessToken) {
    return NextResponse.json({ error: "Unauthenticated" }, { status: 401 });
  }

  const { path } = await context.params;
  const apiUrl = process.env.CHURA_API_URL;
  if (!apiUrl) {
    return NextResponse.json(
      { error: "CHURA_API_URL is not configured" },
      { status: 500 },
    );
  }

  const target = `${apiUrl.replace(/\/$/, "")}/api/v1/${path.join("/")}${new URL(request.url).search}`;
  const body =
    request.method === "GET" || request.method === "HEAD"
      ? undefined
      : await request.arrayBuffer();
  const response = await fetch(target, {
    method: request.method,
    headers: {
      Authorization: `Bearer ${accessToken}`,
      "content-type": request.headers.get("content-type") ?? "application/json",
    },
    body: body && body.byteLength > 0 ? body : undefined,
    cache: "no-store",
  });

  const headers = new Headers({
    "content-type":
      response.headers.get("content-type") ?? "application/json",
  });
  const contentDisposition = response.headers.get("content-disposition");
  if (contentDisposition) {
    headers.set("content-disposition", contentDisposition);
  }
  const contentTypeOptions = response.headers.get("x-content-type-options");
  if (contentTypeOptions) {
    headers.set("x-content-type-options", contentTypeOptions);
  }

  return new NextResponse(response.body, {
    status: response.status,
    headers,
  });
}

export const GET = proxy;
export const POST = proxy;
export const PUT = proxy;
export const DELETE = proxy;
