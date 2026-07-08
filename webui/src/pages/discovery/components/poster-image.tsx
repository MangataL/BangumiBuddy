import { useEffect, useState } from "react";
import { Tv } from "lucide-react";

import { AspectRatio } from "@/components/ui/aspect-ratio";
import { cn } from "@/lib/utils";

export function PosterImage({
  src,
  alt,
  muted = false,
  className,
  imageClassName,
}: {
  src: string;
  alt: string;
  muted?: boolean;
  className?: string;
  imageClassName?: string;
}) {
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    setFailed(false);
  }, [src]);

  return (
    <AspectRatio
      ratio={2 / 3}
      className={cn(
        "overflow-hidden bg-muted/60",
        muted && "bg-muted/50",
        className
      )}
    >
      {src && !failed ? (
        <img
          src={src}
          alt={alt}
          loading="lazy"
          className={cn(
            "h-full w-full object-cover object-center transition-[transform,filter,opacity] duration-300 ease-[cubic-bezier(0.23,1,0.32,1)] motion-reduce:transition-none",
            muted && "opacity-55 grayscale",
            imageClassName
          )}
          onError={() => setFailed(true)}
        />
      ) : (
        <div className="flex h-full w-full flex-col items-center justify-center gap-2 text-muted-foreground/70">
          <Tv className="h-5 w-5" strokeWidth={1.5} />
          <span className="text-[11px] tracking-wide">暂无海报</span>
        </div>
      )}
    </AspectRatio>
  );
}
