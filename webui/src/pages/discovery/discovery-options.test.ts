import { describe, expect, it } from "vitest";

import type {
  BangumiCandidate,
  ReleaseGroupSummary,
} from "@/api/discovery";

import {
  filterScheduleBangumis,
  getCurrentSeason,
  groupBangumisByWeekday,
  weekdayOptions,
} from "./discovery-options";

function candidate(
  id: string,
  weekday: number,
  releaseGroupsKnownEmpty = false
): BangumiCandidate {
  return {
    mikanBangumiID: id,
    name: `番剧 ${id}`,
    posterURL: "",
    weekday,
    airStartDate: "",
    releaseGroupsKnownEmpty,
  };
}

function summary(
  id: string,
  releaseGroupCount: number,
  subscribedReleaseGroupCount = 0
): ReleaseGroupSummary {
  return {
    mikanBangumiID: id,
    releaseGroups: Array.from({ length: releaseGroupCount }, (_, index) => ({
      mikanSubgroupID: `${id}-${index}`,
      name: `字幕组 ${index}`,
      subscribed: index < subscribedReleaseGroupCount,
    })),
  };
}

describe("discovery view model", () => {
  it("按月份确定季度", () => {
    expect(getCurrentSeason(new Date("2026-01-15T12:00:00Z"))).toBe("winter");
    expect(getCurrentSeason(new Date("2026-04-15T12:00:00Z"))).toBe("spring");
    expect(getCurrentSeason(new Date("2026-07-15T12:00:00Z"))).toBe("summer");
    expect(getCurrentSeason(new Date("2026-10-15T12:00:00Z"))).toBe("fall");
  });

  it("按放送日分组并把剧场版保留为独立入口", () => {
    const groups = groupBangumisByWeekday([
      candidate("1", 5),
      candidate("2", 5),
      candidate("3", 0),
      candidate("movie", 7),
    ]);

    expect(groups[5].map((item) => item.mikanBangumiID)).toEqual(["1", "2"]);
    expect(groups[0].map((item) => item.mikanBangumiID)).toEqual(["3"]);
    expect(groups[7].map((item) => item.mikanBangumiID)).toEqual(["movie"]);
    expect(weekdayOptions.some((item) => item.value === 7)).toBe(false);
  });

  it("按可订阅和已订阅状态筛选", () => {
    const items = [
      candidate("none", 1, true),
      candidate("available", 1),
      candidate("subscribed", 1),
    ];
    const summaries = {
      available: summary("available", 3),
      subscribed: summary("subscribed", 4, 1),
    };

    expect(
      filterScheduleBangumis(items, summaries, "available").map(
        (item) => item.mikanBangumiID
      )
    ).toEqual(["available", "subscribed"]);
    expect(
      filterScheduleBangumis(items, summaries, "subscribed").map(
        (item) => item.mikanBangumiID
      )
    ).toEqual(["subscribed"]);
  });
});
