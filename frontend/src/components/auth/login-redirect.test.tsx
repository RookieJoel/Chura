import { render, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { signIn } from "next-auth/react";
import { LoginRedirect } from "@/components/auth/login-redirect";

vi.mock("next-auth/react", () => ({ signIn: vi.fn() }));

describe("LoginRedirect", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("passes the selected destination to Keycloak sign-in", async () => {
    render(
      <LoginRedirect callbackUrl="/projects/p1?tab=members#invite" />,
    );

    await waitFor(() => {
      expect(signIn).toHaveBeenCalledWith("keycloak", {
        callbackUrl: "/projects/p1?tab=members#invite",
      });
    });
  });
});
