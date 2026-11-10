import {
  proxyBackendDelete,
  proxyBackendGet,
  proxyBackendPatch,
} from "@/lib/backend-proxy";

export async function GET(
  request: Request,
  context: { params: Promise<{ id: string }> },
) {
  const { id } = await context.params;
  return proxyBackendGet(
    request,
    `/api/v1/projects/${encodeURIComponent(id)}`,
  );
}

export async function PATCH(
  request: Request,
  context: { params: Promise<{ id: string }> },
) {
  const { id } = await context.params;
  return proxyBackendPatch(
    request,
    `/api/v1/projects/${encodeURIComponent(id)}`,
  );
}

export async function DELETE(
  request: Request,
  context: { params: Promise<{ id: string }> },
) {
  const { id } = await context.params;
  return proxyBackendDelete(
    request,
    `/api/v1/projects/${encodeURIComponent(id)}`,
  );
}
