import { useState, useEffect, useMemo } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import {
  FolderOpen,
  FileText,
  ArrowRight,
  Loader2,
  MapPin,
  Info,
  ChevronDown,
  ChevronLeft,
  CheckCircle2,
  AlertCircle,
  Video,
  Settings,
  Subtitles,
} from "lucide-react";
import magnetAPI, {
  type TorrentFile,
  type DownloadType,
  type AddSubtitlesResult,
} from "@/api/magnet";
import { configAPI } from "@/api/config";
import { cn } from "@/lib/utils";
import { useToast } from "@/hooks/useToast";
import { useMobile } from "@/hooks/useMobile";
import { TorrentDirectoryTree } from "./torrent-directory-tree";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  SourceDirectorySelector,
  type SourceSelection,
} from "./source-directory-selector";
import { extractErrorMessage } from "@/utils/error";
import { EpisodePositionInput } from "@/components/subscription/episode-position-input";
import {
  HybridTooltip,
  HybridTooltipContent,
  HybridTooltipTrigger,
} from "@/components/common/tooltip";
import { TooltipProvider } from "@/components/ui/tooltip";
import { ScrollArea } from "@/components/ui/scroll-area";

interface SubtitleTransferDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  taskID: string;
  torrentFiles: TorrentFile[];
  downloadType: DownloadType;
  onSuccess?: () => void;
}

type ViewState = "selection" | "preview";

