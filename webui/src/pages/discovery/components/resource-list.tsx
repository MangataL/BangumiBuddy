import { Download } from "lucide-react";

import type { ResourceCandidate } from "@/api/discovery";
import { Button } from "@/components/ui/button";

import { formatDate } from "../discovery-options";

export function ResourceList({
  resources,
  onAddResource,
}: {
  resources: ResourceCandidate[];
  onAddResource: (resource: ResourceCandidate) => void;
}) {
  if (resources.length === 0) {
    return (
      <p className="rounded-xl bg-muted/60 px-4 py-6 text-center text-sm text-muted-foreground">
        暂无可用资源
      </p>
    );
  }

  return (
    <div className="divide-y divide-border/70">
      {resources.map((resource) => (
        <div
          key={`${resource.title}-${resource.magnetLink}`}
          className="flex flex-col gap-3 py-4 first:pt-0 sm:flex-row sm:items-center"
        >
          <div className="min-w-0 flex-1">
            <p className="break-all text-sm font-medium leading-5">
              {resource.title}
            </p>
            <div className="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted-foreground">
              {resource.releaseGroup && <span>{resource.releaseGroup}</span>}
              {resource.size && <span>{resource.size}</span>}
              {resource.publishedAt && (
                <span>{formatDate(resource.publishedAt)}</span>
              )}
              <span>
                {resource.suggestedDownloadType === "movie"
                  ? "剧场版"
                  : "番剧"}
              </span>
            </div>
          </div>
          <Button
            type="button"
            variant="outline"
            size="sm"
            className="shrink-0 rounded-lg sm:self-center"
            aria-label={`下载资源 ${resource.title}`}
            onClick={() => onAddResource(resource)}
          >
            <Download className="h-4 w-4" />
            下载
          </Button>
        </div>
      ))}
    </div>
  );
}
