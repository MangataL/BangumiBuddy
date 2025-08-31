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
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { FileVideo, Calendar, Hash } from "lucide-react";
import { TorrentFile } from "@/api/magnet";

interface EpisodeEditDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  file: TorrentFile | null;
  onSave: (fileName: string, updates: Partial<TorrentFile>) => void;
}

export function EpisodeEditDialog({
  open,
  onOpenChange,
  file,
  onSave,
}: EpisodeEditDialogProps) {
  const [season, setSeason] = useState<number>(0);
  const [episode, setEpisode] = useState<number>(0);

  // 当文件变化时，更新本地状态
  useEffect(() => {
    if (file) {
      setSeason(file.season || 0);
      setEpisode(file.episode || 0);
    }
  }, [file]);

  // 格式化季集号码为两位数
  const formatEpisodeNumber = (num: number): string => {
    return num.toString().padStart(2, "0");
  };

  // 处理保存
  const handleSave = () => {
    if (file) {
      onSave(file.fileName, {
        season,
        episode,
      });
      onOpenChange(false);
    }
  };

  // 处理取消
  const handleCancel = () => {
    if (file) {
      setSeason(file.season || 0);
      setEpisode(file.episode || 0);
    }
    onOpenChange(false);
  };

  if (!file) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <FileVideo className="w-5 h-5 text-green-600" />
            编辑季集信息
          </DialogTitle>
          <DialogDescription>
            为文件设置正确的季集信息，这将影响后续的文件整理和入库操作。
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6 py-4">
          {/* 文件信息展示 */}
          <div className="p-4 rounded-lg bg-muted/50 border">
            <div className="flex items-start gap-3">
              <div className="p-2 rounded-md bg-green-500/10 flex-shrink-0">
                <FileVideo className="w-4 h-4 text-green-600" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium break-all leading-tight">
                  {file.fileName}
                </p>
                <div className="mt-2 flex items-center gap-2">
                  <Badge
                    variant="outline"
                    className="text-xs border-green-500/30 text-green-700 dark:text-green-300"
                  >
                    当前: S{formatEpisodeNumber(file.season || 0)}E
                    {formatEpisodeNumber(file.episode || 0)}
                  </Badge>
                </div>
              </div>
            </div>
          </div>

          {/* 季集数设置 */}
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              {/* 季数设置 */}
              <div className="space-y-2">
                <Label className="text-sm font-medium flex items-center gap-2">
                  <Calendar className="w-4 h-4 text-blue-600" />
                  季数
                </Label>
                <div className="flex items-center gap-2">
                  <Label className="text-xs text-muted-foreground whitespace-nowrap">
                    第
                  </Label>
                  <Input
                    type="number"
                    min="0"
                    max="99"
                    value={season}
                    onChange={(e) => setSeason(parseInt(e.target.value) || 0)}
                    className="flex-1 h-9 text-sm rounded-lg border-blue-500/20 focus:border-blue-500 focus:ring-blue-500"
                    placeholder="1"
                  />
                  <Label className="text-xs text-muted-foreground whitespace-nowrap">
                    季
                  </Label>
                </div>
              </div>

              {/* 集数设置 */}
              <div className="space-y-2">
                <Label className="text-sm font-medium flex items-center gap-2">
                  <Hash className="w-4 h-4 text-purple-600" />
                  集数
                </Label>
                <div className="flex items-center gap-2">
                  <Label className="text-xs text-muted-foreground whitespace-nowrap">
                    第
                  </Label>
                  <Input
                    type="number"
                    min="0"
                    max="999"
                    value={episode}
                    onChange={(e) => setEpisode(parseInt(e.target.value) || 0)}
                    className="flex-1 h-9 text-sm rounded-lg border-purple-500/20 focus:border-purple-500 focus:ring-purple-500"
                    placeholder="1"
                  />
                  <Label className="text-xs text-muted-foreground whitespace-nowrap">
                    集
                  </Label>
                </div>
              </div>
            </div>

            {/* 预览 */}
            <div className="p-3 rounded-lg bg-gradient-to-r from-green-500/5 to-blue-500/5 border border-green-500/20">
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">预览效果:</span>
                <Badge
                  variant="outline"
                  className="text-sm border-green-500/30 text-green-700 dark:text-green-300 bg-green-500/10"
                >
                  S{formatEpisodeNumber(season)}E{formatEpisodeNumber(episode)}
                </Badge>
              </div>
            </div>
          </div>
        </div>

        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={handleCancel}>
            取消
          </Button>
          <Button
            onClick={handleSave}
            className="bg-green-600 hover:bg-green-700"
          >
            保存设置
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
