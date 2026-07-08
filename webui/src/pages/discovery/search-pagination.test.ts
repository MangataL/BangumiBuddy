import { describe, expect, it } from "vitest";

import { getSearchPageCount } from "./search-pagination";

describe("search pagination", () => {
  it("把大量资源限制为固定页数", () => {
    expect(getSearchPageCount(1000, 20)).toBe(50);
    expect(getSearchPageCount(21, 20)).toBe(2);
  });

  it("空结果和无效页大小仍保持单页", () => {
    expect(getSearchPageCount(0, 20)).toBe(1);
    expect(getSearchPageCount(100, 0)).toBe(1);
  });
});
