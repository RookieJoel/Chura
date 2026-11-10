import { beforeEach, describe, expect, it, vi } from "vitest";
import { redirect } from "next/navigation";
import Home from "@/app/page";

vi.mock("next/navigation", () => ({ redirect: vi.fn() }));

describe("Home", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("redirects to the Board", () => {
    Home();

    expect(redirect).toHaveBeenCalledWith("/board");
  });
});
