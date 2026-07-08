import { describe, expect, it } from "vitest";

import type {
  BangumiCandidate,
  ReleaseGroupSummary,
} from "@/api/discovery";

import { getReleaseGroupPickerState } from "./release-group-picker-state";

const candidate: BangumiCandidate = {
  mikanBangumiID: "1",
  name: "测试番剧",
  posterURL: "",
  weekday: 5,
  airStartDate: "",
};

function summary(
  releaseGroupCount: number,
  subscribedReleaseGroupCount = 0
): ReleaseGroupSummary {
  return {
    mikanBangumiID: "1",
    releaseGroups: Array.from({ length: releaseGroupCount }, (_, index) => ({
      mikanSubgroupID: String(index),
      name: `字幕组 ${index}`,
      subscribed: index < subscribedReleaseGroupCount,
    })),
  };
}

describe("release group picker state", () => {
  it("区分载入中、可选择和已订阅状态", () => {
    expect(getReleaseGroupPickerState(candidate, undefined, true)).toEqual({
      disabled: false,
      label: "正在载入",
      subscribed: false,
    });
    expect(getReleaseGroupPickerState(candidate, summary(4), false)).toEqual({
      disabled: false,
      label: "4 个字幕组",
      subscribed: false,
    });
    expect(getReleaseGroupPickerState(candidate, summary(4, 1), false)).toEqual({
      disabled: false,
      label: "已订阅 1/4",
      subscribed: true,
    });
  });

  it("确认无字幕组后禁用入口", () => {
    expect(getReleaseGroupPickerState(candidate, summary(0), false)).toEqual({
      disabled: true,
      label: "暂无字幕组",
      subscribed: false,
    });
  });
});
