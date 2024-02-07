export const weekDayMap: Record<string, string> = {
  "0": "星期日",
  "1": "星期一",
  "2": "星期二",
  "3": "星期三",
  "4": "星期四",
  "5": "星期五",
  "6": "星期六",
} as const;

export function getWeekDayText(day: string | number): string {
  const dayStr = day.toString();
  return weekDayMap[dayStr] || "";
}

export function isValidWeekDay(day: string | number): boolean {
  const dayStr = day.toString();
  return dayStr in weekDayMap;
}

export function getSortedWeekDays(): Array<[string, string]> {
  return Object.entries(weekDayMap).sort((a, b) => {
    if (a[0] === "0") return 1;
    if (b[0] === "0") return -1;
    return parseInt(a[0]) - parseInt(b[0]);
  });
}
