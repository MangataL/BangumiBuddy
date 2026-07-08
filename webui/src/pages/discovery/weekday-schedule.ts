export type WeekdayOption = {
  value: number;
  label: string;
};

export function getCurrentWeekday(date = new Date()) {
  return date.getDay();
}

export function getOrderedWeekdays(weekdays: WeekdayOption[], currentWeekday: number) {
  const currentIndex = weekdays.findIndex((weekday) => weekday.value === currentWeekday);
  if (currentIndex < 0) return weekdays;
  return [...weekdays.slice(currentIndex), ...weekdays.slice(0, currentIndex)];
}

export function getAdjacentWeekdayValue(
  weekdays: WeekdayOption[],
  currentWeekday: number,
  direction: -1 | 1
) {
  const currentIndex = weekdays.findIndex((weekday) => weekday.value === currentWeekday);
  if (currentIndex < 0) return weekdays[0]?.value;
  const nextIndex = (currentIndex + direction + weekdays.length) % weekdays.length;
  return weekdays[nextIndex]?.value;
}
