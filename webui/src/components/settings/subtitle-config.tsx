import { useState, useEffect } from "react";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from "@/components/ui/pagination";
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
  Type,
  Eye,
} from "lucide-react";
import { useToast } from "@/hooks/useToast";
import {
  configAPI,
  type SubtitleOperatorConfig,
  type FontMetaSetStats,
  type Font,
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

  // 字体列表分页状态
  const [fonts, setFonts] = useState<Font[]>([]);
  const [fontsLoading, setFontsLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize] = useState(200);
  const [showFontsDialog, setShowFontsDialog] = useState(false);

  // 获取字体列表
  const loadFonts = async (currentPage: number) => {
    try {
      setFontsLoading(true);
      const data = await configAPI.listFonts(currentPage, pageSize);
      setFonts(data);
    } catch (error) {
      const description = extractErrorMessage(error);
      toast({
        title: "获取字体列表失败",
        description,
        variant: "destructive",
      });
    } finally {
      setFontsLoading(false);
    }
  };

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

  // 监听页码和弹窗状态变化
  useEffect(() => {
    if (showFontsDialog && fontStats.initDone) {
      loadFonts(page);
    }
  }, [page, pageSize, showFontsDialog]);

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

  const renderSubName = (
    chineseName: string,
    englishName: string,
    opacity: string = "80"
  ) => {
    if (!chineseName || chineseName === englishName) return null;
    return (
      <span
        className={`text-[10px] text-muted-foreground/${opacity} leading-tight`}
      >
        {chineseName}
      </span>
    );
  };

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

              {fontStats.initDone && fontStats.total > 0 && (
                <Dialog
                  open={showFontsDialog}
                  onOpenChange={setShowFontsDialog}
                >
                  <DialogTrigger asChild>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-8 w-8 p-0"
                      title="查看列表"
                    >
                      <Eye className="h-4 w-4" />
                    </Button>
                  </DialogTrigger>
                  <DialogContent className="max-w-5xl w-[95dvw] max-h-[85vh] flex flex-col rounded-xl border-primary/20 bg-card/95 backdrop-blur-md p-4 sm:p-6 sm:w-full">
                    <DialogHeader className="px-1">
                      <DialogTitle className="flex items-center gap-2 text-xl">
                        <Type className="h-6 w-6 text-primary" />
                        已加载字体库详情
                        <span className="text-sm font-normal text-muted-foreground ml-2">
                          共 {fontStats.total} 个字体条目
                        </span>
                      </DialogTitle>
                    </DialogHeader>

                    <div className="flex-1 overflow-auto mt-4 border rounded-xl bg-background/50">
                      <Table className="relative min-w-[900px]">
                        <TableHeader className="bg-muted/50 sticky top-0 z-10">
                          <TableRow>
                            <TableHead className="w-[25%]">字体名</TableHead>
                            <TableHead className="w-[25%]">家族名</TableHead>
                            <TableHead className="w-[20%]">PS 名称</TableHead>
                            <TableHead className="w-[20%]">字体文件</TableHead>
                            <TableHead className="w-[10%] text-center">
                              字体属性
                            </TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {fontsLoading ? (
                            <TableRow>
                              <TableCell
                                colSpan={5}
                                className="h-64 text-center"
                              >
                                <div className="flex flex-col items-center gap-2">
                                  <Loader2 className="h-8 w-8 animate-spin text-primary" />
                                  <span className="text-sm text-muted-foreground">
                                    正在调取字体元数据...
                                  </span>
                                </div>
                              </TableCell>
                            </TableRow>
                          ) : fonts.length === 0 ? (
                            <TableRow>
                              <TableCell
                                colSpan={5}
                                className="h-32 text-center text-muted-foreground"
                              >
                                未发现已加载的字体元数据
                              </TableCell>
                            </TableRow>
                          ) : (
                            fonts.map((font, index) => (
                              <TableRow
                                key={`${font.fullName}-${index}`}
                                className="hover:bg-muted/30 transition-colors"
                              >
                                <TableCell>
                                  <div className="flex flex-col gap-0.5">
                                    <span className="font-semibold text-sm">
                                      {font.fullName}
                                    </span>
                                    {renderSubName(
                                      font.chineseFullName,
                                      font.fullName
                                    )}
                                  </div>
                                </TableCell>
                                <TableCell>
                                  <div className="flex flex-col gap-0.5">
                                    <span className="text-xs font-medium">
                                      {font.familyName}
                                    </span>
                                    {renderSubName(
                                      font.chineseFamilyName,
                                      font.familyName
                                    )}
                                  </div>
                                </TableCell>
                                <TableCell>
                                  <div className="flex flex-col gap-0.5">
                                    <span className="text-[11px] font-mono text-muted-foreground">
                                      {font.postScriptName}
                                    </span>
                                    {renderSubName(
                                      font.chinesePostScriptName,
                                      font.postScriptName,
                                      "60"
                                    )}
                                  </div>
                                </TableCell>
                                <TableCell>
                                  <div
                                    className="font-mono text-[10px] text-muted-foreground/70 hover:text-foreground transition-colors break-all"
                                    title={font.fontFileName}
                                  >
                                    {font.fontFileName}
                                  </div>
                                </TableCell>
                                <TableCell>
                                  <div className="flex items-center justify-center gap-1">
                                    {/* 细体形态 */}
                                    {font.boldWeight < 400 && (
                                      <Badge
                                        variant="outline"
                                        className="h-5 px-1.5 text-[10px] font-light border-orange-500/20 text-orange-500/70 bg-orange-500/5"
                                      >
                                        L
                                      </Badge>
                                    )}

                                    {/* 常规形态 */}
                                    {font.boldWeight >= 400 &&
                                      font.boldWeight < 700 && (
                                        <Badge
                                          variant="outline"
                                          className="h-5 px-1.5 text-[10px] font-normal border-primary/20 text-primary/70"
                                        >
                                          R
                                        </Badge>
                                      )}

                                    {/* 粗体形态 */}
                                    {font.boldWeight >= 700 && (
                                      <Badge className="h-5 px-1.5 text-[10px] font-bold bg-primary text-primary-foreground">
                                        B
                                      </Badge>
                                    )}

                                    {/* 斜体状态 */}
                                    {font.italic && (
                                      <Badge
                                        variant="secondary"
                                        className="h-5 px-1.5 text-[10px] font-bold italic font-serif"
                                      >
                                        I
                                      </Badge>
                                    )}
                                  </div>
                                </TableCell>
                              </TableRow>
                            ))
                          )}
                        </TableBody>
                      </Table>
                    </div>

                    {/* 分页控制 */}
                    {fontStats.total > pageSize && (
                      <div className="mt-4">
                        <Pagination>
                          <PaginationContent>
                            <PaginationItem>
                              <PaginationPrevious
                                onClick={(e) => {
                                  e.preventDefault();
                                  setPage((p) => Math.max(1, p - 1));
                                }}
                                className={
                                  page === 1
                                    ? "pointer-events-none opacity-50"
                                    : "cursor-pointer"
                                }
                              />
                            </PaginationItem>

                            {/* 简单的分页逻辑 */}
                            {(() => {
                              const totalPages = Math.ceil(
                                fontStats.total / pageSize
                              );
                              const pages = [];
                              // 移动端展示更少页码
                              const maxVisible = 3;
                              let start = Math.max(
                                1,
                                page - Math.floor(maxVisible / 2)
                              );
                              let end = Math.min(
                                totalPages,
                                start + maxVisible - 1
                              );

                              if (end - start < maxVisible - 1) {
                                start = Math.max(1, end - maxVisible + 1);
                              }

                              if (start > 1) {
                                pages.push(
                                  <PaginationItem
                                    key={1}
                                    className="hidden sm:inline-block"
                                  >
                                    <PaginationLink
                                      onClick={(e) => {
                                        e.preventDefault();
                                        setPage(1);
                                      }}
                                      isActive={page === 1}
                                      className="cursor-pointer"
                                    >
                                      1
                                    </PaginationLink>
                                  </PaginationItem>
                                );
                                if (start > 2)
                                  pages.push(
                                    <PaginationItem
                                      key="start-ellipsis"
                                      className="hidden sm:inline-block"
                                    >
                                      <PaginationEllipsis />
                                    </PaginationItem>
                                  );
                              }

                              for (let i = start; i <= end; i++) {
                                pages.push(
                                  <PaginationItem key={i}>
                                    <PaginationLink
                                      onClick={(e) => {
                                        e.preventDefault();
                                        setPage(i);
                                      }}
                                      isActive={page === i}
                                      className="cursor-pointer"
                                    >
                                      {i}
                                    </PaginationLink>
                                  </PaginationItem>
                                );
                              }

                              if (end < totalPages) {
                                if (end < totalPages - 1)
                                  pages.push(
                                    <PaginationItem
                                      key="end-ellipsis"
                                      className="hidden sm:inline-block"
                                    >
                                      <PaginationEllipsis />
                                    </PaginationItem>
                                  );
                                pages.push(
                                  <PaginationItem
                                    key={totalPages}
                                    className="hidden sm:inline-block"
                                  >
                                    <PaginationLink
                                      onClick={(e) => {
                                        e.preventDefault();
                                        setPage(totalPages);
                                      }}
                                      isActive={page === totalPages}
                                      className="cursor-pointer"
                                    >
                                      {totalPages}
                                    </PaginationLink>
                                  </PaginationItem>
                                );
                              }
                              return pages;
                            })()}

                            <PaginationItem>
                              <PaginationNext
                                onClick={(e) => {
                                  e.preventDefault();
                                  setPage((p) =>
                                    Math.min(
                                      Math.ceil(fontStats.total / pageSize),
                                      p + 1
                                    )
                                  );
                                }}
                                className={
                                  page === Math.ceil(fontStats.total / pageSize)
                                    ? "pointer-events-none opacity-50"
                                    : "cursor-pointer"
                                }
                              />
                            </PaginationItem>
                          </PaginationContent>
                        </Pagination>
                      </div>
                    )}
                  </DialogContent>
                </Dialog>
              )}

              <Button
                onClick={initFontMetaSet}
                disabled={initLoading}
                variant="outline"
                size="sm"
                className="rounded-lg h-8 w-8 p-0 sm:w-auto sm:px-3"
                title="初始化"
              >
                {initLoading ? (
                  <>
                    <Loader2 className="h-4 w-4 sm:mr-2 animate-spin" />
                    <span className="hidden sm:inline">初始化中</span>
                  </>
                ) : (
                  <>
                    <Database className="h-4 w-4 sm:mr-2" />
                    <span className="hidden sm:inline">初始化</span>
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
              <div className="flex items-center gap-3">
                <div className="text-2xl font-bold text-primary">
                  {statsLoading ? (
                    <Loader2 className="h-6 w-6 animate-spin" />
                  ) : (
                    fontStats.total.toLocaleString()
                  )}
                </div>
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
