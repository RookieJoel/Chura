import { describe, expect, it } from "vitest";
import {
  DEFAULT_REFLECTION_TOPICS,
  formatMetricValue,
  reflectionTopicsFromTemplate,
} from "@/lib/reflection-utils";

describe("reflection-utils", () => {
  it("falls back to the standard retrospective topics", () => {
    expect(reflectionTopicsFromTemplate(undefined)).toEqual(
      DEFAULT_REFLECTION_TOPICS,
    );
    expect(reflectionTopicsFromTemplate([])).toEqual(DEFAULT_REFLECTION_TOPICS);
  });

  it("keeps template topics when present", () => {
    const topics = [
      { key: "went_well", question: "Custom went well?", required: true },
    ];
    expect(reflectionTopicsFromTemplate(topics)).toEqual(topics);
  });

  it("formats percent metrics with a percent sign", () => {
    expect(formatMetricValue(25, "percent")).toBe("25%");
    expect(formatMetricValue(33.3, "percent")).toBe("33.3%");
    expect(formatMetricValue(4, "items")).toBe("4");
  });
});
