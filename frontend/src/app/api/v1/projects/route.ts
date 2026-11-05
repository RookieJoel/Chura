import { proxyBackendGet, proxyBackendPost } from "@/lib/backend-proxy";

export function GET(request: Request) {
  const { search } = new URL(request.url);
  return proxyBackendGet(request, `/api/v1/projects${search}`);
}

export function POST(request: Request) {
  return proxyBackendPost(request, "/api/v1/projects");
}
