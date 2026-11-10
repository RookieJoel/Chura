import { proxyBackendPut } from "@/lib/backend-proxy";

export async function PUT(
  request: Request,
  context: { params: Promise<{ id: string; userId: string }> },
) {
  const { id, userId } = await context.params;
  return proxyBackendPut(
    request,
    `/api/v1/projects/${encodeURIComponent(id)}/members/${encodeURIComponent(userId)}/role`,
  );
}
