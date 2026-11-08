import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { getCurrentIdentity } from "@/api/auth";
import { getProject, listProjects } from "@/api/projects";
import { getTemplate } from "@/api/templates";
import { ProjectProvider } from "@/components/layout/project-provider";
import { Sidebar } from "@/components/layout/sidebar";

const { pathname } = vi.hoisted(() => ({ pathname: vi.fn() }));

vi.mock("next/navigation", () => ({ usePathname: pathname }));
vi.mock("@/api/auth", () => ({ getCurrentIdentity: vi.fn() }));
vi.mock("@/api/projects", () => ({
  getProject: vi.fn(),
  listProjects: vi.fn(),
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

describe("Sidebar", () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
    pathname.mockReturnValue("/projects/p1/feedback");
    vi.mocked(listProjects).mockResolvedValue([project]);
    vi.mocked(getProject).mockResolvedValue(project);
    vi.mocked(getTemplate).mockResolvedValue(template);
    vi.mocked(getCurrentIdentity).mockResolvedValue({
      authenticated: true,
      user_id: "u1",
      email: "owner@example.com",
      roles: ["member"],
    });
  });

  it("links to selected Project feedback and marks it active beside settings", async () => {
    render(
      <ProjectProvider initialProjectId="p1">
        <Sidebar />
      </ProjectProvider>,
    );

    const feedbackLink = await screen.findByRole("link", {
      name: "Project feedback",
    });
    const settingsLink = screen.getByRole("link", { name: "Project settings" });

    expect(feedbackLink).toHaveAttribute("href", "/projects/p1/feedback");
    expect(feedbackLink.className).toContain("bg-brand-tint");
    expect(settingsLink).toHaveAttribute("href", "/projects/p1");
    expect(settingsLink.className).not.toContain("bg-brand-tint");
  });
});
