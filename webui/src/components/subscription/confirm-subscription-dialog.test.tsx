// @vitest-environment jsdom
import { useState } from "react";
import { MemoryRouter } from "react-router-dom";
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
} from "@testing-library/react";
import { afterEach, beforeAll, beforeEach, expect, it, vi } from "vitest";

import { http } from "@/api";
import { discoveryAPI } from "@/api/discovery";
import DiscoveryPage from "@/pages/discovery";
import metaAPI from "@/api/meta";
import { subscriptionAPI, type ParseRSSResponse } from "@/api/subscription";
import { AddSubscriptionDialog } from "./add-subscription-dialog";
import { ConfirmSubscriptionDialog } from "./confirm-subscription-dialog";

const rss: ParseRSSResponse = {
  name: "测试番剧",
  year: "2020",
  tmdbID: 100,
  rssLink: "https://example.test/feed",
  releaseGroup: "字幕组 A",
  season: 1,
  episodeTotalNum: 12,
  airWeekday: 3,
  posterURL: "first.jpg",
  backdropURL: "background.jpg",
  recommendationDate: "2020-04-05",
  seasons: [
    { number: 0, airDate: "", airWeekday: null, episodeTotalNum: 2, posterURL: "special.jpg" },
    {
      number: 1,
      airDate: "2020-04-01",
      airWeekday: 3,
      episodeTotalNum: 12,
      posterURL: "first.jpg",
    },
    {
      number: 3,
      airDate: "2022-10-01",
      airWeekday: 6,
      episodeTotalNum: 24,
      posterURL: "third.jpg",
    },
    { number: 4, airDate: "", airWeekday: null, episodeTotalNum: 0, posterURL: "series.jpg" },
  ],
};

beforeAll(() => {
  // jsdom 没有 Radix 下拉列表使用的布局 API。
  Element.prototype.scrollIntoView = () => {};
  vi.stubGlobal(
    "ResizeObserver",
    class {
      observe() {}
      unobserve() {}
      disconnect() {}
    }
  );
});
beforeEach(() => {
  vi.spyOn(subscriptionAPI, "previewRSSMatch").mockResolvedValue([]);
  vi.spyOn(http, "post").mockResolvedValue({});
});
afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

const submit = () =>
  screen.getByRole<HTMLButtonElement>("button", { name: "确认订阅" });
const episodes = () => screen.getByLabelText<HTMLInputElement>("总集数");

const weekday = () => screen.getByRole("combobox", { name: "更新时间" });

function pickWeekday(label: string) {
  fireEvent.keyDown(weekday(), { key: "Enter" });
  fireEvent.click(screen.getByRole("option", { name: label }));
}

function openSeasons() {
  fireEvent.keyDown(screen.getByRole("combobox", { name: "季度" }), {
    key: "Enter",
  });
}

function input(element: HTMLElement, value: string) {
  fireEvent.change(element, { target: { value } });
}

function openConfirmation(result = rss) {
  render(
    <ConfirmSubscriptionDialog
      parseRSSRsp={result}
      defaultPriority={5}
      onClose={() => {}}
    />
  );
}

it("保留推荐季，切季后使用该季海报，并且仅提交最终的订阅字段", async () => {
  openConfirmation();
  expect(screen.getByRole("combobox", { name: "季度" }).textContent).toBe(
    "第 1 季"
  );
  expect(screen.queryByText(/推荐第/)).toBeNull();
  expect(episodes().value).toBe("12");
  expect(weekday().textContent).toBe("周三");
  openSeasons();
  expect(screen.getAllByRole("option").map((item) => item.textContent)).toEqual(
    ["第 1 季", "第 3 季", "第 4 季"]
  );
  fireEvent.click(screen.getByRole("option", { name: "第 3 季" }));
  expect(screen.getByRole("img").getAttribute("src")).toBe("third.jpg");
  expect(screen.getByText("2022")).toBeTruthy();
  expect(episodes().value).toBe("24");
  expect(weekday().textContent).toBe("周六");
  input(episodes(), "21");
  await act(async () => fireEvent.click(submit()));
  expect(http.post).toHaveBeenCalledWith("/bangumis", {
    rssLink: rss.rssLink,
    releaseGroup: "字幕组 A",
    tmdbID: 100,
    season: 3,
    episodeTotalNum: 21,
    priority: 5,
    episodeOffset: 0,
    airWeekday: 6,
    episodeLocation: "",
    includeRegs: [],
    excludeRegs: [],
  });
});

it("未推荐季时等待手选，没有集数的季需要填写有效集数", () => {
  openConfirmation({
    ...rss,
    season: 0,
    episodeTotalNum: 0,
    posterURL: "series.jpg",
    recommendationDate: undefined,
  });
  expect(screen.getByText("请选择季度")).toBeTruthy();
  expect(submit().disabled).toBe(true);
  openSeasons();
  fireEvent.click(screen.getByRole("option", { name: "第 4 季" }));
  expect(screen.getByRole("img").getAttribute("src")).toBe("series.jpg");
  expect(submit().disabled).toBe(true);
  for (const value of ["-1", "0.5", "", "0"]) {
    input(episodes(), value);
    expect(submit().disabled).toBe(true);
  }
  input(episodes(), "3");
  expect(weekday().textContent).toBe("请选择更新星期");
  expect(submit().disabled).toBe(true);
  pickWeekday("周日");
  expect(submit().disabled).toBe(false);
  input(screen.getByLabelText("字幕组"), "   ");
  expect(submit().disabled).toBe(true);
});

