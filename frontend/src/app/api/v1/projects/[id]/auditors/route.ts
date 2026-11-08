import { proxyBackendGet, proxyBackendPost } from "@/lib/backend-proxy";

export async function GET(
  request: Request,
  context: { params: Promise<{ id: string }> },
) {
  const { id } = await context.params;
  return proxyBackendGet(
    request,
    `/api/v1/projects/${encodeURIComponent(id)}/auditors`,
  );
}

/** A Team Member assigns an Auditor to the Project (by picking them on a Work Item). */
export async function POST(
  request: Request,
  context: { params: Promise<{ id: string }> },
) {
  const { id } = await context.params;
  return proxyBackendPost(
    request,
    `/api/v1/projects/${encodeURIComponent(id)}/auditors`,
  );
}
