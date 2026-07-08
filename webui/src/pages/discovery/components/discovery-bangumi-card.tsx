import type { ReactNode } from "react";

import type { BangumiCandidate } from "@/api/discovery";
import { cn } from "@/lib/utils";

import { PosterImage } from "./poster-image";

interface DiscoveryBangumiCardProps {
  item: BangumiCandidate;
  muted?: boolean;
  badge?: ReactNode;
  children?: ReactNode;
  onOpenDetail: (id: string) => void;
}

export function DiscoveryBangumiCard({
  item,
  muted = false,
  badge,
  children,
  onOpenDetail,
}: DiscoveryBangumiCardProps) {
  const openDetail = () => onOpenDetail(item.mikanBangumiID);

  return (
    <article className="discovery-poster group/card min-w-0">
      <button
        type="button"
        className={cn(
          "discovery-poster__frame relative block w-full overflow-hidden rounded-2xl text-left",
          "shadow-[0_1px_2px_rgba(15,23,42,0.05),0_8px_20px_-14px_rgba(15,23,42,0.35)]",
          "transition-[transform,box-shadow] duration-200 ease-[cubic-bezier(0.23,1,0.32,1)]",
          "hover:-translate-y-1 hover:shadow-[0_18px_36px_-18px_rgba(15,23,42,0.45)]",
          "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2 focus-visible:ring-offset-background",
          "active:translate-y-0 active:scale-[0.97] active:duration-100",
          "motion-reduce:transform-none motion-reduce:transition-none"
        )}
        onClick={openDetail}
        aria-label={`查看 ${item.name} 详情`}
      >
        <PosterImage
          src={item.posterURL}
          alt=""
          muted={muted}
          className="rounded-none"
          imageClassName="discovery-poster__image"
        />

        <div
          aria-hidden
          className="pointer-events-none absolute inset-0 bg-gradient-to-t from-black/85 via-black/30 via-40% to-transparent to-70%"
        />

        <div className="absolute inset-x-0 bottom-0 z-[1] px-2.5 pb-2.5 pt-12 sm:px-3 sm:pb-3">
          <span className="line-clamp-2 text-[13px] font-semibold leading-snug tracking-tight text-white [text-shadow:0_1px_2px_rgba(0,0,0,0.45)] sm:text-sm">
            {item.name}
          </span>
        </div>

        {badge}
      </button>

      {children ? (
        <div className="mt-2 min-w-0 px-0.5">{children}</div>
      ) : null}
    </article>
  );
}