export function SubtitleTransferDialog({
  open,
  onOpenChange,
  taskID,
  torrentFiles,
  downloadType,
  onSuccess,
}: SubtitleTransferDialogProps) {
  const { toast } = useToast();
  const isMobile = useMobile();
  const [view, setView] = useState<ViewState>("selection");
  const [initialPath, setInitialPath] = useState<string>("/");
  const [transferring, setTransferring] = useState(false);
  const [loadingPreview, setLoadingPreview] = useState(false);

  // 选择状态
  const [selectedSource, setSelectedSource] = useState<SourceSelection | null>(
    null
  );
  const [selectedTargetDir, setSelectedTargetDir] = useState<string | null>(
    null
  );
  const [preserveOriginal, setPreserveOriginal] = useState(true);

  // 预览状态
  const [previewItems, setPreviewItems] = useState<
    Record<string, AddSubtitlesResult>
  >({});
  const [confirmedFiles, setConfirmedFiles] = useState<Record<string, boolean>>(
    {}
  );

  const [initialPathLoaded, setInitialPathLoaded] = useState(false);
  const [episodeLocation, setEpisodeLocation] = useState<string>("");
  const [episodeOffset, setEpisodeOffset] = useState<string>("");
  const [season, setSeason] = useState<string>("");
  const [extensionLevel, setExtensionLevel] = useState<string>("");
  const [episodeToolsExpanded, setEpisodeToolsExpanded] = useState(!isMobile);

  const selectablePreviewEntries = useMemo(
    () => Object.entries(previewItems).filter(([, res]) => !res.error),
    [previewItems]
  );
  const selectablePreviewCount = selectablePreviewEntries.length;

  // 加载默认路径
  useEffect(() => {
    if (open && !initialPathLoaded) {
      const loadDefaultPath = async () => {
        try {
          const config = await configAPI.getDownloadManagerConfig();
          const defaultPath =
            downloadType === "tv" ? config.tvSavePath : config.movieSavePath;

          if (defaultPath) {
            setInitialPath(defaultPath);
          } else {
            setInitialPath("/");
          }
        } catch (error) {
          console.error("Failed to load default path:", error);
          setInitialPath("/");
        } finally {
          setInitialPathLoaded(true);
        }
      };

      loadDefaultPath();
      setView("selection");
      setSelectedSource(null);
      setSelectedTargetDir(null);
      setPreviewItems({});
      setConfirmedFiles({});
      setEpisodeLocation("");
      setEpisodeOffset("");
      setSeason("");
      setExtensionLevel("");
      setEpisodeToolsExpanded(!isMobile);
    } else if (!open) {
      setInitialPathLoaded(false);
    }
  }, [open, downloadType]);

  // 生成预览
  const handleGeneratePreview = async () => {
    const sourcePath = selectedSource?.path;

    if (!sourcePath) {
      toast({
        title: "请选择源",
        description: "请选择目录或具体字幕文件",
        variant: "destructive",
      });
      return;
    }

    if (selectedTargetDir === null) {
      toast({
        title: "请选择目标",
        description: "请选择转移的目标位置",
        variant: "destructive",
      });
      return;
    }

    setLoadingPreview(true);
    try {
      const offsetValue =
        episodeOffset.trim() === "" ? undefined : Number(episodeOffset);
      const seasonValue = season.trim() === "" ? undefined : Number(season);
      const extensionLevelValue =
        extensionLevel.trim() === "" ? undefined : Number(extensionLevel);

      const resp = await magnetAPI.previewAddSubtitles(taskID, {
        subtitlePath: sourcePath, // 如果选了多个文件，预览接口目前可能需要调整，这里暂传第一个或目录
        dstPath: selectedTargetDir,
        episodeLocation: episodeLocation || undefined,
        episodeOffset: offsetValue,
        season: seasonValue,
        extensionLevel: extensionLevelValue,
      });

      // 如果选了具体文件，预览结果应该只包含该文件
      let filteredResults = resp.subtitleFiles;
      if (selectedSource?.type === "file") {
        filteredResults = Object.fromEntries(
          Object.entries(resp.subtitleFiles).filter(
            ([path]) => path === selectedSource.path
          )
        );
      }

      setPreviewItems(filteredResults);
      // 默认勾选没有错误的项目
      const initialConfirmed: Record<string, boolean> = {};
      Object.entries(filteredResults).forEach(([path, res]) => {
        initialConfirmed[path] = !res.error;
      });
      setConfirmedFiles(initialConfirmed);
      setView("preview");
    } catch (error) {
      toast({
        title: "生成预览失败",
        description: extractErrorMessage(error),
        variant: "destructive",
      });
    } finally {
      setLoadingPreview(false);
    }
  };

  // 执行转移
  const handleTransfer = async () => {
    const filesToTransfer: Record<string, string> = {};
    Object.entries(previewItems).forEach(([path, res]) => {
      if (confirmedFiles[path] && res.targetPath) {
        filesToTransfer[path] = res.targetPath;
      }
    });

    if (Object.keys(filesToTransfer).length === 0) {
      toast({
        title: "未选择文件",
        description: "请至少勾选一个有效的转移项",
        variant: "destructive",
      });
      return;
    }

    setTransferring(true);
    try {
      const resp = await magnetAPI.addSubtitles(taskID, {
        subtitleFiles: filesToTransfer,
        preserveOriginal: preserveOriginal,
      });
      toast({
        title: "转移完成",
        description: `成功转移 ${resp.successCount} 个文件`,
      });
      onSuccess?.();
      onOpenChange(false);
    } catch (error) {
      toast({
        title: "转移失败",
        description: extractErrorMessage(error),
        variant: "destructive",
      });
    } finally {
      setTransferring(false);
    }
  };

  const renderSelection = () => (
    <div className="flex-1 min-h-0 overflow-hidden py-4">
      {isMobile ? (
        <Tabs defaultValue="source" className="h-full flex flex-col min-h-0">
          <TabsList className="grid w-full grid-cols-2 rounded-xl">
            <TabsTrigger value="source" className="rounded-lg">
              源字幕
            </TabsTrigger>
            <TabsTrigger value="target" className="rounded-lg">
              目标位置
            </TabsTrigger>
          </TabsList>
          <TabsContent
            value="source"
            className="flex-1 min-h-0 mt-4 flex-col data-[state=active]:flex data-[state=inactive]:hidden"
          >
            <SourceDirectorySelector
              initialPath={initialPath}
              selectedSource={selectedSource}
              onSelectSource={setSelectedSource}
              isMobile={true}
            />
          </TabsContent>
          <TabsContent
            value="target"
            className="flex-1 min-h-0 mt-4 overflow-hidden flex flex-col gap-3 data-[state=active]:flex data-[state=inactive]:hidden"
          >
            <Label className="text-sm font-medium">
              选择目标（目录或媒体文件）
            </Label>
            <TorrentDirectoryTree
              files={torrentFiles}
              selectedPath={selectedTargetDir || ""}
              onSelect={setSelectedTargetDir}
              className="flex-1"
            />
            {selectedTargetDir !== null && (
              <div className="p-3 rounded-xl bg-purple-500/5 border border-purple-500/10">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2 min-w-0">
                    {torrentFiles.find((f) => f.fileName === selectedTargetDir)
                      ?.media ? (
                      <Video className="w-4 h-4 text-purple-600 flex-shrink-0" />
                    ) : (
                      <FolderOpen className="w-4 h-4 text-purple-600 flex-shrink-0" />
                    )}
                    <span className="text-xs font-medium text-purple-700">
                      已选目标: {selectedTargetDir || "根目录"}
                    </span>
                  </div>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setSelectedTargetDir(null)}
                    className="h-6 px-2 text-[10px] hover:bg-purple-500/10 text-purple-600"
                  >
                    重置
                  </Button>
                </div>
              </div>
            )}
          </TabsContent>
        </Tabs>
      ) : (
        <div className="grid grid-cols-2 gap-6 h-full min-h-0">
          <div className="flex flex-col h-full min-h-0">
            <Label className="text-sm font-semibold mb-3 flex items-center gap-2">
              <FolderOpen className="w-4 h-4 text-blue-500" />
              1. 选择源字幕 (目录或具体文件)
            </Label>
            <SourceDirectorySelector
              initialPath={initialPath}
              selectedSource={selectedSource}
              onSelectSource={setSelectedSource}
              className="flex-1"
            />
          </div>
          <div className="flex flex-col h-full min-h-0">
            <Label className="text-sm font-semibold mb-3 flex items-center gap-2">
              <Video className="w-4 h-4 text-purple-500" />
              2. 选择目标位置 (目录或媒体文件)
            </Label>
            <TorrentDirectoryTree
              files={torrentFiles}
              selectedPath={selectedTargetDir || ""}
              onSelect={setSelectedTargetDir}
              className="flex-1"
            />
            {selectedTargetDir !== null && (
              <div className="mt-3 p-3 rounded-xl bg-purple-500/5 border border-purple-500/10">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2 min-w-0">
                    {torrentFiles.find((f) => f.fileName === selectedTargetDir)
                      ?.media ? (
                      <Video className="w-4 h-4 text-purple-600 flex-shrink-0" />
                    ) : (
                      <FolderOpen className="w-4 h-4 text-purple-600 flex-shrink-0" />
                    )}
                    <span className="text-xs font-medium text-purple-700">
                      已选目标: {selectedTargetDir || "根目录"}
                    </span>
                  </div>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setSelectedTargetDir(null)}
                    className="h-6 px-2 text-[10px] hover:bg-purple-500/10 text-purple-600"
                  >
                    重置
                  </Button>
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );

  const getFileStem = (name: string): string => {
    const lastDot = name.lastIndexOf(".");
    return lastDot > 0 ? name.slice(0, lastDot) : name;
  };

  const renderPreview = () => (
    <div className="flex-1 min-h-0 flex flex-col py-4 overflow-hidden">
      <div className="flex items-center justify-between mb-4 px-1">
        <div className="flex items-center gap-2">
          <div className="p-1.5 rounded-lg bg-green-500/10">
            <CheckCircle2 className="w-4 h-4 text-green-500" />
          </div>
          <h3 className="font-semibold text-base">转移预览与确认</h3>
        </div>
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2 px-2 py-1 rounded-lg hover:bg-muted/50 transition-colors">
            <Checkbox
              id="preserve-original"
              checked={preserveOriginal}
              onCheckedChange={(checked) => setPreserveOriginal(!!checked)}
            />
            <Label
              htmlFor="preserve-original"
              className="text-xs font-medium cursor-pointer select-none"
            >
              保留原字幕
            </Label>
          </div>
          <div className="flex items-center gap-2 px-2 py-1 rounded-lg hover:bg-muted/50 transition-colors">
            <Checkbox
              id="select-all"
              checked={
                selectablePreviewCount > 0 &&
                selectablePreviewEntries.every(([path]) => confirmedFiles[path])
              }
              onCheckedChange={(checked) => {
                const next: Record<string, boolean> = { ...confirmedFiles };
                selectablePreviewEntries.forEach(([path]) => {
                  next[path] = !!checked;
                });
                setConfirmedFiles(next);
              }}
            />
            <Label
              htmlFor="select-all"
              className="text-xs font-medium cursor-pointer select-none"
            >
              全选
            </Label>
          </div>
          <span className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground bg-muted/50 px-2 py-1 rounded-md">
            {selectablePreviewCount} 个项目
          </span>
        </div>
      </div>

      <ScrollArea className="flex-1 border rounded-2xl bg-muted/5">
        <div className="p-4 space-y-4">
          {Object.entries(previewItems).map(([path, res]) => (
            <div
              key={path}
              className={cn(
                "group relative flex items-start gap-4 p-4 rounded-xl border transition-all duration-200",
                res.error
                  ? "bg-red-500/5 border-red-500/20 shadow-sm"
                  : confirmedFiles[path]
                  ? "bg-white dark:bg-zinc-900 border-blue-500/30 shadow-sm"
                  : "bg-white dark:bg-zinc-900 border-zinc-200 dark:border-zinc-800 opacity-70 shadow-none"
              )}
            >
              <div className="flex flex-col items-center gap-3">
                <Checkbox
                  id={path}
                  checked={confirmedFiles[path]}
                  onCheckedChange={(checked) =>
                    setConfirmedFiles((prev) => ({
                      ...prev,
                      [path]: !!checked,
                    }))
                  }
                  disabled={!!res.error}
                  className="mt-1"
                />
              </div>

              <div className="flex-1 min-w-0 flex flex-col gap-4">
                <div className="grid grid-cols-1 lg:grid-cols-[1fr_auto_1fr] items-start gap-4">
                  {/* Source */}
                  <div className="space-y-1.5">
                    <div className="flex items-center gap-1.5 text-muted-foreground">
                      <Subtitles className="w-3.5 h-3.5" />
                      <span className="text-[10px] font-bold uppercase tracking-wider opacity-70">
                        源文件
                      </span>
                    </div>
                    <div
                      className="text-sm font-medium leading-relaxed break-all"
                      title={path}
                    >
                      {path.split("/").pop()}
                    </div>
                  </div>

                  {/* Desktop Arrow */}
                  <div className="hidden lg:flex flex-col items-center self-stretch py-2">
                    <div className="h-full w-[1px] bg-gradient-to-b from-transparent via-zinc-200 dark:via-zinc-800 to-transparent flex items-center justify-center">
                      <ArrowRight className="w-4 h-4 text-muted-foreground bg-white dark:bg-zinc-900 rounded-full p-0.5 border" />
                    </div>
                  </div>

                  {/* Mobile Divider */}
                  <div className="lg:hidden flex items-center gap-2 py-1">
                    <div className="h-[1px] flex-1 bg-zinc-100 dark:bg-zinc-800" />
                    <ArrowRight className="w-3.5 h-3.5 text-muted-foreground opacity-50" />
                    <div className="h-[1px] flex-1 bg-zinc-100 dark:bg-zinc-800" />
                  </div>

                  {/* Target */}
                  <div className="space-y-1.5">
                    <div
                      className={cn(
                        "flex items-center gap-1.5",
                        res.error ? "text-red-500" : "text-blue-600"
                      )}
                    >
                      <Subtitles className="w-3.5 h-3.5" />
                      <span className="text-[10px] font-bold uppercase tracking-wider opacity-70">
                        重命名后
                      </span>
                    </div>
                    <div
                      className={cn(
                        "text-sm font-bold leading-relaxed break-all",
                        res.error
                          ? "text-red-500"
                          : "text-zinc-700 dark:text-zinc-200"
                      )}
                      title={res.newFileName || "无法匹配"}
                    >
                      {!res.error && res.newFileName
                        ? (() => {
                            const mediaBase = res.mediaFileName
                              ? getFileStem(res.mediaFileName)
                              : "";
                            if (
                              mediaBase &&
                              res.newFileName.startsWith(mediaBase)
                            ) {
                              return (
                                <>
                                  <span className="text-blue-600 dark:text-blue-400">
                                    {mediaBase}
                                  </span>
                                  <span className="text-zinc-400 dark:text-zinc-500 font-normal">
                                    {res.newFileName.slice(mediaBase.length)}
                                  </span>
                                </>
                              );
                            }
                            return res.newFileName;
                          })()
                        : "无法匹配"}
                    </div>
                  </div>
                </div>

                {/* Additional Info */}
                {(res.targetPath || res.error) && (
                  <div className="pt-3 border-t border-zinc-100 dark:border-zinc-800/50 space-y-2">
                    {res.targetPath && (
                      <div className="flex items-start gap-2 text-[11px] text-muted-foreground bg-muted/30 p-2.5 rounded-lg border border-zinc-100 dark:border-zinc-800/50">
                        <MapPin className="w-3.5 h-3.5 mt-0.5 flex-shrink-0" />
                        <span className="break-all font-mono">
                          {res.targetPath}
                        </span>
                      </div>
                    )}
                    {res.error && (
                      <div className="flex items-start gap-2 text-[11px] text-red-500 font-semibold bg-red-500/10 p-2.5 rounded-lg border border-red-500/10">
                        <AlertCircle className="w-3.5 h-3.5 mt-0.5 flex-shrink-0" />
                        <span className="break-words">{res.error}</span>
                      </div>
                    )}
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      </ScrollArea>
    </div>
  );

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        className={cn(
          "flex flex-col rounded-2xl min-h-0",
          isMobile
            ? "w-[95vw] max-w-[95vw] h-[85vh] p-4"
            : "w-[80vw] max-w-5xl h-[80vh] p-6"
        )}
        onOpenAutoFocus={(event) => event.preventDefault()}
      >
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2 text-xl">
            <div className="p-2 rounded-xl bg-blue-500/10">
              <FileText className="w-5 h-5 text-blue-600" />
            </div>
            字幕转移助手
          </DialogTitle>
          <DialogDescription>
            {view === "selection"
              ? "配置源字幕路径与目标媒体路径"
              : "确认最终的转移映射关系"}
          </DialogDescription>
        </DialogHeader>

        {view === "selection" ? renderSelection() : renderPreview()}

        {/* 字幕识别修正工具栏 */}
        {view === "selection" && downloadType === "tv" && (
          <div className="py-2 border-t mt-auto">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setEpisodeToolsExpanded(!episodeToolsExpanded)}
              className="w-full justify-between h-9 text-xs text-blue-600 hover:bg-blue-500/5 rounded-xl transition-all"
            >
              <div className="flex items-center gap-2">
                <Settings className="w-3.5 h-3.5" />
                字幕识别修正
              </div>
              <ChevronDown
                className={cn(
                  "w-4 h-4 transition-transform duration-300",
                  episodeToolsExpanded && "rotate-180"
                )}
              />
            </Button>

            {episodeToolsExpanded && (
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4 pt-3 pb-1 animate-in fade-in slide-in-from-top-2 duration-300">
                <EpisodePositionInput
                  value={episodeLocation}
                  onChange={setEpisodeLocation}
                  className="space-y-1.5"
                  showTooltipButton={false}
                  labelContainerClassName="gap-1.5 ml-1"
                  labelClassName="text-[11px] font-medium text-muted-foreground"
                  inputClassName="h-9 rounded-xl text-sm bg-muted/20 border-none focus-visible:ring-blue-500/30"
                />

                <div className="space-y-1.5">
                  <div className="flex items-center gap-1.5 ml-1">
                    <Label className="text-[11px] font-medium text-muted-foreground">
                      集数偏移
                    </Label>
                    <TooltipProvider>
                      <HybridTooltip>
                        <HybridTooltipTrigger>
                          <Info className="h-3 w-3 text-muted-foreground/60 hover:text-blue-500 transition-colors" />
                        </HybridTooltipTrigger>
                        <HybridTooltipContent>
                          正数推迟，负数提前
                        </HybridTooltipContent>
                      </HybridTooltip>
                    </TooltipProvider>
                  </div>
                  <Input
                    type="number"
                    value={episodeOffset}
                    onChange={(e) => setEpisodeOffset(e.target.value)}
                    placeholder="0"
                    className="h-9 rounded-xl text-sm bg-muted/20 border-none focus-visible:ring-blue-500/30"
                  />
                </div>

                <div className="space-y-1.5">
                  <div className="flex items-center gap-1.5 ml-1">
                    <Label className="text-[11px] font-medium text-muted-foreground">
                      匹配季数
                    </Label>
                    <TooltipProvider>
                      <HybridTooltip>
                        <HybridTooltipTrigger>
                          <Info className="h-3 w-3 text-muted-foreground/60 hover:text-blue-500 transition-colors" />
                        </HybridTooltipTrigger>
                        <HybridTooltipContent>
                          指定字幕文件所属的季度信息，留空则自动识别
                        </HybridTooltipContent>
                      </HybridTooltip>
                    </TooltipProvider>
                  </div>
                  <Input
                    type="number"
                    value={season}
                    onChange={(e) => setSeason(e.target.value)}
                    placeholder="自动"
                    className="h-9 rounded-xl text-sm bg-muted/20 border-none focus-visible:ring-blue-500/30"
                  />
                </div>

                <div className="space-y-1.5">
                  <div className="flex items-center gap-1.5 ml-1">
                    <Label className="text-[11px] font-medium text-muted-foreground">
                      扩展层级
                    </Label>
                    <TooltipProvider>
                      <HybridTooltip>
                        <HybridTooltipTrigger>
                          <Info className="h-3 w-3 text-muted-foreground/60 hover:text-blue-500 transition-colors" />
                        </HybridTooltipTrigger>
                        <HybridTooltipContent>
                          字幕文件多级后缀保留层级 (如: 设置为2时，a.b.c.ass
                          -&gt; .c.ass)，默认全保留
                        </HybridTooltipContent>
                      </HybridTooltip>
                    </TooltipProvider>
                  </div>
                  <Input
                    type="number"
                    value={extensionLevel}
                    onChange={(e) => setExtensionLevel(e.target.value)}
                    placeholder="自动"
                    className="h-9 rounded-xl text-sm bg-muted/20 border-none focus-visible:ring-blue-500/30"
                  />
                </div>
              </div>
            )}
          </div>
        )}

        <DialogFooter className="gap-2 sm:gap-0">
          {view === "selection" ? (
            <>
              <Button
                variant="ghost"
                onClick={() => onOpenChange(false)}
                className="rounded-xl"
              >
                取消
              </Button>
              <Button
                onClick={handleGeneratePreview}
                disabled={
                  !selectedSource ||
                  selectedTargetDir === null ||
                  loadingPreview
                }
                className="bg-blue-600 hover:bg-blue-700 rounded-xl px-8"
              >
                {loadingPreview ? (
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                ) : (
                  <ArrowRight className="w-4 h-4 mr-2" />
                )}
                生成预览
              </Button>
            </>
          ) : (
            <>
              <Button
                variant="outline"
                onClick={() => setView("selection")}
                className="rounded-xl"
              >
                <ChevronLeft className="w-4 h-4 mr-2" />
                返回修改
              </Button>
              <Button
                onClick={handleTransfer}
                disabled={
                  transferring || Object.values(confirmedFiles).every((v) => !v)
                }
                className="bg-green-600 hover:bg-green-700 rounded-xl px-8"
              >
                {transferring ? (
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                ) : (
                  <CheckCircle2 className="w-4 h-4 mr-2" />
                )}
                确认转移
              </Button>
            </>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
