import { proxyBackendGet, proxyBackendPost } from "@/lib/backend-proxy";

export async function GET(
  request: Request,
  context: { params: Promise<{ id: string }> },
) {
  const { id } = await context.params;
  return proxyBackendGet(
    request,
    `/api/v1/work-items/${encodeURIComponent(id)}/comments`,
  );
}

export async function POST(
  request: Request,
  context: { params: Promise<{ id: string }> },
) {
  const { id } = await context.params;
  return proxyBackendPost(
    request,
    `/api/v1/work-items/${encodeURIComponent(id)}/comments`,
  );
}
