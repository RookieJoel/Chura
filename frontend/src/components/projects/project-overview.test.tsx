import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { getProject, listProjects } from "@/api/projects";
import { ProjectProvider } from "@/components/layout/project-provider";
import { ProjectOverview } from "@/components/projects/project-overview";

vi.mock("@/api/projects", () => ({
  getProject: vi.fn(),
  listProjects: vi.fn(),
}));

const project = {
  id: "p1",
  name: "Alpha",
  description: "A project description",
  template_id: "se",
  mode: "se",
};

describe("ProjectOverview", () => {
  beforeEach(() => {
    localStorage.clear();
    vi.mocked(listProjects).mockResolvedValue([project]);
    vi.mocked(getProject).mockResolvedValue(project);
  });

  it("loads the requested Project details without Sprint or Work Item content", async () => {
    render(
      <ProjectProvider>
        <ProjectOverview projectId="p1" />
      </ProjectProvider>,
    );

    expect(await screen.findByRole("heading", { name: "Alpha" })).toBeInTheDocument();
    expect(screen.getByText("A project description")).toBeInTheDocument();
    expect(screen.queryByText(/Sprint|Work Item/)).not.toBeInTheDocument();
  });

  it("shows a clear access error when the requested Project cannot be loaded", async () => {
    vi.mocked(getProject).mockRejectedValue(new Error("not a project member"));
    render(
      <ProjectProvider>
        <ProjectOverview projectId="missing" />
      </ProjectProvider>,
    );

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "not a project member",
    );
  });
});
