import { describe, expect, it } from "vitest";

import {
  extractBangumiNameFromParseError,
  isTMDBMatchFailure,
  resolveDiscoveryBangumiName,
} from "./parse-fallback";

describe("extractBangumiNameFromParseError", () => {
  it("extracts bangumi name from TMDB not-found message", () => {
    expect(
      extractBangumiNameFromParseError(
        "未搜索到番剧，番剧名称: Clevatess II-魔兽之王与虚假的勇者传承-"
      )
    ).toBe("Clevatess II-魔兽之王与虚假的勇者传承-");
  });

  it("returns empty string when message has no bangumi name", () => {
    expect(extractBangumiNameFromParseError("网络错误")).toBe("");
  });
});

describe("isTMDBMatchFailure", () => {
  it("detects TMDB search miss errors", () => {
    expect(
      isTMDBMatchFailure(
        "未搜索到番剧，番剧名称: Clevatess II-魔兽之王与虚假的勇者传承-"
      )
    ).toBe(true);
    expect(isTMDBMatchFailure("网络错误")).toBe(false);
  });
});

describe("resolveDiscoveryBangumiName", () => {
  it("prefers matching detail name", () => {
    expect(
      resolveDiscoveryBangumiName({
        bangumiID: "1",
        detail: { mikanBangumiID: "1", name: "详情名" },
        bangumis: [{ mikanBangumiID: "1", name: "季度名" }],
        errorMessage: "未搜索到番剧，番剧名称: 错误名",
      })
    ).toBe("详情名");
  });

  it("falls back to season list, then search, then error message", () => {
    expect(
      resolveDiscoveryBangumiName({
        bangumiID: "2",
        bangumis: [{ mikanBangumiID: "2", name: "季度名" }],
      })
    ).toBe("季度名");

    expect(
      resolveDiscoveryBangumiName({
        bangumiID: "3",
        searchBangumis: [{ mikanBangumiID: "3", name: "搜索名" }],
      })
    ).toBe("搜索名");

    expect(
      resolveDiscoveryBangumiName({
        bangumiID: "4",
        errorMessage: "未搜索到番剧，番剧名称: 错误名",
      })
    ).toBe("错误名");
  });
});
