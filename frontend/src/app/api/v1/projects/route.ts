import { proxyBackendGet } from "@/lib/backend-proxy";

export function GET(request: Request) {
  const { search } = new URL(request.url);
  return proxyBackendGet(request, `/api/v1/projects${search}`);
}
