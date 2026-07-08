import { Clapperboard } from "lucide-react";

import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

import { weekdayOptions } from "../discovery-options";
import { getOrderedWeekdays } from "../weekday-schedule";

interface BroadcastDayStripProps {
  activeWeekday: number;
  currentWeekday: number;
  counts: Record<number, number>;
  onChange: (weekday: number) => void;
  onBrowseMovies: () => void;
}

export function BroadcastDayStrip(props: BroadcastDayStripProps) {
  const orderedWeekdays = getOrderedWeekdays(
    weekdayOptions,
    props.currentWeekday
  );

  return (
    <div className="flex items-stretch gap-1.5 sm:gap-2">
      <div
        role="group"
        aria-label="按放送日筛选"
        className="scrollbar-hide flex min-w-0 flex-1 snap-x snap-mandatory gap-1.5 overflow-x-auto sm:gap-2"
      >
        {orderedWeekdays.map((weekday) => {
          const selected = weekday.value === props.activeWeekday;
          const count = props.counts[weekday.value] ?? 0;
          const weekdayShort = weekday.label.replace("星期", "周");
          return (
            <button
              key={weekday.value}
              type="button"
              aria-pressed={selected}
              aria-label={`${weekday.label}，${count} 部`}
              className={cn(
                "group inline-flex shrink-0 snap-start items-center gap-1.5 rounded-full px-2.5 py-1.5 text-left transition-[background-color,border-color,transform] duration-150 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-primary active:scale-[0.98] motion-reduce:transition-none sm:rounded-xl sm:px-3 sm:py-2",
                selected
                  ? "border-0 bg-gradient-to-r from-primary to-blue-500 text-white"
                  : "border border-border/80 bg-background hover:border-primary/40"
              )}
              onClick={() => props.onChange(weekday.value)}
            >
              <span className="whitespace-nowrap text-xs font-semibold sm:text-sm">
                {weekdayShort}
              </span>
              <span
                className={cn(
                  "rounded-full px-1.5 py-0.5 text-[11px] tabular-nums",
                  selected
                    ? "bg-white/15 text-white"
                    : "bg-muted text-muted-foreground"
                )}
              >
                {count}
              </span>
            </button>
          );
        })}
      </div>
      <Button
        type="button"
        variant="outline"
        className={cn(
          "h-auto shrink-0 gap-1.5 rounded-full px-2.5 py-1.5 focus-visible:ring-inset focus-visible:ring-offset-0 sm:rounded-xl sm:px-3 sm:py-2",
          props.activeWeekday === 7
            ? "border-0 bg-gradient-to-r from-primary to-blue-500 text-white hover:bg-gradient-to-r hover:from-primary hover:to-blue-500 hover:opacity-90 hover:text-white"
            : "bg-background"
        )}
        aria-pressed={props.activeWeekday === 7}
        aria-label={`剧场版，${props.counts[7] ?? 0} 部`}
        onClick={props.onBrowseMovies}
      >
        <Clapperboard
          className={cn(
            "h-3.5 w-3.5 sm:h-4 sm:w-4",
            props.activeWeekday === 7 ? "text-white" : "text-primary"
          )}
        />
        <span className="text-xs font-semibold sm:text-sm">剧场版</span>
        <span
          className={cn(
            "rounded-full px-1.5 py-0.5 text-[11px] font-normal tabular-nums",
            props.activeWeekday === 7
              ? "bg-white/15 text-white"
              : "bg-muted text-muted-foreground"
          )}
        >
          {props.counts[7] ?? 0}
        </span>
      </Button>
    </div>
  );
}
