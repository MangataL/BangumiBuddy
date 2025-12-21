import { Info } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { TooltipProvider } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";
import {
  HybridTooltip,
  HybridTooltipContent,
  HybridTooltipTrigger,
} from "@/components/common/tooltip";

interface EpisodePositionInputProps {
  value: string;
  onChange: (value: string) => void;
  label?: string;
  className?: string;
  inputClassName?: string;
  labelClassName?: string;
  labelContainerClassName?: string;
  showTooltipButton?: boolean;
}

export function EpisodePositionInput({
  value,
  onChange,
  label = "集数定位",
  className = "space-y-2",
  inputClassName,
  labelClassName,
  labelContainerClassName,
  showTooltipButton = true,
}: EpisodePositionInputProps) {
  return (
    <div className={className}>
      <div className={cn("flex items-center gap-2", labelContainerClassName)}>
        <Label htmlFor="episodePosition" className={labelClassName}>
          {label}
        </Label>
        <TooltipProvider>
          <HybridTooltip>
            <HybridTooltipTrigger asChild={showTooltipButton}>
              {showTooltipButton ? (
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-5 w-5 rounded-full"
                >
                  <Info className="h-3.5 w-3.5 text-muted-foreground" />
                </Button>
              ) : (
                <Info className="h-3 w-3 text-muted-foreground/60 hover:text-blue-500 transition-colors" />
              )}
            </HybridTooltipTrigger>
            <HybridTooltipContent className="max-w-xs space-y-2">
              <p>
                集数定位用来确定文件名中集数的位置，用于部分命名不规范解析集数出错的文件，一般不需要设置。
              </p>
              <p>
                集数定位使用
                <code className="text-red-500 bg-gray-100/10 px-1 rounded">
                  {"{ep}"}
                </code>
                来表示集数的位置，例如：
              </p>
              <p>
                <code className="text-red-500 bg-gray-100/10 px-1 rounded">
                  {"前缀{ep}后缀"}
                </code>
                可以用来定位{" "}
                <code className="text-red-500 bg-gray-100/10 px-1 rounded">
                  {"前缀01后缀.mp4"}
                </code>
                中的01，获取到正确的集数
              </p>
            </HybridTooltipContent>
          </HybridTooltip>
        </TooltipProvider>
      </div>
      <Input
        id="episodePosition"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className={cn("rounded-xl", inputClassName)}
      />
    </div>
  );
}
