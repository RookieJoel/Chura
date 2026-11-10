import { proxyBackendGet } from "@/lib/backend-proxy";

export async function GET(
  request: Request,
  context: { params: Promise<{ id: string }> },
) {
  const { id } = await context.params;
  return proxyBackendGet(
    request,
    `/api/v1/templates/${encodeURIComponent(id)}`,
  );
}
