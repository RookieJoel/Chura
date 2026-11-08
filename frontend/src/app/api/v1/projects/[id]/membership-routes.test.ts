import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  proxyBackendDelete,
  proxyBackendGet,
  proxyBackendPatch,
  proxyBackendPost,
  proxyBackendPut,
} from "@/lib/backend-proxy";
import {
  DELETE as deleteProject,
  GET as getProject,
  PATCH as patchProject,
} from "./route";
import { POST as addMember } from "./members/route";
import { PUT as assignRole } from "./members/[userId]/role/route";
import { DELETE as removeMember } from "./members/[userId]/route";
import { GET as listAuditors } from "./auditors/route";
import { POST as selfAssignAuditor } from "./auditors/me/route";
import {
  GET as listFeedback,
  POST as submitFeedback,
} from "./feedback/route";

vi.mock("@/lib/backend-proxy", () => ({
  proxyBackendDelete: vi.fn(),
  proxyBackendGet: vi.fn(),
  proxyBackendPatch: vi.fn(),
  proxyBackendPost: vi.fn(),
  proxyBackendPut: vi.fn(),
}));

describe("Project membership routes", () => {
  beforeEach(() => {
    vi.mocked(proxyBackendDelete).mockReset();
    vi.mocked(proxyBackendGet).mockReset();
    vi.mocked(proxyBackendPatch).mockReset();
    vi.mocked(proxyBackendPost).mockReset();
    vi.mocked(proxyBackendPut).mockReset();
  });

  it("forwards Project detail requests to the authenticated proxy", async () => {
    const request = new Request("http://localhost/api/v1/projects/p%2F1");
    const context = { params: Promise.resolve({ id: "p/1" }) };
    await getProject(request, context);

    expect(proxyBackendGet).toHaveBeenCalledWith(
      request,
      "/api/v1/projects/p%2F1",
    );
  });

  it("forwards member creation requests and body", async () => {
    const request = new Request("http://localhost/api/v1/projects/p1/members", {
      method: "POST",
      body: JSON.stringify({ email: "member@example.com" }),
    });
    const context = { params: Promise.resolve({ id: "p1" }) };
    await addMember(request, context);

    expect(proxyBackendPost).toHaveBeenCalledWith(
      request,
      "/api/v1/projects/p1/members",
    );
  });

  it("forwards role updates with encoded member IDs", async () => {
    const request = new Request(
      "http://localhost/api/v1/projects/p1/members/u%2F2/role",
      { method: "PUT", body: JSON.stringify({ role: "developer" }) },
    );
    const context = { params: Promise.resolve({ id: "p1", userId: "u/2" }) };
    await assignRole(request, context);

    expect(proxyBackendPut).toHaveBeenCalledWith(
      request,
      "/api/v1/projects/p1/members/u%2F2/role",
    );
  });

  it("forwards member removal requests with encoded IDs", async () => {
    const request = new Request(
      "http://localhost/api/v1/projects/p1/members/u%2F2",
      { method: "DELETE" },
    );
    const context = { params: Promise.resolve({ id: "p1", userId: "u/2" }) };
    await removeMember(request, context);

    expect(proxyBackendDelete).toHaveBeenCalledWith(
      request,
      "/api/v1/projects/p1/members/u%2F2",
    );
  });

  it("forwards project update, delete, auditor, and feedback routes", async () => {
    const context = { params: Promise.resolve({ id: "p/1" }) };
    const patchRequest = new Request("http://localhost/api/v1/projects/p%2F1", {
      method: "PATCH",
      body: JSON.stringify({ name: "Beta" }),
    });
    await patchProject(patchRequest, context);
    expect(proxyBackendPatch).toHaveBeenCalledWith(
      patchRequest,
      "/api/v1/projects/p%2F1",
    );

    const deleteRequest = new Request(
      "http://localhost/api/v1/projects/p%2F1",
      { method: "DELETE" },
    );
    await deleteProject(deleteRequest, context);
    expect(proxyBackendDelete).toHaveBeenCalledWith(
      deleteRequest,
      "/api/v1/projects/p%2F1",
    );

    const auditorsRequest = new Request(
      "http://localhost/api/v1/projects/p%2F1/auditors",
    );
    await listAuditors(auditorsRequest, context);
    expect(proxyBackendGet).toHaveBeenCalledWith(
      auditorsRequest,
      "/api/v1/projects/p%2F1/auditors",
    );

    const assignRequest = new Request(
      "http://localhost/api/v1/projects/p%2F1/auditors/me",
      { method: "POST" },
    );
    await selfAssignAuditor(assignRequest, context);
    expect(proxyBackendPost).toHaveBeenCalledWith(
      assignRequest,
      "/api/v1/projects/p%2F1/auditors/me",
    );

    const feedbackGet = new Request(
      "http://localhost/api/v1/projects/p%2F1/feedback",
    );
    await listFeedback(feedbackGet, context);
    expect(proxyBackendGet).toHaveBeenCalledWith(
      feedbackGet,
      "/api/v1/projects/p%2F1/feedback",
    );

    const feedbackPost = new Request(
      "http://localhost/api/v1/projects/p%2F1/feedback",
      { method: "POST", body: JSON.stringify({ body: "note" }) },
    );
    await submitFeedback(feedbackPost, context);
    expect(proxyBackendPost).toHaveBeenCalledWith(
      feedbackPost,
      "/api/v1/projects/p%2F1/feedback",
    );
  });
});
