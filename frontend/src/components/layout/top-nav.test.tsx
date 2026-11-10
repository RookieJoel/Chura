import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { getProject, listProjects } from "@/api/projects";
import { getCurrentIdentity } from "@/api/auth";
import { getTemplate } from "@/api/templates";
import { AppShell } from "@/components/layout/app-shell";
import { ProjectProvider } from "@/components/layout/project-provider";
import { TopNav } from "@/components/layout/top-nav";

const { pathname } = vi.hoisted(() => ({ pathname: { current: "/backlog" } }));

vi.mock("@/api/projects", () => ({
  getProject: vi.fn(),
  listProjects: vi.fn(),
}));
vi.mock("@/api/auth", () => ({ getCurrentIdentity: vi.fn() }));
vi.mock("@/api/templates", () => ({ getTemplate: vi.fn() }));
vi.mock("next/navigation", () => ({
  usePathname: () => pathname.current,
  useRouter: (() => {
    const router = { push: vi.fn(), refresh: vi.fn(), replace: vi.fn() };
    return () => router;
  })(),
}));

const projects = [
  { id: "p1", name: "Alpha", description: "", template_id: "se", mode: "se" },
  { id: "p2", name: "Beta", description: "", template_id: "se", mode: "se" },
];
const projectDetails = {
  ...projects[0],
  created_by: "u1",
  members: [
    {
      id: "u1",
      email: "owner@example.com",
      name: "Project Owner",
      role: "product_owner",
      initials: "PO",
      color: "#64748B",
    },
  ],
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

function renderTopNav() {
  return render(
    <ProjectProvider>
      <TopNav />
    </ProjectProvider>,
  );
}

describe("TopNav project selector", () => {
  beforeEach(() => {
    pathname.current = "/backlog";
    localStorage.clear();
    vi.mocked(listProjects).mockReset();
    vi.mocked(listProjects).mockResolvedValue(projects);
    vi.mocked(getProject).mockResolvedValue(projectDetails);
    vi.mocked(getTemplate).mockResolvedValue(template);
    vi.mocked(getCurrentIdentity).mockResolvedValue({
      authenticated: true,
      user_id: "u1",
      email: "user@example.com",
      roles: ["member"],
    });
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            authenticated: true,
            user_id: "u1",
            email: "user@example.com",
            roles: ["member"],
          }),
        ),
      ),
    );
  });

  it("loads accessible Projects, replaces an inaccessible stored ID, and persists selection", async () => {
    localStorage.setItem("chura.selectedProjectId", "stale");
    vi.mocked(listProjects).mockResolvedValue(projects);
    renderTopNav();

    const selector = await screen.findByRole("combobox", { name: "Project" });
    await waitFor(() => expect(selector).toHaveValue("p1"));
    expect(localStorage.getItem("chura.selectedProjectId")).toBe("p1");

    fireEvent.change(selector, { target: { value: "p2" } });
    expect(selector).toHaveValue("p2");
    expect(localStorage.getItem("chura.selectedProjectId")).toBe("p2");
  });

  it("replaces the shared member list when switching Projects", async () => {
    vi.mocked(listProjects).mockResolvedValue(projects);
    vi.mocked(getProject).mockImplementation(async (projectId) => ({
      ...projectDetails,
      id: projectId,
      name: projectId === "p1" ? "Alpha" : "Beta",
      members: [
        {
          id: projectId === "p1" ? "u1" : "u2",
          email: projectId === "p1" ? "alpha@example.com" : "beta@example.com",
          name: projectId === "p1" ? "Alpha Member" : "Beta Member",
          role: "",
          initials: "AM",
          color: "#64748B",
        },
      ],
    }));

    render(
      <ProjectProvider>
        <AppShell>
          <main>Page content</main>
        </AppShell>
      </ProjectProvider>,
    );
    const selector = await screen.findByRole("combobox", { name: "Project" });
    expect(await screen.findAllByTitle("Alpha Member")).toHaveLength(2);

    fireEvent.change(selector, { target: { value: "p2" } });

    expect(await screen.findAllByTitle("Beta Member")).toHaveLength(2);
    expect(screen.queryByTitle("Alpha Member")).not.toBeInTheDocument();
  });

  it("uses the project provider in the actual application shell", async () => {
    vi.mocked(listProjects).mockResolvedValue(projects);
    render(
      <ProjectProvider>
        <AppShell>
          <main>Page content</main>
        </AppShell>
      </ProjectProvider>,
    );

    expect(await screen.findByRole("combobox", { name: "Project" })).toHaveValue("p1");
    expect(await screen.findAllByText("Alpha")).not.toHaveLength(0);
    expect(await screen.findAllByTitle("Project Owner")).toHaveLength(2);
    expect(screen.getByText("Page content")).toBeInTheDocument();
  });

  it("shows loading and empty Project states", async () => {
    vi.mocked(listProjects).mockResolvedValue([]);
    renderTopNav();
    expect(screen.getByRole("status", { name: "Loading Projects" })).toBeInTheDocument();
    expect(await screen.findByText("No accessible Projects")).toBeInTheDocument();
    expect(screen.getByRole("combobox", { name: "Project" })).toBeDisabled();
    expect(screen.getByRole("link", { name: "Create Project" })).toHaveAttribute(
      "href",
      "/projects/new",
    );
    expect(localStorage.getItem("chura.selectedProjectId")).toBeNull();
  });

  it("shows an explicit empty member state for a Project with no members", async () => {
    vi.mocked(getProject).mockResolvedValue({
      ...projectDetails,
      members: [],
    });
    renderTopNav();

    expect(await screen.findByText("No Project members")).toBeInTheDocument();
  });

  it("shows loading and error feedback while member data is unavailable", async () => {
    vi.mocked(getProject).mockReturnValue(new Promise(() => {}));
    renderTopNav();
    expect(await screen.findByText("Loading Project members…")).toBeInTheDocument();
  });

  it("shows member-load errors in the shared shell", async () => {
    vi.mocked(getProject).mockRejectedValue(
      new Error("Project members unavailable"),
    );
    renderTopNav();

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "Project members unavailable",
    );
  });

  it("keeps the Project overview shell free of Sprint and Work Item controls", async () => {
    pathname.current = "/projects/p1";
    vi.mocked(listProjects).mockResolvedValue(projects);
    render(
      <ProjectProvider>
        <AppShell minimal>
          <main>Project overview content</main>
        </AppShell>
      </ProjectProvider>,
    );

    expect(await screen.findByText("Project overview content")).toBeInTheDocument();
    const selector = screen.getByRole("combobox", { name: "Project" });
    expect(selector).toBeDisabled();
    expect(screen.queryByText("Search work items")).not.toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Sprints" })).not.toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Board" })).not.toBeInTheDocument();
  });

  it("locks Project switching while viewing Project settings", async () => {
    pathname.current = "/projects/p1";
    renderTopNav();
    const selector = await screen.findByRole("combobox", { name: "Project" });
    await waitFor(() => expect(selector).toHaveValue("p1"));
    expect(selector).toBeDisabled();
  });

  it("shows discovery errors and allows retry", async () => {
    vi.mocked(listProjects)
      .mockRejectedValueOnce(new Error("service unavailable"))
      .mockResolvedValueOnce(projects);
    renderTopNav();
    expect(await screen.findByRole("alert")).toHaveTextContent("service unavailable");
    fireEvent.click(screen.getByRole("button", { name: "Retry" }));
    await waitFor(() =>
      expect(screen.getByRole("combobox", { name: "Project" })).toHaveValue("p1"),
    );
    expect(listProjects).toHaveBeenCalledTimes(2);
  });
});
