import { proxyBackendGet } from "@/lib/backend-proxy";

export function GET(request: Request) {
  return proxyBackendGet(request, "/api/v1/templates");
}
