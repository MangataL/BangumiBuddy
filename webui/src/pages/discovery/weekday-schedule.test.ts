import { describe, expect, it } from "vitest";

import {
  getAdjacentWeekdayValue,
  getCurrentWeekday,
  getOrderedWeekdays,
  type WeekdayOption,
} from "./weekday-schedule";

const weekdays: WeekdayOption[] = [
  { value: 1, label: "星期一" },
  { value: 2, label: "星期二" },
  { value: 3, label: "星期三" },
  { value: 4, label: "星期四" },
  { value: 5, label: "星期五" },
  { value: 6, label: "星期六" },
  { value: 0, label: "星期日" },
];

describe("weekday schedule helpers", () => {
  it("会用真实日期换算当前星期", () => {
    expect(getCurrentWeekday(new Date("2026-07-08T12:00:00+08:00"))).toBe(3);
  });

  it("会把当前星期排在第一列", () => {
    expect(getOrderedWeekdays(weekdays, 3).map((weekday) => weekday.value)).toEqual([
      3, 4, 5, 6, 0, 1, 2,
    ]);
  });

  it("上一列和下一列会按当前排序循环", () => {
    expect(getAdjacentWeekdayValue(weekdays, 3, 1)).toBe(4);
    expect(getAdjacentWeekdayValue(weekdays, 3, -1)).toBe(2);
    expect(getAdjacentWeekdayValue(weekdays, 0, 1)).toBe(1);
  });
});
