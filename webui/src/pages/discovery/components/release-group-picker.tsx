import { useState } from "react";
import { Loader2, RefreshCw, Users } from "lucide-react";

import type {
  BangumiCandidate,
  ReleaseGroupCandidate,
  ReleaseGroupSummary,
} from "@/api/discovery";
import { Button } from "@/components/ui/button";
import {
  Drawer,
  DrawerContent,
  DrawerDescription,
  DrawerHeader,
  DrawerTitle,
  DrawerTrigger,
} from "@/components/ui/drawer";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { useMobile } from "@/hooks/useMobile";
import { cn } from "@/lib/utils";

import { getReleaseGroupPickerState } from "../release-group-picker-state";
import { ReleaseGroupList } from "./release-group-list";

interface ReleaseGroupPickerProps {
  item: BangumiCandidate;
  summary?: ReleaseGroupSummary;
  loading: boolean;
  error?: string;
  className?: string;
  onEnsureSummary: () => void;
  onAdd: (group: ReleaseGroupCandidate) => void;
  onViewSubscription: (id: string) => void;
}

export function ReleaseGroupPicker(props: ReleaseGroupPickerProps) {
  const isMobile = useMobile();
  const [open, setOpen] = useState(false);
  const pickerState = getReleaseGroupPickerState(
    props.item,
    props.summary,
    props.loading
  );

  const handleOpenChange = (nextOpen: boolean) => {
    setOpen(nextOpen);
    if (nextOpen && !props.summary && !pickerState.disabled) {
      props.onEnsureSummary();
    }
  };

  const handleAdd = (group: ReleaseGroupCandidate) => {
    setOpen(false);
    props.onAdd(group);
  };

  const trigger = (
    <Button
      type="button"
      variant="ghost"
      size="sm"
      disabled={pickerState.disabled}
      className={cn(
        "h-8 max-w-full justify-start gap-1.5 rounded-lg px-2 text-xs font-medium text-muted-foreground hover:bg-primary/10 hover:text-primary",
        pickerState.subscribed &&
          "text-emerald-600 dark:text-emerald-400",
        props.className
      )}
      aria-label={
        pickerState.disabled
          ? `${props.item.name} 暂无字幕组`
          : `为 ${props.item.name} 选择字幕组`
      }
    >
      {props.loading ? (
        <Loader2 className="h-3.5 w-3.5 animate-spin motion-reduce:animate-none" />
      ) : (
        <Users className="h-3.5 w-3.5" />
      )}
      <span className="truncate">
        {pickerState.label}
      </span>
    </Button>
  );

  const content = (
    <>
      {props.loading && !props.summary ? (
        <div className="flex min-h-28 items-center justify-center gap-2 text-sm text-muted-foreground">
          <Loader2 className="h-4 w-4 animate-spin motion-reduce:animate-none" />
          正在载入字幕组
        </div>
      ) : props.error && !props.summary ? (
        <div className="rounded-xl bg-destructive/5 p-4 text-center">
          <p className="text-sm text-muted-foreground">{props.error}</p>
          <Button
            type="button"
            variant="outline"
            size="sm"
            className="mt-3 rounded-lg"
            onClick={props.onEnsureSummary}
          >
            <RefreshCw className="h-4 w-4" />
            重试
          </Button>
        </div>
      ) : (
        <ReleaseGroupList
          groups={props.summary?.releaseGroups ?? []}
          compact
          onAdd={handleAdd}
          onViewSubscription={props.onViewSubscription}
        />
      )}
    </>
  );

  if (isMobile) {
    return (
      <Drawer
        open={open}
        onOpenChange={handleOpenChange}
        shouldScaleBackground={false}
      >
        <DrawerTrigger asChild>{trigger}</DrawerTrigger>
        <DrawerContent className="max-h-[82dvh] motion-reduce:animate-none motion-reduce:transition-none">
          <DrawerHeader className="text-left">
            <DrawerTitle className="pr-8">{props.item.name}</DrawerTitle>
            <DrawerDescription>选择要订阅的字幕组</DrawerDescription>
          </DrawerHeader>
          <div className="overflow-y-auto px-4 pb-6">{content}</div>
        </DrawerContent>
      </Drawer>
    );
  }

  return (
    <Popover open={open} onOpenChange={handleOpenChange}>
      <PopoverTrigger asChild>{trigger}</PopoverTrigger>
      <PopoverContent
        align="start"
        sideOffset={8}
        className="w-[360px] max-w-[calc(100vw-2rem)] rounded-2xl p-3 shadow-xl motion-reduce:animate-none motion-reduce:transition-none"
      >
        <div className="mb-3 px-1">
          <p className="line-clamp-2 text-sm font-semibold">{props.item.name}</p>
          <p className="mt-0.5 text-xs text-muted-foreground">
            选择要订阅的字幕组
          </p>
        </div>
        <div className="max-h-[360px] overflow-y-auto pr-1">{content}</div>
      </PopoverContent>
    </Popover>
  );
}
