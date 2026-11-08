import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createComment, listComments } from "@/api/work-items";
import { WorkItemComments } from "@/components/work-items/work-item-comments";

vi.mock("@/api/work-items", () => ({
  listComments: vi.fn(),
  createComment: vi.fn(),
}));

describe("WorkItemComments", () => {
  beforeEach(() => {
    vi.mocked(listComments).mockReset();
    vi.mocked(createComment).mockReset();
  });

  it("shows author, timestamp, and empty/loading/error states", async () => {
    vi.mocked(listComments).mockResolvedValueOnce([]);
    const { rerender } = render(<WorkItemComments workItemId="wi-1" />);
    expect(screen.getByText("Loading comments…")).toBeInTheDocument();
    expect(await screen.findByText("No comments yet.")).toBeInTheDocument();

    vi.mocked(listComments).mockRejectedValueOnce(new Error("forbidden (403)"));
    rerender(<WorkItemComments workItemId="wi-2" />);
    expect(await screen.findByRole("alert")).toHaveTextContent("forbidden (403)");

    vi.mocked(listComments).mockResolvedValueOnce([
      {
        id: "c1",
        workItemId: "wi-3",
        authorUserId: "u1",
        authorName: "Ada",
        authorEmail: "ada@example.com",
        body: "Looks good",
        createdAt: "2026-03-01T12:00:00Z",
      },
    ]);
    rerender(<WorkItemComments workItemId="wi-3" />);

    expect(await screen.findByText("Looks good")).toBeInTheDocument();
    expect(screen.getByText("Ada")).toBeInTheDocument();
    expect(screen.getByText("Newest first")).toBeInTheDocument();
  });

  it("posts a comment and prepends it newest-first", async () => {
    vi.mocked(listComments).mockResolvedValue([]);
    vi.mocked(createComment).mockResolvedValue({
      id: "c2",
      workItemId: "wi-1",
      authorUserId: "u2",
      authorName: "Grace",
      authorEmail: "grace@example.com",
      body: "Ship it",
      createdAt: "2026-03-02T09:00:00Z",
    });

    render(<WorkItemComments workItemId="wi-1" />);
    await screen.findByText("No comments yet.");

    fireEvent.change(screen.getByLabelText("Add comment"), {
      target: { value: "Ship it" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Post comment" }));

    await waitFor(() =>
      expect(createComment).toHaveBeenCalledWith("wi-1", "Ship it"),
    );
    expect(await screen.findByText("Ship it")).toBeInTheDocument();
    expect(screen.getByText("Grace")).toBeInTheDocument();
    expect(screen.getByLabelText("Add comment")).toHaveValue("");
  });

  it("keeps unauthorized post errors visible", async () => {
    vi.mocked(listComments).mockResolvedValue([]);
    vi.mocked(createComment).mockRejectedValue(new Error("forbidden (403)"));

    render(<WorkItemComments workItemId="wi-1" />);
    await screen.findByText("No comments yet.");

    fireEvent.change(screen.getByLabelText("Add comment"), {
      target: { value: "Nope" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Post comment" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      /forbidden \(403\).*self-assign in Settings/i,
    );
    expect(screen.getByLabelText("Add comment")).toHaveValue("Nope");
  });
});
