import { useState, useEffect } from "react";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import {
  HybridTooltip,
  HybridTooltipContent,
  HybridTooltipTrigger,
} from "@/components/common/tooltip";
import { TooltipProvider } from "@/components/ui/tooltip";
import {
  Database,
  Loader2,
  Info,
  RefreshCw,
  CheckCircle,
  AlertCircle,
} from "lucide-react";
import { useToast } from "@/hooks/useToast";
import {
  configAPI,
  type SubtitleOperatorConfig,
  type FontMetaSetStats,
} from "@/api/config";
import { extractErrorMessage } from "@/utils/error";

interface SubtitleConfigProps {
  subtitleConfig: SubtitleOperatorConfig;
  setSubtitleConfig: React.Dispatch<
    React.SetStateAction<SubtitleOperatorConfig>
  >;
  loading: boolean;
  onFontStatsUpdate?: () => void;
}

export function SubtitleConfig({
  subtitleConfig,
  setSubtitleConfig,
  loading,
  onFontStatsUpdate,
}: SubtitleConfigProps) {
  const { toast } = useToast();
  const [initLoading, setInitLoading] = useState(false);
  const [statsLoading, setStatsLoading] = useState(false);
  const [fontStats, setFontStats] = useState<FontMetaSetStats>({
    total: 0,
    initDone: false,
  });

  // 获取字体库状态
  const loadFontStats = async () => {
    try {
      setStatsLoading(true);
      const stats = await configAPI.getSubtitleFontMetaSetStats();
      setFontStats(stats);
    } catch (error) {
      const description = extractErrorMessage(error);
      toast({
        title: "获取字体库状态失败",
        description,
        variant: "destructive",
      });
    } finally {
      setStatsLoading(false);
    }
  };

  // 初始化字体库
  const initFontMetaSet = async () => {
    try {
      setInitLoading(true);
      await configAPI.initSubtitleFontMetaSet();
      toast({
        title: "初始化成功",
        description: "字体库已成功初始化",
        variant: "default",
      });
      // 重新加载状态
      await loadFontStats();
      // 通知父组件更新字体库状态
      onFontStatsUpdate?.();
    } catch (error) {
      const description = extractErrorMessage(error);
      toast({
        title: "初始化失败",
        description,
        variant: "destructive",
      });
    } finally {
      setInitLoading(false);
    }
  };

  // 组件加载时获取状态
  useEffect(() => {
    loadFontStats();
  }, []);

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2">
        <h3 className="text-lg font-semibold">字幕子集化配置</h3>
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
                将字幕文件中使用的字体提取并嵌入到字幕文件中，
                <br />
                确保字幕在不同设备上显示效果一致，
                <br />
                避免客户端因字体缺失而无法按照预期显示字幕，
                <br />
                使用子集化会使得字幕文件的大小增加
                <br />
              </p>
            </HybridTooltipContent>
          </HybridTooltip>
        </TooltipProvider>
      </div>

      <div className="space-y-4 pl-4 border-l-2 border-primary/10">
        {/* 使用OTF字体 */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Label htmlFor="use-otf">使用OTF字体</Label>
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
                    标准的ass字幕是不支持内嵌OTF字体的，
                    <br />
                    开启这个选项可能会导致部分客户端无法正常显示字幕
                  </p>
                </HybridTooltipContent>
              </HybridTooltip>
            </TooltipProvider>
          </div>
          <Switch
            id="use-otf"
            checked={subtitleConfig.useOTF}
            onCheckedChange={(checked) =>
              setSubtitleConfig((prev) => ({
                ...prev,
                useOTF: checked,
              }))
            }
            disabled={loading}
          />
        </div>

        {/* 使用相似字体 */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Label htmlFor="use-similar-font">使用相似字体</Label>
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
                    当找不到指定字体时，使用相似的字体替代（同一字体族内，但字体字重相似的字体）
                  </p>
                </HybridTooltipContent>
              </HybridTooltip>
            </TooltipProvider>
          </div>
          <Switch
            id="use-similar-font"
            checked={subtitleConfig.useSimilarFont}
            onCheckedChange={(checked) =>
              setSubtitleConfig((prev) => ({
                ...prev,
                useSimilarFont: checked,
              }))
            }
            disabled={loading}
          />
        </div>

        {/* 使用系统字体目录 */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Label htmlFor="use-system-fonts-dir">
              初始化字体库使用系统字体目录
            </Label>
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
                  <p>从系统字体目录中查找字体文件</p>
                </HybridTooltipContent>
              </HybridTooltip>
            </TooltipProvider>
          </div>
          <Switch
            id="use-system-fonts-dir"
            checked={subtitleConfig.useSystemFontsDir}
            onCheckedChange={(checked) =>
              setSubtitleConfig((prev) => ({
                ...prev,
                useSystemFontsDir: checked,
              }))
            }
            disabled={loading}
          />
        </div>

        {/* 覆盖已存在的子集字体 */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Label htmlFor="cover-exist-sub-font">覆盖已子集化的字幕</Label>
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
                  <p>如果字幕本身已经子集化，是否覆盖重新生成</p>
                </HybridTooltipContent>
              </HybridTooltip>
            </TooltipProvider>
          </div>
          <Switch
            id="cover-exist-sub-font"
            checked={subtitleConfig.coverExistSubFont}
            onCheckedChange={(checked) =>
              setSubtitleConfig((prev) => ({
                ...prev,
                coverExistSubFont: checked,
              }))
            }
            disabled={loading}
          />
        </div>

        {/* 生成新文件 */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Label htmlFor="generate-new-file">生成新文件</Label>
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
                    生成新的字幕文件而不是修改原文件，
                    <br />
                    开启后会生成xxx.subset.xxx(如zh-cn).ass的的文件
                  </p>
                </HybridTooltipContent>
              </HybridTooltip>
            </TooltipProvider>
          </div>
          <Switch
            id="generate-new-file"
            checked={subtitleConfig.generateNewFile}
            onCheckedChange={(checked) =>
              setSubtitleConfig((prev) => ({
                ...prev,
                generateNewFile: checked,
              }))
            }
            disabled={loading}
          />
        </div>

        {/* 检查字形 */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Label htmlFor="check-glyphs">检查字形</Label>
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
                    检查字体中是否包含所需的字形
                    <br />
                    开启后如果字体库缺少对应字形则会报错并停止子集化
                  </p>
                </HybridTooltipContent>
              </HybridTooltip>
            </TooltipProvider>
          </div>
          <Switch
            id="check-glyphs"
            checked={subtitleConfig.checkGlyphs}
            onCheckedChange={(checked) =>
              setSubtitleConfig((prev) => ({
                ...prev,
                checkGlyphs: checked,
              }))
            }
            disabled={loading}
          />
        </div>
      </div>

      {/* 字体库管理 - 操作面板式 */}
      <div className="space-y-4">
        <h4 className="text-md font-medium text-muted-foreground">
          字体库管理
        </h4>

        {/* 状态面板 */}
        <div className="rounded-lg border bg-card p-4">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2">
              <Database className="h-5 w-5 text-primary" />
              <span className="font-medium">字体库状态</span>
            </div>
            <div className="flex items-center gap-2">
              <Button
                onClick={loadFontStats}
                disabled={statsLoading}
                variant="ghost"
                size="sm"
                className="h-8 w-8 p-0"
              >
                <RefreshCw
                  className={`h-4 w-4 ${statsLoading ? "animate-spin" : ""}`}
                />
              </Button>
              <Button
                onClick={initFontMetaSet}
                disabled={initLoading}
                variant="outline"
                size="sm"
                className="rounded-lg"
              >
                {initLoading ? (
                  <>
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                    初始化中
                  </>
                ) : (
                  <>
                    <Database className="h-4 w-4 mr-2" />
                    初始化
                  </>
                )}
              </Button>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            {/* 字体总数 */}
            <div className="space-y-1">
              <div className="flex items-center gap-2">
                <span className="text-sm text-muted-foreground">字体总数</span>
              </div>
              <div className="text-2xl font-bold text-primary">
                {statsLoading ? (
                  <Loader2 className="h-6 w-6 animate-spin" />
                ) : (
                  fontStats.total.toLocaleString()
                )}
              </div>
            </div>

            {/* 初始化状态 */}
            <div className="space-y-1">
              <div className="flex items-center gap-2">
                <span className="text-sm text-muted-foreground">
                  初始化状态
                </span>
              </div>
              <div className="flex items-center gap-2">
                {statsLoading ? (
                  <Loader2 className="h-6 w-6 animate-spin" />
                ) : fontStats.initDone ? (
                  <>
                    <CheckCircle className="h-6 w-6 text-green-500" />
                    <span className="text-sm font-medium text-green-600">
                      已完成
                    </span>
                  </>
                ) : (
                  <>
                    <AlertCircle className="h-6 w-6 text-orange-500" />
                    <span className="text-sm font-medium text-orange-600">
                      未初始化
                    </span>
                  </>
                )}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
