import { useState } from "react";
import { Search, Loader2, Check } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { cn } from "@/lib/utils";
import metaAPI, { type Meta } from "@/api/meta";
import { useToast } from "@/hooks/useToast";
import { extractErrorMessage } from "@/utils/error";

interface TMDBInputProps {
  type: "tv" | "movie"; // 区分番剧或剧场版
  value: number; // TMDB ID 值
  onTMDBIDChange: (tmdbID: number) => void; // 仅TMDB ID变化回调（手动输入时）
  onMetaChange?: (meta: Meta) => void; // 完整元数据变化回调（确认获取或搜索选择时）
  label?: string; // 可选标签
  placeholder?: string; // 占位符
  className?: string; // 自定义样式
  error?: string; // 错误提示
}

export function TMDBInput({
  type,
  value,
  onTMDBIDChange,
  onMetaChange,
  label = "TMDB ID",
  placeholder = "输入 TMDB ID 或点击搜索",
  className,
  error,
}: TMDBInputProps) {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const { toast } = useToast();

  // 处理手动输入TMDB ID
  const handleManualInput = (tmdbID: number) => {
    onTMDBIDChange(tmdbID);
  };

  // 确认获取完整的Meta信息
  const handleConfirmMeta = async () => {
    if (!value) {
      toast({
        title: "请输入TMDB ID",
        variant: "destructive",
      });
      return;
    }

    setLoading(true);
    try {
      // 根据类型获取Meta信息
      const meta =
        type === "tv"
          ? await metaAPI.getTVMeta(value)
          : await metaAPI.getMovieMeta(value);

      // 同时更新ID和完整元数据
      onTMDBIDChange(meta.tmdbID);
      if (onMetaChange) {
        onMetaChange(meta);
      }

      toast({
        title: "获取TMDB信息成功",
        variant: "default",
      });
    } catch (error) {
      const description = extractErrorMessage(error);
      toast({
        title: "获取TMDB信息失败",
        description: description,
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className={cn("grid gap-2", className)}>
      {label && <Label>{label}</Label>}
      <div className="flex gap-2">
        <Input
          inputMode="numeric"
          value={value || ""}
          onChange={(e) => {
            const newValue = e.target.value.trim();
            handleManualInput(newValue ? parseInt(newValue) : 0);
          }}
          placeholder={placeholder}
          className={cn("rounded-xl flex-1", error && "border-destructive")}
        />
        <Button
          type="button"
          variant="outline"
          size="icon"
          onClick={handleConfirmMeta}
          disabled={loading || !value}
          className="rounded-xl flex-shrink-0"
          title="确认获取TMDB信息"
        >
          {loading ? (
            <Loader2 className="h-4 w-4 animate-spin" />
          ) : (
            <Check className="h-4 w-4" />
          )}
        </Button>
        <Button
          type="button"
          variant="outline"
          size="icon"
          onClick={() => setDialogOpen(true)}
          className="rounded-xl flex-shrink-0"
          title="搜索TMDB"
        >
          <Search className="h-4 w-4" />
        </Button>
      </div>
      {error && <span className="text-sm text-destructive">{error}</span>}

      <SearchDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        type={type}
        onSelect={(tmdbID, meta) => {
          // 搜索选择时，同时更新ID和完整元数据
          if (meta) {
            onTMDBIDChange(meta.tmdbID);
            if (onMetaChange) {
              onMetaChange(meta);
            }
          }
          setDialogOpen(false);
        }}
      />
    </div>
  );
}

interface SearchDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  type: "tv" | "movie";
  onSelect: (tmdbID: number, meta?: Meta) => void;
}

function SearchDialog({
  open,
  onOpenChange,
  type,
  onSelect,
}: SearchDialogProps) {
  const { toast } = useToast();
  const [searchName, setSearchName] = useState("");
  const [loading, setLoading] = useState(false);
  const [results, setResults] = useState<Meta[]>([]);
  const [hasSearched, setHasSearched] = useState(false);

  // 执行搜索
  const handleSearch = async () => {
    if (!searchName.trim()) {
      toast({
        title: "请输入搜索名称",
        variant: "destructive",
      });
      return;
    }

    setLoading(true);
    setHasSearched(true);
    try {
      const data =
        type === "tv"
          ? await metaAPI.searchTVs(searchName)
          : await metaAPI.searchMovies(searchName);
      setResults(data);
    } catch (error) {
      const description = extractErrorMessage(error);
      toast({
        title: "搜索失败",
        description: description,
        variant: "destructive",
      });
      setResults([]);
    } finally {
      setLoading(false);
    }
  };

  // 重置状态
  const handleClose = () => {
    setSearchName("");
    setResults([]);
    setHasSearched(false);
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="w-[95dvw] sm:max-w-3xl max-h-[90dvh] flex flex-col rounded-xl">
        <DialogHeader>
          <DialogTitle className="text-xl anime-gradient-text">
            搜索{type === "tv" ? "番剧" : "剧场版"}
          </DialogTitle>
          <DialogDescription>
            输入{type === "tv" ? "番剧" : "剧场版"}名称进行搜索
          </DialogDescription>
        </DialogHeader>

        {/* 搜索输入区域 */}
        <div className="flex gap-2">
          <Input
            value={searchName}
            onChange={(e) => setSearchName(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter" && !loading) {
                handleSearch();
              }
            }}
            placeholder={`请输入${type === "tv" ? "番剧" : "剧场版"}名称`}
            className="rounded-xl flex-1"
            autoFocus
          />
          <Button
            onClick={handleSearch}
            disabled={loading}
            className="rounded-xl anime-button bg-gradient-to-r from-primary to-blue-500"
          >
            {loading ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                搜索中
              </>
            ) : (
              <>
                <Search className="h-4 w-4 mr-2" />
                搜索
              </>
            )}
          </Button>
        </div>

        {/* 搜索结果区域 */}
        <div className="flex-1 min-h-0">
          {loading ? (
            <div className="flex items-center justify-center h-[300px]">
              <Loader2 className="h-8 w-8 animate-spin text-primary" />
            </div>
          ) : hasSearched && results.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-[300px] text-muted-foreground">
              <Search className="h-12 w-12 mb-4 opacity-20" />
              <p>未找到相关结果</p>
              <p className="text-sm mt-2">请尝试其他关键词</p>
            </div>
          ) : results.length > 0 ? (
            <ScrollArea className="h-[400px] pr-4">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                {results.map((meta) => (
                  <ResultCard
                    key={meta.tmdbID}
                    meta={meta}
                    type={type}
                    onSelect={() => onSelect(meta.tmdbID, meta)}
                  />
                ))}
              </div>
            </ScrollArea>
          ) : (
            <div className="flex flex-col items-center justify-center h-[300px] text-muted-foreground">
              <Search className="h-12 w-12 mb-4 opacity-20" />
              <p>输入名称开始搜索</p>
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}

interface ResultCardProps {
  meta: Meta;
  type: "tv" | "movie";
  onSelect: () => void;
}

function ResultCard({ meta, type, onSelect }: ResultCardProps) {
  return (
    <Card className="overflow-hidden border-primary/10 hover:border-primary/30 transition-all duration-300 hover:shadow-lg">
      <div className="flex gap-3 p-3">
        {/* 海报图片 */}
        <div className="w-20 h-[120px] flex-shrink-0 rounded-lg overflow-hidden bg-muted">
          {meta.posterURL ? (
            <img
              src={meta.posterURL}
              alt={meta.chineseName}
              className="w-full h-full object-cover"
              loading="lazy"
            />
          ) : (
            <div className="w-full h-full flex items-center justify-center text-muted-foreground text-xs">
              无图片
            </div>
          )}
        </div>

        {/* 内容区域 */}
        <div className="flex-1 min-w-0 flex flex-col">
          {/* 标题和年份 */}
          <div className="mb-2">
            <h4 className="font-semibold text-sm line-clamp-1 mb-1">
              {meta.chineseName}
            </h4>
            <div className="flex items-center gap-2 flex-wrap">
              {meta.year && (
                <Badge variant="outline" className="text-xs">
                  {meta.year}
                </Badge>
              )}
              {type === "tv" && meta.season > 0 && (
                <Badge variant="outline" className="text-xs">
                  第 {meta.season} 季
                </Badge>
              )}
              {meta.episodeTotalNum > 0 && (
                <Badge variant="outline" className="text-xs">
                  {meta.episodeTotalNum} 集
                </Badge>
              )}
            </div>
          </div>

          {/* 类型标签 */}
          {meta.genres && (
            <div className="flex flex-wrap gap-1 mb-2">
              {meta.genres
                .split(",")
                .slice(0, 3)
                .map((genre, index) => (
                  <Badge
                    key={index}
                    variant="secondary"
                    className="text-xs px-2 py-0 bg-primary/10 text-primary border-none"
                  >
                    {genre.trim()}
                  </Badge>
                ))}
            </div>
          )}

          {/* 简介 */}
          {meta.overview && (
            <p className="text-xs text-muted-foreground line-clamp-2 mb-2">
              {meta.overview}
            </p>
          )}

          {/* 选择按钮 */}
          <div className="mt-auto">
            <Button
              size="sm"
              onClick={onSelect}
              className="rounded-lg w-full bg-gradient-to-r from-primary to-blue-500 anime-button"
            >
              选择此项 (ID: {meta.tmdbID})
            </Button>
          </div>
        </div>
      </div>
    </Card>
  );
}
