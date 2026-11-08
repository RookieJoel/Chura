import { proxyBackendGet } from "@/lib/backend-proxy";

/** Every Auditor in the system, for picking a Work Item's Auditor. */
export async function GET(request: Request) {
  return proxyBackendGet(request, "/api/v1/auditors");
}
