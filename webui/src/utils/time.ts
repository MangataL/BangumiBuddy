import { toZonedTime, format } from "date-fns-tz";

export const weekDayMap: Record<string, string> = {
  "0": "周日",
  "1": "周一",
  "2": "周二",
  "3": "周三",
  "4": "周四",
  "5": "周五",
  "6": "周六",
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

export function formatDate(dateString: string) {
  const date = new Date(dateString);
  const timeZone = Intl.DateTimeFormat().resolvedOptions().timeZone;
  const zonedDate = toZonedTime(date, timeZone);
  try {
    return format(zonedDate, "yyyy/MM/dd HH:mm", { timeZone });
  } catch (error) {
    console.error(error);
    return "";
  }
}
