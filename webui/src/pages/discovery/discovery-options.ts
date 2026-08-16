import type {
  BangumiCandidate,
  ReleaseGroupSummary,
  Season,
} from "@/api/discovery";
import { SeasonSet } from "@/api/discovery";

import {
  getSubscribedReleaseGroupCount,
  hasLoadedNoReleaseGroups,
} from "./release-group-state";
import type { WeekdayOption } from "./weekday-schedule";

export const seasonOptions: { value: Season; label: string }[] = [
  { value: SeasonSet.Winter, label: "冬季" },
  { value: SeasonSet.Spring, label: "春季" },
  { value: SeasonSet.Summer, label: "夏季" },
  { value: SeasonSet.Fall, label: "秋季" },
];

export const weekdayOptions: WeekdayOption[] = [
  { value: 1, label: "星期一" },
  { value: 2, label: "星期二" },
  { value: 3, label: "星期三" },
  { value: 4, label: "星期四" },
  { value: 5, label: "星期五" },
  { value: 6, label: "星期六" },
  { value: 0, label: "星期日" },
];

export type ScheduleFilter = "all" | "available" | "subscribed";

export function isMovieBangumi(item: BangumiCandidate) {
  return item.weekday === 7;
}

export function getCurrentSeason(date = new Date()): Season {
  const month = date.getMonth() + 1;
  if (month <= 3) return SeasonSet.Winter;
  if (month <= 6) return SeasonSet.Spring;
  if (month <= 9) return SeasonSet.Summer;
  return SeasonSet.Fall;
}

export function groupBangumisByWeekday(
  bangumis: BangumiCandidate[]
): Record<number, BangumiCandidate[]> {
  return bangumis.reduce<Record<number, BangumiCandidate[]>>((groups, item) => {
    const weekday = item.weekday ?? 0;
    groups[weekday] = groups[weekday] ?? [];
    groups[weekday].push(item);
    return groups;
  }, {});
}

export function filterScheduleBangumis(
  bangumis: BangumiCandidate[],
  summaries: Record<string, ReleaseGroupSummary>,
  filter: ScheduleFilter
): BangumiCandidate[] {
  if (filter === "available") {
    return bangumis.filter((item) => {
      const summary = summaries[item.mikanBangumiID];
      return !hasLoadedNoReleaseGroups(item, summary);
    });
  }
  if (filter === "subscribed") {
    return bangumis.filter(
      (item) =>
        getSubscribedReleaseGroupCount(summaries[item.mikanBangumiID]) > 0
    );
  }
  return bangumis;
}

export function formatWeekdayLabel(value: number) {
  return weekdayOptions.find((weekday) => weekday.value === value)?.label ?? "放送日未知";
}

export function formatDate(value?: string) {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleDateString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  });
}
