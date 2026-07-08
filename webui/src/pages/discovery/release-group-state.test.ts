import { describe, expect, it } from "vitest";

import type {
  BangumiCandidate,
  ReleaseGroupCandidate,
  ReleaseGroupSummary,
} from "@/api/discovery";

import {
  getSubscribedReleaseGroupCount,
  hasLoadedNoReleaseGroups,
  orderReleaseGroups,
} from "./release-group-state";

function createCandidate(id: string): BangumiCandidate {
  return {
    mikanBangumiID: id,
    name: `番剧 ${id}`,
    posterURL: "",
    weekday: 1,
    airStartDate: "",
  };
}

function createGroup(
  id: string,
  subscribed = false
): ReleaseGroupCandidate {
  return {
    mikanSubgroupID: id,
    name: `字幕组 ${id}`,
    subscribed,
  };
}

describe("release group state", () => {
  it("已订阅字幕组置顶且各分组保持原顺序", () => {
    const groups = [
      createGroup("a"),
      createGroup("subscribed-a", true),
      createGroup("b"),
      createGroup("subscribed-b", true),
    ];

    expect(
      orderReleaseGroups(groups).map((item) => item.mikanSubgroupID)
    ).toEqual(["subscribed-a", "subscribed-b", "a", "b"]);
    expect(groups.map((item) => item.mikanSubgroupID)).toEqual([
      "a",
      "subscribed-a",
      "b",
      "subscribed-b",
    ]);
  });

  it("从 summary 中派生已订阅字幕组数量", () => {
    const summary: ReleaseGroupSummary = {
      mikanBangumiID: "1",
      releaseGroups: [
        { mikanSubgroupID: "10", name: "组一", subscribed: true },
        { mikanSubgroupID: "20", name: "组二", subscribed: false },
      ],
    };

    expect(getSubscribedReleaseGroupCount(summary)).toBe(1);
    expect(getSubscribedReleaseGroupCount(undefined)).toBe(0);
  });

  it("服务端预判或已加载空 summary 时进入灰化状态", () => {
    const candidate = createCandidate("1");
    expect(hasLoadedNoReleaseGroups(candidate, undefined)).toBe(false);
    expect(
      hasLoadedNoReleaseGroups(
        { ...candidate, releaseGroupsKnownEmpty: true },
        undefined
      )
    ).toBe(true);
    expect(
      hasLoadedNoReleaseGroups(candidate, {
        mikanBangumiID: "1",
        releaseGroups: [],
      })
    ).toBe(true);
    expect(
      hasLoadedNoReleaseGroups(candidate, {
        mikanBangumiID: "1",
        releaseGroups: [
          { mikanSubgroupID: "10", name: "组一", subscribed: false },
        ],
      })
    ).toBe(false);
  });
});
