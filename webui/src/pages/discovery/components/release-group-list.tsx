import { CheckCircle2 } from "lucide-react";

import type { ReleaseGroupCandidate } from "@/api/discovery";
import { Button } from "@/components/ui/button";

interface ReleaseGroupListProps {
  groups: ReleaseGroupCandidate[];
  onAdd: (group: ReleaseGroupCandidate) => void;
  onViewSubscription: (id: string) => void;
  compact?: boolean;
}

export function ReleaseGroupList(props: ReleaseGroupListProps) {
  if (props.groups.length === 0) {
    return (
      <p className="rounded-xl bg-muted/60 px-4 py-6 text-center text-sm text-muted-foreground">
        暂无可订阅字幕组
      </p>
    );
  }

  return (
    <div className="space-y-2">
      {props.groups.map((group) => (
        <div
          key={group.mikanSubgroupID}
          className="flex items-center gap-3 rounded-xl border border-border/80 bg-background px-3 py-2.5"
        >
          <div className="min-w-0 flex-1">
            <p className="truncate text-sm font-medium">{group.name}</p>
          </div>
          {group.subscribed ? (
            group.subscriptionID ? (
              <Button
                type="button"
                variant="outline"
                size="sm"
                className="shrink-0 rounded-lg"
                aria-label={`查看 ${group.name} 的订阅`}
                onClick={() => props.onViewSubscription(group.subscriptionID!)}
              >
                <CheckCircle2 className="h-4 w-4 text-emerald-600 dark:text-emerald-400" />
                {!props.compact && "查看"}
              </Button>
            ) : (
              <span className="inline-flex shrink-0 items-center gap-1 text-xs font-medium text-emerald-600 dark:text-emerald-400">
                <CheckCircle2 className="h-4 w-4" />
                已订阅
              </span>
            )
          ) : (
            <Button
              type="button"
              size="sm"
              className="shrink-0 rounded-lg bg-gradient-to-r from-primary to-blue-500 text-white hover:opacity-90"
              aria-label={`订阅字幕组 ${group.name}`}
              onClick={() => props.onAdd(group)}
            >
              订阅
            </Button>
          )}
        </div>
      ))}
    </div>
  );
}
