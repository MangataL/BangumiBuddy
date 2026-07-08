import { AlertCircle, Compass, RefreshCw } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";

import { discoveryPosterGridClassName } from "../poster-grid";

export function PosterGridSkeleton({ count = 10 }: { count?: number }) {
  return (
    <div
      className={discoveryPosterGridClassName}
      aria-label="正在加载放送表"
    >
      {Array.from({ length: count }, (_, index) => (
        <div key={index} className="min-w-0">
          <Skeleton className="aspect-[2/3] w-full rounded-2xl motion-reduce:animate-none" />
          <Skeleton className="mt-2.5 h-3 w-3/5 rounded-full motion-reduce:animate-none" />
        </div>
      ))}
    </div>
  );
}

export function EmptyDiscoveryState({
  title,
  description,
}: {
  title: string;
  description: string;
}) {
  return (
    <div className="flex min-h-64 flex-col items-center justify-center px-6 text-center">
      <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-primary/10 text-primary">
        <Compass className="h-5 w-5" />
      </div>
      <h3 className="text-base font-semibold">{title}</h3>
      <p className="mt-1 max-w-sm text-sm leading-6 text-muted-foreground">
        {description}
      </p>
    </div>
  );
}

export function DiscoveryErrorState({
  title = "暂时无法加载",
  description,
  onRetry,
}: {
  title?: string;
  description: string;
  onRetry: () => void;
}) {
  return (
    <div className="flex min-h-64 flex-col items-center justify-center px-6 text-center">
      <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-destructive/10 text-destructive">
        <AlertCircle className="h-5 w-5" />
      </div>
      <h3 className="text-base font-semibold">{title}</h3>
      <p className="mt-1 max-w-md text-sm leading-6 text-muted-foreground">
        {description}
      </p>
      <Button variant="outline" className="mt-5 rounded-xl" onClick={onRetry}>
        <RefreshCw className="h-4 w-4" />
        重试
      </Button>
    </div>
  );
}
