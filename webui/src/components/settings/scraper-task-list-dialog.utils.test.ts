import { describe, expect, it } from "vitest";

import type { ScraperTask } from "@/api/config";

import {
  filterScraperTasksByBangumiName,
  getScraperTaskBangumiNames,
} from "./scraper-task-list-dialog.utils";

function createTask(id: number, bangumiName: string): ScraperTask {
  return {
    id,
    tmdbID: id,
    filePath: `/tmp/${id}.mkv`,
    bangumiName,
    posterURL: "",
    season: 1,
    episode: id,
    statuses: ["pending"],
  };
}

describe("getScraperTaskBangumiNames", () => {
  it("返回去重后的番剧名称并保留原始顺序", () => {
    const tasks = [
      createTask(1, "孤独摇滚"),
      createTask(2, "少女乐队的呐喊"),
      createTask(3, "孤独摇滚"),
    ];

    expect(getScraperTaskBangumiNames(tasks)).toEqual([
      "孤独摇滚",
      "少女乐队的呐喊",
    ]);
  });

  it("会忽略空白番剧名称", () => {
    const tasks = [createTask(1, "  "), createTask(2, "BanG Dream! It's MyGO!!!!!")];

    expect(getScraperTaskBangumiNames(tasks)).toEqual([
      "BanG Dream! It's MyGO!!!!!",
    ]);
  });
});

describe("filterScraperTasksByBangumiName", () => {
  it("在未选择番剧时返回全部任务", () => {
    const tasks = [createTask(1, "孤独摇滚"), createTask(2, "少女乐队的呐喊")];

    expect(filterScraperTasksByBangumiName(tasks, "")).toEqual(tasks);
    expect(filterScraperTasksByBangumiName(tasks, "   ")).toEqual(tasks);
  });

  it("按选中的番剧名称精确筛选", () => {
    const tasks = [
      createTask(1, "BanG Dream! Ave Mujica"),
      createTask(2, "BanG Dream! Ave Mujica"),
      createTask(3, "Girls Band Cry"),
    ];

    expect(
      filterScraperTasksByBangumiName(tasks, "BanG Dream! Ave Mujica")
    ).toEqual([tasks[0], tasks[1]]);
  });

  it("筛选时会忽略选中值的首尾空白", () => {
    const tasks = [
      createTask(2, "Girls Band Cry"),
    ];

    expect(filterScraperTasksByBangumiName(tasks, "  Girls Band Cry  ")).toEqual([
      tasks[0],
    ]);
  });
});
