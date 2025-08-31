import { useState, useEffect } from "react";
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
import {
  Folder,
  FolderOpen,
  FileText,
  ArrowRight,
  Loader2,
} from "lucide-react";
import magnetAPI, { type TorrentFile, type DownloadType } from "@/api/magnet";
import { configAPI } from "@/api/config";
import { cn } from "@/lib/utils";
import { useToast } from "@/hooks/useToast";
import { useMobile } from "@/hooks/useMobile";
import { TorrentDirectoryTree } from "./torrent-directory-tree";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { SourceDirectorySelector } from "./source-directory-selector";

interface SubtitleTransferDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  taskID: string;
  torrentFiles: TorrentFile[];
  downloadType: DownloadType;
  onSuccess?: () => void;
}

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
  const [initialPath, setInitialPath] = useState<string>("/");
  const [transferring, setTransferring] = useState(false);
  const [selectedSourceDir, setSelectedSourceDir] = useState<string>("");
  const [selectedTargetDir, setSelectedTargetDir] = useState<string>("");
  const [initialPathLoaded, setInitialPathLoaded] = useState(false);

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
      setSelectedSourceDir("");
      setSelectedTargetDir("");
    } else if (!open) {
      // 对话框关闭时重置状态
      setInitialPathLoaded(false);
    }
  }, [open, downloadType]);

  // 处理转移
  const handleTransfer = async () => {
    if (!selectedSourceDir) {
      toast({
        title: "请选择源目录",
        description: "请先选择包含字幕文件的源目录",
        variant: "destructive",
      });
      return;
    }

    // selectedTargetDir 可以是空字符串（表示根目录），所以不验证

    setTransferring(true);
    try {
      await magnetAPI.addSubtitles(taskID, {
        subtitleDir: selectedSourceDir,
        dstDir: selectedTargetDir, // 空字符串代表根目录
      });
      toast({
        title: "转移成功",
        description: "字幕文件已成功转移到目标目录",
      });
      onSuccess?.();
      onOpenChange(false);
    } catch (error) {
      toast({
        title: "转移失败",
        description: error instanceof Error ? error.message : "未知错误",
        variant: "destructive",
      });
    } finally {
      setTransferring(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        className={cn(
          "flex flex-col rounded-xl",
          isMobile
            ? "w-[92vw] max-w-[92vw] h-[82vh] max-h-[82vh] p-4"
            : "max-w-5xl max-h-[85vh] p-6"
        )}
      >
        <DialogHeader>
          <DialogTitle
            className={cn(
              "flex items-center gap-2",
              isMobile ? "text-base" : "text-lg"
            )}
          >
            <FileText
              className={
                isMobile ? "w-4 h-4 text-blue-600" : "w-5 h-5 text-blue-600"
              }
            />
            转移字幕文件
          </DialogTitle>
          <DialogDescription className={isMobile ? "text-xs" : "text-sm"}>
            选择字幕文件的源目录和要转移到的目标目录
          </DialogDescription>
        </DialogHeader>

        <div
          className={cn("flex-1 overflow-hidden", isMobile ? "py-3" : "py-4")}
        >
          {isMobile ? (
            // 移动端：使用 Tabs 切换
            <Tabs defaultValue="source" className="h-full flex flex-col">
              <TabsList className="grid w-full grid-cols-2">
                <TabsTrigger value="source">
                  <FolderOpen className="w-3.5 h-3.5 mr-1.5" />
                  源目录
                </TabsTrigger>
                <TabsTrigger value="target">
                  <Folder className="w-3.5 h-3.5 mr-1.5" />
                  目标目录
                </TabsTrigger>
              </TabsList>

              <TabsContent
                value="source"
                className="flex-1 overflow-hidden mt-3"
              >
                <SourceDirectorySelector
                  initialPath={initialPath}
                  selectedDir={selectedSourceDir}
                  onSelectDir={setSelectedSourceDir}
                  isMobile={true}
                />
              </TabsContent>

              <TabsContent
                value="target"
                className="flex-1 overflow-hidden mt-3"
              >
                <div className="space-y-3 h-full flex flex-col">
                  <Label className="text-sm font-medium">
                    选择目标目录（从当前任务文件树中选择）
                  </Label>

                  {/* 已选择的目标目录 */}
                  {selectedTargetDir !== undefined && (
                    <div className="p-2 rounded-lg bg-green-500/5 border border-green-500/20">
                      <div className="flex items-center gap-2">
                        <Folder className="w-3.5 h-3.5 text-green-600 flex-shrink-0" />
                        <code className="text-xs break-all text-green-700 dark:text-green-300">
                          {selectedTargetDir || "根目录"}
                        </code>
                      </div>
                    </div>
                  )}

                  <TorrentDirectoryTree
                    files={torrentFiles}
                    selectedPath={selectedTargetDir}
                    onSelect={setSelectedTargetDir}
                  />
                </div>
              </TabsContent>
            </Tabs>
          ) : (
            // 桌面端：左右分栏
            <div className="grid grid-cols-2 gap-4 h-full">
              {/* 左侧：源目录 */}
              <div className="space-y-3 flex flex-col h-full overflow-hidden">
                <Label className="text-sm font-medium flex items-center gap-2">
                  <FolderOpen className="w-4 h-4 text-blue-600" />
                  源目录（字幕所在位置）
                </Label>

                <SourceDirectorySelector
                  initialPath={initialPath}
                  selectedDir={selectedSourceDir}
                  onSelectDir={setSelectedSourceDir}
                  isMobile={false}
                />
              </div>

              {/* 右侧：目标目录 */}
              <div className="space-y-3 flex flex-col h-full overflow-hidden">
                <Label className="text-sm font-medium flex items-center gap-2">
                  <Folder className="w-4 h-4 text-green-600" />
                  目标目录（转移到）
                </Label>

                {/* 已选择的目标目录 */}
                {selectedTargetDir !== undefined && (
                  <div className="p-3 rounded-lg bg-green-500/5 border border-green-500/20">
                    <div className="flex items-center gap-2">
                      <Folder className="w-4 h-4 text-green-600 flex-shrink-0" />
                      <code className="text-xs break-all text-green-700 dark:text-green-300">
                        {selectedTargetDir || "根目录"}
                      </code>
                    </div>
                  </div>
                )}

                <TorrentDirectoryTree
                  files={torrentFiles}
                  selectedPath={selectedTargetDir}
                  onSelect={setSelectedTargetDir}
                />

                {/* 操作提示 */}
                <div className="flex items-start gap-2 p-2 rounded-lg bg-muted/50 text-xs text-muted-foreground">
                  <div className="flex-shrink-0 mt-0.5">💡</div>
                  <p>从当前任务的文件树中选择目标目录</p>
                </div>
              </div>
            </div>
          )}
        </div>

        <DialogFooter className={cn("gap-2", isMobile && "flex-col-reverse")}>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            className={isMobile ? "w-full" : ""}
          >
            取消
          </Button>
          <Button
            onClick={handleTransfer}
            disabled={!selectedSourceDir || transferring}
            className={cn(
              "bg-blue-600 hover:bg-blue-700",
              isMobile && "w-full"
            )}
          >
            {transferring ? (
              <>
                <Loader2
                  className={
                    isMobile
                      ? "w-3.5 h-3.5 mr-1.5 animate-spin"
                      : "w-4 h-4 mr-2 animate-spin"
                  }
                />
                转移中...
              </>
            ) : (
              <>
                <ArrowRight
                  className={isMobile ? "w-3.5 h-3.5 mr-1.5" : "w-4 h-4 mr-2"}
                />
                开始转移
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
