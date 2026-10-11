import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { ProfileMenu } from "@/components/layout/profile-menu";

describe("ProfileMenu", () => {
  it("shows the backend-verified identity and logs out", async () => {
    global.fetch = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          authenticated: true,
          user_id: "user-1",
          email: "user@example.com",
          roles: ["member", "auditor"],
        }),
      ),
    );
    render(<ProfileMenu />);
    await screen.findByRole("button", { name: /user@example.com/i });
    fireEvent.click(screen.getByRole("button", { name: /user@example.com/i }));

    expect(screen.getByText("ID: user-1")).toBeInTheDocument();
    expect(screen.getByText("member")).toBeInTheDocument();
    expect(screen.getByText("auditor")).toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: "Log out" })).toBeInTheDocument();
  });
});
