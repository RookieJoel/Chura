import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import LoginPage from "@/app/login/page";

vi.mock("@/components/auth/login-redirect", () => ({
  LoginRedirect: ({ callbackUrl }: { callbackUrl: string }) => (
    <div data-testid="callback-url">{callbackUrl}</div>
  ),
}));

async function renderLoginPage(callbackUrl?: string | string[]) {
  const page = await LoginPage({
    params: Promise.resolve({}),
    searchParams: Promise.resolve(
      callbackUrl === undefined ? {} : { callbackUrl },
    ),
  });
  return render(page);
}

describe("LoginPage", () => {
  it("defaults a missing callback URL to the Board", async () => {
    await renderLoginPage();

    expect(screen.getByTestId("callback-url")).toHaveTextContent("/board");
  });

  it("defaults malformed callback URLs to the Board", async () => {
    await renderLoginPage("http://[");

    expect(screen.getByTestId("callback-url")).toHaveTextContent("/board");
  });

  it("defaults rejected cross-origin callback URLs to the Board", async () => {
    await renderLoginPage("https://example.com/projects");

    expect(screen.getByTestId("callback-url")).toHaveTextContent("/board");
  });

  it("defaults repeated callback parameters to the Board", async () => {
    await renderLoginPage(["/projects", "/backlog"]);

    expect(screen.getByTestId("callback-url")).toHaveTextContent("/board");
  });

  it("preserves safe same-origin callback path, query, and fragment", async () => {
    await renderLoginPage("/projects/p1?tab=members#invite");

    expect(screen.getByTestId("callback-url")).toHaveTextContent(
      "/projects/p1?tab=members#invite",
    );
  });

  it("preserves accepted absolute same-origin callbacks", async () => {
    await renderLoginPage("http://localhost/projects/p1?tab=members#invite");

    expect(screen.getByTestId("callback-url")).toHaveTextContent(
      "/projects/p1?tab=members#invite",
    );
  });
});