it("重新确认和搜索 TMDB 时使用当前 RSS，忽略先前 ID 的迟到结果", async () => {
  const earlier = Promise.withResolvers<ParseRSSResponse>();
  const parse = vi
    .spyOn(subscriptionAPI, "parseRSS")
    .mockReturnValueOnce(earlier.promise)
    .mockResolvedValueOnce({
      ...rss,
      name: "另一部番剧",
      tmdbID: 300,
      season: 0,
      seasons: [],
      recommendationDate: undefined,
    });
  vi.spyOn(metaAPI, "searchTVs").mockResolvedValue([
    {
      tmdbID: 300,
      chineseName: "新搜索结果",
      year: "",
      season: 7,
      episodeTotalNum: 52,
      posterURL: "",
      backdropURL: "",
      airWeekday: null,
      overview: "",
      genres: "",
    },
  ]);
  openConfirmation();
  input(screen.getByPlaceholderText("输入 TMDB ID 或点击搜索"), "200");
  expect(screen.queryByRole("img")).toBeNull();
  expect(submit().disabled).toBe(true);
  fireEvent.click(screen.getByTitle("确认获取TMDB信息"));
  fireEvent.click(screen.getByTitle("搜索TMDB"));
  input(screen.getByPlaceholderText("请输入番剧名称"), "新搜索结果");
  fireEvent.click(screen.getByRole("button", { name: "搜索" }));
  fireEvent.click(await screen.findByRole("button", { name: /选择此项/ }));
  expect(await screen.findByDisplayValue("另一部番剧")).toBeTruthy();
  for (const tmdbID of [200, 300]) {
    expect(parse).toHaveBeenCalledWith(
      { rssLink: rss.rssLink, tmdbID },
      expect.any(AbortSignal)
    );
  }
  const season = screen.getByLabelText<HTMLInputElement>("季度");
  expect(season.value).toBe("");
  input(season, "-2");
  input(episodes(), "8");
  expect(submit().disabled).toBe(true);
  input(season, "2");
  pickWeekday("周一");
  await act(async () => earlier.resolve({ ...rss, tmdbID: 200 }));
  expect(season.value).toBe("2");
  expect(episodes().value).toBe("8");
  expect(screen.getByDisplayValue("另一部番剧")).toBeTruthy();
  expect(submit().disabled).toBe(false);
});

it("RSS 入口每次打开都使用本次解析的季；关闭后迟到的解析不会重开确认框", async () => {
  const earlier = Promise.withResolvers<ParseRSSResponse>();
  vi.spyOn(subscriptionAPI, "parseRSS")
    .mockReturnValueOnce(earlier.promise)
    .mockResolvedValueOnce(rss)
    .mockResolvedValueOnce({
      ...rss,
      name: "下一个订阅",
      season: 0,
      recommendationDate: undefined,
    });
  function AddSubscription() {
    const [open, setOpen] = useState(true);
    return (
      <>
        <button onClick={() => setOpen(true)}>添加</button>
        <AddSubscriptionDialog open={open} onOpenChange={setOpen} />
      </>
    );
  }
  const start = () => {
    input(screen.getByLabelText("订阅链接"), rss.rssLink);
    fireEvent.click(screen.getByRole("button", { name: "解析" }));
  };
  const reopen = () =>
    fireEvent.click(screen.getByRole("button", { name: "添加" }));
  render(<AddSubscription />);
  start();
  fireEvent.click(screen.getByRole("button", { name: "取消" }));
  reopen();
  start();
  await screen.findByText("订阅确认");
  openSeasons();
  fireEvent.click(screen.getByRole("option", { name: /第 3 季/ }));
  fireEvent.click(screen.getByRole("button", { name: "取消" }));
  reopen();
  start();
  await screen.findByDisplayValue("下一个订阅");
  expect(screen.getByText("请选择季度")).toBeTruthy();
  expect(episodes().value).toBe("");
  await act(async () => earlier.resolve({ ...rss, name: "已关闭的订阅" }));
  expect(screen.queryByDisplayValue("已关闭的订阅")).toBeNull();
  expect(submit().disabled).toBe(true);
});

it("发现入口使用字幕组的解析结果和原有优先级", async () => {
  vi.spyOn(discoveryAPI, "listBangumis").mockResolvedValue([
    {
      mikanBangumiID: "10",
      name: "发现番剧",
      weekday: new Date().getDay(),
      posterURL: "",
      airStartDate: "",
    },
  ]);
  vi.spyOn(discoveryAPI, "batchReleaseGroups").mockResolvedValue({
    items: [
      {
        mikanBangumiID: "10",
        releaseGroups: [
          {
            mikanSubgroupID: "20",
            name: "字幕组 A",
            subscribed: false,
            previousMaxPriority: 9,
          },
        ],
      },
    ],
    failures: [],
  });
  vi.spyOn(discoveryAPI, "parseCandidateRSS").mockResolvedValue({
    ...rss,
    season: 3,
    episodeTotalNum: 24,
    posterURL: "third.jpg",
    recommendationDate: undefined,
  });
  render(
    <MemoryRouter>
      <DiscoveryPage />
    </MemoryRouter>
  );
  fireEvent.click(
    await screen.findByRole("button", { name: "为 发现番剧 选择字幕组" })
  );
  fireEvent.click(
    await screen.findByRole("button", { name: "订阅字幕组 字幕组 A" })
  );
  await screen.findByText("订阅确认");
  expect(screen.getByRole("combobox", { name: "季度" }).textContent).toBe(
    "第 3 季"
  );
  await act(async () => fireEvent.click(submit()));
  expect(http.post).toHaveBeenCalledWith(
    "/bangumis",
    expect.objectContaining({ priority: 9, season: 3, episodeTotalNum: 24 })
  );
});
