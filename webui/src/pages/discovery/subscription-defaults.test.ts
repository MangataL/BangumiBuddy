import { describe, expect, it } from "vitest";

import type { ReleaseGroupCandidate } from "@/api/discovery";

import { getDefaultSubscriptionPriority } from "./subscription-defaults";

function group(previousMaxPriority?: number): ReleaseGroupCandidate {
  return {
    mikanSubgroupID: "10",
    name: "测试字幕组",
    subscribed: false,
    previousMaxPriority,
  };
}

describe("discovery subscription defaults", () => {
  it("使用该字幕组历史订阅中的最大优先级", () => {
    expect(getDefaultSubscriptionPriority(group(9))).toBe(9);
  });

  it("没有历史优先级时回退为 1", () => {
    expect(getDefaultSubscriptionPriority(group())).toBe(1);
    expect(getDefaultSubscriptionPriority(group(0))).toBe(1);
  });
});
