import { useMemo } from "react";
import { RefreshCw, Search, Sparkles } from "lucide-react";

import type { Season } from "@/api/discovery";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { cn } from "@/lib/utils";

import { seasonOptions } from "../discovery-options";

interface DiscoveryToolbarProps {
  year: number;
  season: Season;
  query: string;
  loading: boolean;
  onYearChange: (year: number) => void;
  onSeasonChange: (season: Season) => void;
  onQueryChange: (query: string) => void;
  onSearch: () => void;
  onRefresh: () => void;
}

type SeasonControlsDensity = "compact" | "comfortable";

interface SeasonControlsProps {
  density: SeasonControlsDensity;
  year: number;
  season: Season;
  years: number[];
  loading: boolean;
  onYearChange: (year: number) => void;
  onSeasonChange: (season: Season) => void;
  onRefresh: () => void;
}

function SeasonControls(props: SeasonControlsProps) {
  const compact = props.density === "compact";

  return (
    <div className={cn("flex items-center", compact ? "gap-1.5" : "gap-2")}>
      <Select
        value={String(props.year)}
        onValueChange={(value) => props.onYearChange(Number(value))}
      >
        <SelectTrigger
          aria-label="选择年份"
          className={cn(
            "bg-background",
            compact
              ? "h-8 w-[76px] rounded-lg px-2 text-xs"
              : "h-10 w-[104px] rounded-xl"
          )}
        >
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {props.years.map((year) => (
            <SelectItem key={year} value={String(year)}>
              {year}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      <Select
        value={props.season}
        onValueChange={(value) => props.onSeasonChange(value as Season)}
      >
        <SelectTrigger
          aria-label="选择季度"
          className={cn(
            "bg-background",
            compact
              ? "h-8 w-[68px] rounded-lg px-2 text-xs"
              : "h-10 w-[104px] rounded-xl"
          )}
        >
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {seasonOptions.map((item) => (
            <SelectItem key={item.value} value={item.value}>
              {item.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      <Button
        type="button"
        variant="outline"
        size="icon"
        className={cn(
          "bg-background",
          compact
            ? "h-8 w-8 rounded-lg"
            : "h-10 w-auto rounded-xl px-3"
        )}
        onClick={props.onRefresh}
        aria-label="刷新当前内容"
      >
        <RefreshCw
          className={cn(
            compact ? "h-3.5 w-3.5" : "h-4 w-4",
            props.loading && "animate-spin motion-reduce:animate-none"
          )}
        />
        {!compact && <span>刷新</span>}
      </Button>
    </div>
  );
}

export function DiscoveryToolbar(props: DiscoveryToolbarProps) {
  const years = useMemo(() => {
    const current = new Date().getFullYear();
    return Array.from({ length: 15 }, (_, index) => current - index);
  }, []);
  const seasonLabel =
    seasonOptions.find((item) => item.value === props.season)?.label ?? "";

  const seasonControlProps = {
    year: props.year,
    season: props.season,
    years,
    loading: props.loading,
    onYearChange: props.onYearChange,
    onSeasonChange: props.onSeasonChange,
    onRefresh: props.onRefresh,
  };

  return (
    <header className="grid flex-none gap-2.5 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-end sm:gap-4">
      <div className="order-1 flex items-center justify-between gap-3 sm:items-baseline sm:justify-start">
        <div className="flex items-baseline gap-3">
          <h1 className="flex items-center gap-1.5 text-xl font-bold anime-gradient-text xs:gap-2 sm:text-2xl md:text-3xl">
            <Sparkles className="h-4 w-4 flex-shrink-0 text-primary animate-pulse xs:h-5 xs:w-5 sm:h-6 sm:w-6" />
            <span>发现</span>
          </h1>
          <span className="hidden text-sm text-muted-foreground sm:inline">
            {props.year} · {seasonLabel}放送
          </span>
        </div>

        <div className="sm:hidden">
          <SeasonControls density="compact" {...seasonControlProps} />
        </div>
      </div>

      <div className="order-3 hidden sm:order-2 sm:block">
        <SeasonControls density="comfortable" {...seasonControlProps} />
      </div>

      <div className="order-2 flex gap-2 sm:order-3 sm:col-span-2">
        <div className="relative min-w-0 flex-1">
          <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground sm:left-3.5" />
          <Input
            value={props.query}
            aria-label="搜索番剧和资源"
            placeholder="搜索番剧、剧场版或单集资源"
            className="h-10 rounded-xl border-border/80 bg-background pl-9 pr-3 shadow-none focus-visible:ring-inset focus-visible:ring-primary focus-visible:ring-offset-0 sm:h-11 sm:pl-10"
            onChange={(event) => props.onQueryChange(event.target.value)}
            onKeyDown={(event) => {
              if (event.key === "Enter") props.onSearch();
            }}
          />
        </div>
        <Button
          type="button"
          className="anime-button h-10 rounded-xl bg-gradient-to-r from-primary to-blue-500 px-3 text-white transition-transform duration-150 hover:opacity-90 active:scale-[0.97] motion-reduce:transition-none sm:h-11 sm:px-4"
          onClick={props.onSearch}
        >
          <Search className="h-4 w-4" />
          <span className="hidden sm:inline">搜索</span>
          <span className="sr-only sm:hidden">搜索</span>
        </Button>
      </div>
    </header>
  );
}
