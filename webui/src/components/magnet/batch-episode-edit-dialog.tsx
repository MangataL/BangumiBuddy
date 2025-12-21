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
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { FileVideo, Calendar, Hash, ArrowRight, Loader2 } from "lucide-react";
import { TorrentFile } from "@/api/magnet";
import magnetAPI from "@/api/magnet";
import { ScrollArea } from "@/components/ui/scroll-area";
import { cn } from "@/lib/utils";

interface BatchEpisodeEditDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  taskID: string;
  triggerFile: TorrentFile | null;
  allFiles: TorrentFile[];
  onSave: (
    updates: { fileName: string; updates: Partial<TorrentFile> }[]
  ) => void;
}

export function BatchEpisodeEditDialog({
  open,
  onOpenChange,
  taskID,
  triggerFile,
  allFiles,
  onSave,
}: BatchEpisodeEditDialogProps) {
  const [season, setSeason] = useState<number>(0);
  const [offset, setOffset] = useState<number>(0);
  const [loading, setLoading] = useState(false);
  const [similarFileNames, setSimilarFileNames] = useState<string[]>([]);

  // 获取相似文件
  useEffect(() => {
    if (open && triggerFile && taskID) {
      setLoading(true);
      setSeason(triggerFile.season || 1);
      setOffset(0);
      magnetAPI
        .findSimilarFiles(taskID, triggerFile.fileName)
        .then((files) => {
          setSimilarFileNames(files);
        })
        .finally(() => {
          setLoading(false);
        });
    }
  }, [open, triggerFile, taskID]);

  // 匹配出的文件对象
  const matchedFiles = useMemo(() => {
    return allFiles.filter((f) => similarFileNames.includes(f.fileName));
  }, [allFiles, similarFileNames]);

  // 格式化季集号码为两位数
  const formatEpisodeNumber = (num: number): string => {
    return (num || 0).toString().padStart(2, "0");
  };

  // 处理保存
  const handleSave = () => {
    const updates = matchedFiles.map((f) => ({
      fileName: f.fileName,
      updates: {
        season,
        episode: Math.max(0, (f.episode || 0) + offset),
      },
    }));
    onSave(updates);
    onOpenChange(false);
  };

  if (!triggerFile) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="flex flex-col rounded-xl border-primary/20 bg-card/95 backdrop-blur-md overflow-hidden w-[92vw] max-w-[92vw] h-[80vh] max-h-[80vh] p-4 sm:max-w-2xl sm:h-auto sm:max-h-[90vh] sm:p-6">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <FileVideo className="w-5 h-5 text-blue-600" />
            批量编辑相似文件
          </DialogTitle>
          <DialogDescription>
            系统已自动识别出任务中相似的视频文件，你可以统一设置它们的季度并应用集数偏移。
          </DialogDescription>
        </DialogHeader>

        <div className="flex-1 overflow-hidden py-4 space-y-6 flex flex-col min-h-0">
          {/* 设置区域 */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 p-4 rounded-xl bg-primary/5 border border-primary/10 flex-shrink-0">
            <div className="space-y-2">
              <Label className="text-sm font-medium flex items-center gap-2 text-blue-600">
                <Calendar className="w-4 h-4" />
                统一设置季度
              </Label>
              <div className="flex items-center gap-2">
                <span className="text-xs text-muted-foreground">第</span>
                <Input
                  type="number"
                  min="0"
                  value={season}
                  onChange={(e) => setSeason(parseInt(e.target.value) || 0)}
                  className="h-9 w-20"
                />
                <span className="text-xs text-muted-foreground">季</span>
              </div>
            </div>

            <div className="space-y-2">
              <Label className="text-sm font-medium flex items-center gap-2 text-purple-600">
                <Hash className="w-4 h-4" />
                集数偏移 (Offset)
              </Label>
              <div className="flex items-center gap-2">
                <Input
                  type="number"
                  value={offset}
                  onChange={(e) => setOffset(parseInt(e.target.value) || 0)}
                  className="h-9 w-24"
                  placeholder="+n 或 -n"
                />
                <div className="flex flex-col">
                  <span className="text-[10px] text-muted-foreground leading-none">
                    在原识别集数上加减
                  </span>
                </div>
              </div>
            </div>
          </div>

          {/* 预览区域 */}
          <div className="space-y-2 flex flex-col min-h-0 flex-1 overflow-hidden">
            <Label className="text-sm font-medium text-muted-foreground px-1 flex items-center justify-between">
              <span>识别结果预览 ({matchedFiles.length} 个文件)</span>
              {loading && <Loader2 className="w-3 h-3 animate-spin" />}
            </Label>

            <ScrollArea className="flex-1 border rounded-lg bg-card overflow-y-auto">
              <div className="p-4 space-y-4">
                {matchedFiles.map((file, idx) => {
                  const newEp = Math.max(0, (file.episode || 0) + offset);
                  return (
                    <div
                      key={idx}
                      className="flex flex-col gap-1.5 pb-4 border-b last:border-0 last:pb-0"
                    >
                      <div
                        className="text-[11px] text-muted-foreground break-all leading-normal"
                        title={file.fileName}
                      >
                        {file.fileName.split("/").pop()}
                      </div>
                      <div className="flex items-center gap-2">
                        <Badge
                          variant="secondary"
                          className="text-[10px] font-normal opacity-70 px-1.5 py-0"
                        >
                          S{formatEpisodeNumber(file.season)}E
                          {formatEpisodeNumber(file.episode)}
                        </Badge>
                        <ArrowRight className="w-3 h-3 text-muted-foreground" />
                        <Badge className="bg-green-500/10 text-green-600 border-green-500/20 text-[10px] font-bold px-1.5 py-0">
                          S{formatEpisodeNumber(season)}E
                          {formatEpisodeNumber(newEp)}
                        </Badge>
                      </div>
                    </div>
                  );
                })}
                {matchedFiles.length === 0 && !loading && (
                  <div className="py-8 text-center text-sm text-muted-foreground">
                    未找到相似文件
                  </div>
                )}
              </div>
            </ScrollArea>
          </div>
        </div>

        <DialogFooter className="pt-2 border-t">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            取消
          </Button>
          <Button
            onClick={handleSave}
            disabled={matchedFiles.length === 0 || loading}
            className="bg-blue-600 hover:bg-blue-700"
          >
            确认批量修改
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
