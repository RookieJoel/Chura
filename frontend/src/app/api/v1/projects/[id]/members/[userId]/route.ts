import { proxyBackendDelete } from "@/lib/backend-proxy";

export async function DELETE(
  request: Request,
  context: { params: Promise<{ id: string; userId: string }> },
) {
  const { id, userId } = await context.params;
  return proxyBackendDelete(
    request,
    `/api/v1/projects/${encodeURIComponent(id)}/members/${encodeURIComponent(userId)}`,
  );
}
