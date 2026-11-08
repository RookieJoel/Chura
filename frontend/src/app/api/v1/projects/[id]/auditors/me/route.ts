import { proxyBackendPost } from "@/lib/backend-proxy";

export async function POST(
  request: Request,
  context: { params: Promise<{ id: string }> },
) {
  const { id } = await context.params;
  return proxyBackendPost(
    request,
    `/api/v1/projects/${encodeURIComponent(id)}/auditors/me`,
  );
}
