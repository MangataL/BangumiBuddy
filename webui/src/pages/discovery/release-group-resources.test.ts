import { describe, expect, it } from "vitest";

import type {
  ReleaseGroupCandidate,
  ResourceCandidate,
} from "@/api/discovery";

import { getResourcesForReleaseGroup } from "./release-group-resources";

const group: ReleaseGroupCandidate = {
  mikanSubgroupID: "124",
  name: "夜莺工作室",
  subscribed: false,
};

function resource(
  title: string,
  mikanSubgroupID?: string,
  releaseGroup = ""
): ResourceCandidate {
  return {
    title,
    mikanSubgroupID,
    magnetLink: `magnet:?xt=urn:btih:${title}`,
    size: "500MB",
    releaseGroup,
    suggestedDownloadType: "tv",
  };
}

describe("release group resources", () => {
  it("优先使用蜜柑字幕组 ID 关联资源", () => {
    const resources = [
      resource("目标资源", "124", "标题中的其他名称"),
      resource("其他资源", "395", "夜莺工作室"),
    ];

    expect(
      getResourcesForReleaseGroup(group, resources).map((item) => item.title)
    ).toEqual(["目标资源"]);
  });

  it("资源缺少字幕组 ID 时按名称关联", () => {
    const resources = [
      resource("目标资源", undefined, "【夜莺工作室】"),
      resource("其他资源", undefined, "哆啦字幕组"),
    ];

    expect(
      getResourcesForReleaseGroup(group, resources).map((item) => item.title)
    ).toEqual(["目标资源"]);
  });
});
