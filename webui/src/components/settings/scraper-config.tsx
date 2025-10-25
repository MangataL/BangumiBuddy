import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import {
  HybridTooltip,
  HybridTooltipContent,
  HybridTooltipTrigger,
} from "@/components/common/tooltip";
import { TooltipProvider } from "@/components/ui/tooltip";
import { Info } from "lucide-react";
import type { ScraperConfig } from "@/api/config";

interface ScraperConfigProps {
  scraperConfig: ScraperConfig;
  setScraperConfig: React.Dispatch<React.SetStateAction<ScraperConfig>>;
  loading: boolean;
}

export function ScraperConfigComponent({
  scraperConfig,
  setScraperConfig,
  loading,
}: ScraperConfigProps) {
  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2">
        <h3 className="text-lg font-semibold">媒体库刮削配置</h3>
      </div>

      <div className="space-y-4 pl-4 border-l-2 border-primary/10">
        {/* 缺失元数据填充开关 */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Label htmlFor="scraper-enable">缺失元数据填充</Label>
            <TooltipProvider>
              <HybridTooltip>
                <HybridTooltipTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-5 w-5 rounded-full"
                  >
                    <Info className="h-3.5 w-3.5 text-muted-foreground" />
                  </Button>
                </HybridTooltipTrigger>
                <HybridTooltipContent>
                  <p>
                    当前只对订阅番剧生效（用于新番添加到媒体库时tmdb数据不全而刮削错误或缺失数据），
                    <br />
                    会定时去检查媒体库刮削生成的nfo文件，如果标题、剧情简介和图片与tmdb数据不一致则更新，
                    <br />
                    只对打开后的新增番剧生效，并且当上述数据都更新成功后不再检查更新。
                  </p>
                </HybridTooltipContent>
              </HybridTooltip>
            </TooltipProvider>
          </div>
          <Switch
            id="scraper-enable"
            checked={scraperConfig.enable}
            onCheckedChange={(checked) =>
              setScraperConfig((prev) => ({
                ...prev,
                enable: checked,
              }))
            }
            disabled={loading}
          />
        </div>

        {/* 检查时间间隔配置 - 只在启用时显示 */}
        {scraperConfig.enable && (
          <div className="space-y-2 pl-4 border-l-2 border-primary/20">
            <div className="flex items-center gap-2">
              <Label htmlFor="check-interval">检查时间间隔（小时）</Label>
              <TooltipProvider>
                <HybridTooltip>
                  <HybridTooltipTrigger asChild>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-5 w-5 rounded-full"
                    >
                      <Info className="h-3.5 w-3.5 text-muted-foreground" />
                    </Button>
                  </HybridTooltipTrigger>
                  <HybridTooltipContent>
                    <p>
                      设置检查元数据的时间间隔，单位为小时。
                      <br />
                      建议设置为24小时（即1天）
                    </p>
                  </HybridTooltipContent>
                </HybridTooltip>
              </TooltipProvider>
            </div>
            <Input
              id="check-interval"
              type="number"
              min="1"
              value={scraperConfig.checkInterval}
              onChange={(e) => {
                const hours = parseInt(e.target.value, 10);
                if (!isNaN(hours) && hours > 0) {
                  setScraperConfig((prev) => ({
                    ...prev,
                    checkInterval: hours,
                  }));
                }
              }}
              disabled={loading}
              className="w-full"
            />
          </div>
        )}
      </div>
    </div>
  );
}
