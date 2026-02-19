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
import { Badge } from "@/components/ui/badge";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { FileVideo, Film, Tv, RotateCcw } from "lucide-react";
import {
  TorrentFile,
  TorrentFileMeta,
  DownloadTypeSet,
  DownloadTypeLabels,
  type DownloadType,
} from "@/api/magnet";
import { TMDBInput } from "@/components/tmdb";
import { type Meta } from "@/api/meta";

interface FileMetaEditDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  file: TorrentFile | null;
  /** 任务级别的下载类型，作为默认值 */
  taskDownloadType: DownloadType;
  onSave: (fileName: string, updates: Partial<TorrentFile>) => void;
}

export function FileMetaEditDialog({
  open,
  onOpenChange,
  file,
  taskDownloadType,
  onSave,
}: FileMetaEditDialogProps) {
  const [mediaType, setMediaType] = useState<DownloadType>(taskDownloadType);
  const [tmdbID, setTmdbID] = useState<number>(0);
  const [chineseName, setChineseName] = useState<string>("");
  const [year, setYear] = useState<string>("");
  const [hasCustomMeta, setHasCustomMeta] = useState(false);
  const [tmdbError, setTmdbError] = useState("");
  const isValidTMDBID = Number.isInteger(tmdbID) && tmdbID > 0;

  // 当文件变化时，更新本地状态
  useEffect(() => {
    if (file) {
      if (file.meta) {
        setMediaType(file.meta.mediaType || taskDownloadType);
        setTmdbID(file.meta.tmdbID || 0);
        setChineseName(file.meta.chineseName || "");
        setYear(file.meta.year || "");
        setHasCustomMeta(true);
      } else {
        setMediaType(taskDownloadType);
        setTmdbID(0);
        setChineseName("");
        setYear("");
        setHasCustomMeta(false);
      }
      setTmdbError("");
    }
  }, [file, taskDownloadType]);

  // 处理TMDB元数据变化
  const handleMetaChange = (meta: Meta) => {
    setTmdbID(meta.tmdbID);
    setChineseName(meta.chineseName || "");
    setYear(meta.year || "");
    if (meta.tmdbID > 0) {
      setTmdbError("");
    }
  };

  // 处理保存
  const handleSave = () => {
    if (!file) return;
    if (!isValidTMDBID) {
      setTmdbError("请填写 TMDB ID");
      return;
    }

    const fileMeta: TorrentFileMeta = {
      mediaType,
      tmdbID,
      chineseName,
      year,
    };

    onSave(file.fileName, { meta: fileMeta });
    onOpenChange(false);
  };

  // 重置为使用任务级别元数据
  const handleReset = () => {
    if (!file) return;

    onSave(file.fileName, { meta: undefined });
    onOpenChange(false);
  };

  // 处理取消
  const handleCancel = () => {
    onOpenChange(false);
  };

  if (!file) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="rounded-xl border-primary/20 bg-card/95 backdrop-blur-md w-[92vw] max-w-[92vw] p-4 sm:max-w-md sm:p-6">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <FileVideo className="w-5 h-5 text-primary" />
            <span className="anime-gradient-text">自定义文件元数据</span>
          </DialogTitle>
          <DialogDescription>
            为该文件设置独立的媒体类型和 TMDB
            信息，覆盖任务级别的默认设置
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-5 py-4">
          {/* 文件信息展示 */}
          <div className="p-3 rounded-lg bg-muted/50 border">
            <div className="flex items-start gap-2.5">
              <div className="p-1.5 rounded-md bg-primary/10 flex-shrink-0">
                <FileVideo className="w-4 h-4 text-primary" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium break-all leading-tight">
                  {file.fileName.split("/").pop()}
                </p>
                {hasCustomMeta && (
                  <Badge
                    variant="outline"
                    className="mt-1.5 text-xs border-primary/30 text-primary"
                  >
                    已自定义
                  </Badge>
                )}
              </div>
            </div>
          </div>

          {/* 媒体类型选择 */}
          <div className="space-y-2">
            <Label className="text-sm font-medium flex items-center gap-2">
              {mediaType === DownloadTypeSet.TV ? (
                <Tv className="w-4 h-4 text-blue-600" />
              ) : (
                <Film className="w-4 h-4 text-pink-600" />
              )}
              媒体类型
            </Label>
            <Select
              value={mediaType}
              onValueChange={(value) => setMediaType(value as DownloadType)}
            >
              <SelectTrigger className="rounded-lg border-primary/20">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={DownloadTypeSet.TV}>
                  <div className="flex items-center gap-2">
                    <Tv className="w-3.5 h-3.5 text-blue-500" />
                    {DownloadTypeLabels[DownloadTypeSet.TV]}
                  </div>
                </SelectItem>
                <SelectItem value={DownloadTypeSet.Movie}>
                  <div className="flex items-center gap-2">
                    <Film className="w-3.5 h-3.5 text-pink-500" />
                    {DownloadTypeLabels[DownloadTypeSet.Movie]}
                  </div>
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* TMDB ID 输入 */}
          <div className="space-y-2">
            <TMDBInput
              type={mediaType}
              value={tmdbID}
              onTMDBIDChange={(id) => {
                setTmdbID(id);
                if (Number.isInteger(id) && id > 0) {
                  setTmdbError("");
                }
              }}
              onMetaChange={handleMetaChange}
              label="TMDB ID"
              placeholder="输入 TMDB ID 或点击搜索"
              error={tmdbError}
            />
          </div>

          {/* 预览效果 */}
          {(tmdbID > 0 || chineseName) && (
            <div className="p-3 rounded-lg bg-gradient-to-r from-primary/5 to-blue-500/5 border border-primary/20">
              <div className="space-y-2">
                <span className="text-sm text-muted-foreground">
                  预览效果:
                </span>
                <div className="flex items-center gap-2 flex-wrap">
                  <Badge
                    variant="outline"
                    className={
                      mediaType === DownloadTypeSet.TV
                        ? "text-xs border-blue-500/30 text-blue-700 dark:text-blue-300"
                        : "text-xs border-pink-500/30 text-pink-700 dark:text-pink-300"
                    }
                  >
                    {mediaType === DownloadTypeSet.TV ? (
                      <Tv className="w-3 h-3 mr-1" />
                    ) : (
                      <Film className="w-3 h-3 mr-1" />
                    )}
                    {DownloadTypeLabels[mediaType]}
                  </Badge>
                  {chineseName && (
                    <Badge
                      variant="secondary"
                      className="text-xs bg-primary/10 text-primary"
                    >
                      {chineseName}
                    </Badge>
                  )}
                  {year && (
                    <Badge variant="secondary" className="text-xs">
                      {year}
                    </Badge>
                  )}
                  {tmdbID > 0 && (
                    <Badge variant="outline" className="text-xs">
                      TMDB: {tmdbID}
                    </Badge>
                  )}
                </div>
              </div>
            </div>
          )}
        </div>

        <DialogFooter className="gap-2 flex-col-reverse sm:flex-row">
          {hasCustomMeta && (
            <Button
              variant="outline"
              onClick={handleReset}
              className="rounded-xl text-orange-600 hover:text-orange-700 hover:bg-orange-500/10 border-orange-500/30"
            >
              <RotateCcw className="w-4 h-4 mr-1.5" />
              恢复默认
            </Button>
          )}
          <div className="flex-1" />
          <Button variant="outline" onClick={handleCancel} className="rounded-xl">
            取消
          </Button>
          <Button
            onClick={handleSave}
            className="rounded-xl bg-gradient-to-r from-primary to-blue-500 anime-button"
          >
            保存设置
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
