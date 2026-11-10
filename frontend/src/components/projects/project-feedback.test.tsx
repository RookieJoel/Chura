import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { getCurrentIdentity } from "@/api/auth";
import {
  listProjectAuditors,
  listProjectFeedback,
  listProjects,
  submitProjectFeedback,
} from "@/api/projects";
import { getTemplate } from "@/api/templates";
import { ProjectProvider } from "@/components/layout/project-provider";
import { ProjectFeedbackPage } from "@/components/projects/project-feedback";

vi.mock("@/api/auth", () => ({ getCurrentIdentity: vi.fn() }));
vi.mock("@/api/projects", () => ({
  listProjectAuditors: vi.fn(),
  listProjectFeedback: vi.fn(),
  listProjects: vi.fn(),
  getProject: vi.fn(),
  submitProjectFeedback: vi.fn(),
}));
vi.mock("@/api/templates", () => ({ getTemplate: vi.fn() }));

const project = {
  id: "p1",
  name: "Alpha",
  description: "A project description",
  template_id: "se",
  mode: "se",
  created_by: "u1",
  members: [],
};
const template = {
  id: "se",
  name: "Software Engineering",
  mode: "se",
  work_item_types: ["user_story", "task", "bug"],
  roles: ["product_owner", "scrum_master", "developer"],
  default_role: "developer",
  creator_role: "product_owner",
};
const assignment = {
  project_id: "p1",
  auditor_user_id: "a1",
  auditor_name: "Auditor",
  auditor_email: "auditor@example.com",
  assigned_at: "2026-01-02T03:04:05Z",
};

function renderFeedbackPage() {
  return render(
    <ProjectProvider initialProjectId="p1">
      <ProjectFeedbackPage projectId="p1" />
    </ProjectProvider>,
  );
}

describe("ProjectFeedbackPage", () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
    vi.mocked(listProjects).mockResolvedValue([project]);
    vi.mocked(getTemplate).mockResolvedValue(template);
    vi.mocked(listProjectAuditors).mockResolvedValue([]);
    vi.mocked(listProjectFeedback).mockResolvedValue([]);
    vi.mocked(submitProjectFeedback).mockResolvedValue({
      id: "f1",
      project_id: "p1",
      author_user_id: "a1",
      author_name: "Auditor",
      author_email: "auditor@example.com",
      body: "Looks good",
      created_at: "2026-01-02T03:05:05Z",
    });
    vi.mocked(getCurrentIdentity).mockResolvedValue({
      authenticated: true,
      user_id: "u1",
      email: "member@example.com",
      roles: ["member"],
    });
  });

  it("shows an explicit loading state while feedback is fetched", async () => {
    vi.mocked(listProjectFeedback).mockReturnValue(new Promise(() => {}));

    renderFeedbackPage();

    expect(await screen.findByRole("status")).toHaveTextContent(
      "Loading feedback…",
    );
  });

  it("shows an explicit error when feedback cannot be loaded", async () => {
    vi.mocked(listProjectFeedback).mockRejectedValue(
      new Error("Unable to load feedback"),
    );

    renderFeedbackPage();

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "Unable to load feedback",
    );
  });

  it("shows an empty state when there is no feedback", async () => {
    renderFeedbackPage();

    expect(await screen.findByText("No Project feedback yet.")).toBeInTheDocument();
  });

  it("lets Project members read history without seeing the submission form", async () => {
    vi.mocked(listProjectFeedback).mockResolvedValue([
      {
        id: "f1",
        project_id: "p1",
        author_user_id: "a1",
        author_name: "Auditor",
        author_email: "auditor@example.com",
        body: "Keep going",
        created_at: "2026-01-02T03:05:05Z",
      },
    ]);

    renderFeedbackPage();

    expect(await screen.findByText("Keep going")).toBeInTheDocument();
    expect(screen.queryByLabelText("Add feedback")).not.toBeInTheDocument();
  });

  it("allows only an assigned Auditor to submit and adds the result to history", async () => {
    vi.mocked(getCurrentIdentity).mockResolvedValue({
      authenticated: true,
      user_id: "a1",
      email: "auditor@example.com",
      roles: ["auditor"],
    });
    vi.mocked(listProjectAuditors).mockResolvedValue([assignment]);

    renderFeedbackPage();

    fireEvent.change(await screen.findByLabelText("Add feedback"), {
      target: { value: "Looks good" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Submit feedback" }));

    await waitFor(() =>
      expect(submitProjectFeedback).toHaveBeenCalledWith("p1", "Looks good"),
    );
    expect(await screen.findByText("Looks good")).toBeInTheDocument();
  });

  it("does not show the form to an Auditor who is not assigned", async () => {
    vi.mocked(getCurrentIdentity).mockResolvedValue({
      authenticated: true,
      user_id: "a1",
      email: "auditor@example.com",
      roles: ["auditor"],
    });

    renderFeedbackPage();

    await waitFor(() => expect(listProjectAuditors).toHaveBeenCalledWith("p1"));
    expect(screen.queryByLabelText("Add feedback")).not.toBeInTheDocument();
  });

  it("surfaces submission errors and preserves the entered feedback", async () => {
    vi.mocked(getCurrentIdentity).mockResolvedValue({
      authenticated: true,
      user_id: "a1",
      email: "auditor@example.com",
      roles: ["auditor"],
    });
    vi.mocked(listProjectAuditors).mockResolvedValue([assignment]);
    vi.mocked(submitProjectFeedback).mockRejectedValue(
      new Error("Unable to submit feedback"),
    );

    renderFeedbackPage();

    const textarea = await screen.findByLabelText("Add feedback");
    fireEvent.change(textarea, { target: { value: "Please review" } });
    fireEvent.click(screen.getByRole("button", { name: "Submit feedback" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "Unable to submit feedback",
    );
    expect(textarea).toHaveValue("Please review");
  });
});
