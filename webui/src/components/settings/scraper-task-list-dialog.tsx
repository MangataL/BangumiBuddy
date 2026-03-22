import { useState, useEffect, useCallback, useMemo } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Loader2,
  RefreshCw,
  Play,
  PlayCircle,
  ImageOff,
  FileText,
  Info,
  ListTodo,
  AlertCircle,
  Clock,
} from "lucide-react";
import { configAPI, type ScraperTask, type ScrapeStatus } from "@/api/config";
import { useToast } from "@/hooks/useToast";
import { extractErrorMessage } from "@/utils/error";
import { cn } from "@/lib/utils";
import {
  filterScraperTasksByBangumiName,
  getScraperTaskBangumiNames,
} from "./scraper-task-list-dialog.utils";

interface ScraperTaskListDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

// 刮削状态标签映射
const statusLabelMap: Record<ScrapeStatus, string> = {
  pending: "待刮削",
  missingTitle: "标题缺失",
  missingPlot: "剧情简介缺失",
  missingImage: "单集海报缺失",
};

const ALL_BANGUMI_OPTION = "__all_bangumi__";

function ScraperStatusBadges({ statuses }: { statuses: ScrapeStatus[] }) {
  if (!statuses || statuses.length === 0) return null;
  return (
    <div className="flex flex-wrap gap-2">
      {statuses.map((status) => {
        const isPending = status === "pending";
        const Icon = isPending ? Clock : AlertCircle;
        return (
          <Badge
            key={status}
            variant={isPending ? "secondary" : "destructive"}
            className={cn(
              "flex items-center gap-1.5 px-2.5 py-1 text-xs font-bold rounded-lg border shadow-sm transition-all duration-300",
              isPending
                ? "bg-muted/50 text-muted-foreground border-muted-foreground/20"
                : "bg-destructive/10 text-destructive border-destructive/20 hover:bg-destructive/20"
            )}
          >
            <Icon className={cn("h-3.5 w-3.5", !isPending && "animate-pulse")} />
            {statusLabelMap[status]}
          </Badge>
        );
      })}
    </div>
  );
}

function PosterImage({
  posterURL,
  bangumiName,
  season,
}: {
  posterURL: string;
  bangumiName: string;
  season: number;
}) {
  const [imgError, setImgError] = useState(false);

  return (
    <div className="relative group/poster flex-shrink-0 w-16 h-24 sm:w-20 sm:h-28 overflow-hidden rounded-xl border border-primary/5 shadow-sm group-hover/poster:shadow-md transition-shadow duration-300">
      {!posterURL || imgError ? (
        <div className="w-full h-full bg-muted/50 flex items-center justify-center">
          <ImageOff className="h-6 w-6 text-muted-foreground/50" />
        </div>
      ) : (
        <img
          src={posterURL}
          alt={bangumiName}
          className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-200"
          onError={() => setImgError(true)}
        />
      )}
      {/* 季度信息参考订阅日历界面展示在右上角 */}
      <div className="absolute top-0 right-0 bg-primary/80 backdrop-blur-sm text-primary-foreground text-[10px] sm:text-[11px] font-semibold px-2 py-0.5 rounded-bl-xl">
        S{season}
      </div>
    </div>
  );
}

function ScraperTaskItem({
  task,
  onTrigger,
  triggering,
}: {
  task: ScraperTask;
  onTrigger: (id: number) => void;
  triggering: boolean;
}) {
  return (
    <div className="relative overflow-hidden rounded-2xl border border-primary/5 bg-muted/20 p-3 sm:p-4 hover:bg-muted/30 hover:border-primary/20 transition-all duration-300 group">
      <div className="flex gap-4">
        {/* 左侧海报 */}
        <PosterImage
          posterURL={task.posterURL}
          bangumiName={task.bangumiName}
          season={task.season}
        />

        {/* 右侧信息 */}
        <div className="flex-1 min-w-0 flex flex-col justify-between py-0.5">
          <div className="space-y-2">
            <div className="flex items-start justify-between gap-2">
              <div className="min-w-0 flex-1">
                {/* 番剧名和集数信息 */}
                <div className="flex items-baseline gap-2 min-w-0">
                  <h4 className="font-bold text-base sm:text-lg leading-tight truncate text-primary">
                    {task.bangumiName}
                  </h4>
                  <span className="text-sm text-muted-foreground flex-shrink-0 font-medium">
                    第 {task.episode} 集
                  </span>
                </div>
              </div>

              {/* 触发刮削按钮 */}
              <Button
                variant="ghost"
                size="icon"
                className="h-9 w-9 rounded-full flex-shrink-0 bg-background/50 opacity-100 sm:opacity-40 sm:group-hover:opacity-100 transition-all hover:bg-primary hover:text-primary-foreground shadow-md border border-primary/5"
                onClick={() => onTrigger(task.id)}
                disabled={triggering}
              >
                {triggering ? (
                  <Loader2 className="h-5 w-5 animate-spin" />
                ) : (
                  <Play className="h-5 w-5 fill-current" />
                )}
              </Button>
            </div>

            <ScraperStatusBadges statuses={task.statuses} />
          </div>

          {/* 文件名区域 */}
          <div className="mt-3 flex items-center gap-2 px-2 text-muted-foreground/60">
            <FileText className="h-3 w-3 flex-shrink-0" />
            <p
              className="text-[10px] sm:text-xs truncate font-mono"
              title={task.filePath}
            >
              {task.filePath.split("/").pop()}
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}

interface ScraperTaskToolbarActionsProps {
  loading: boolean;
  triggeringAll: boolean;
  hasTasks: boolean;
  compact?: boolean;
  onRefresh: () => void;
  onTriggerAll: () => void;
}

function ScraperTaskToolbarActions({
  loading,
  triggeringAll,
  hasTasks,
  compact = false,
  onRefresh,
  onTriggerAll,
}: ScraperTaskToolbarActionsProps) {
  return (
    <>
      <Button
        variant="outline"
        className={cn(
          "rounded-xl anime-button border-primary/20 hover:bg-primary/5 transition-colors",
          compact ? "h-10 w-10 p-0 shadow-sm" : "px-3 py-2"
        )}
        onClick={onRefresh}
        disabled={loading}
        aria-label="刷新"
      >
        {loading ? (
          <Loader2 className={cn("h-4 w-4 animate-spin", !compact && "mr-1.5")} />
        ) : (
          <RefreshCw className={cn("h-4 w-4", !compact && "mr-1.5")} />
        )}
        {!compact && <span>刷新</span>}
      </Button>
      <Button
        className={cn(
          "rounded-xl anime-button bg-gradient-to-r from-primary to-blue-500 shadow-lg shadow-primary/20",
          compact ? "h-10 w-10 p-0" : "px-3 py-2"
        )}
        onClick={onTriggerAll}
        disabled={triggeringAll || !hasTasks || loading}
        aria-label="触发全部刮削"
      >
        {triggeringAll ? (
          <Loader2 className={cn("h-4 w-4 animate-spin", !compact && "mr-1.5")} />
        ) : (
          <PlayCircle className={cn("h-4 w-4", !compact && "mr-1.5")} />
        )}
        {!compact && <span>触发全部刮削</span>}
      </Button>
    </>
  );
}

interface BangumiFilterSelectProps {
  activeBangumiName: string;
  candidateBangumiNames: string[];
  disabled?: boolean;
  onSelectBangumiName: (value: string) => void;
}

function BangumiFilterSelect({
  activeBangumiName,
  candidateBangumiNames,
  disabled = false,
  onSelectBangumiName,
}: BangumiFilterSelectProps) {
  return (
    <Select
      value={activeBangumiName || ALL_BANGUMI_OPTION}
      onValueChange={(value) =>
        onSelectBangumiName(value === ALL_BANGUMI_OPTION ? "" : value)
      }
      disabled={disabled}
    >
      <SelectTrigger className="h-10 rounded-xl border-primary/10 bg-background/70 shadow-sm focus:ring-primary/20">
        <SelectValue placeholder="按番剧名称筛选" />
      </SelectTrigger>
      <SelectContent className="rounded-xl">
        <SelectItem value={ALL_BANGUMI_OPTION}>全部番剧</SelectItem>
        {candidateBangumiNames.map((bangumiName) => (
          <SelectItem key={bangumiName} value={bangumiName}>
            {bangumiName}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}

export function ScraperTaskListDialog({
  open,
  onOpenChange,
}: ScraperTaskListDialogProps) {
  const { toast } = useToast();
  const [tasks, setTasks] = useState<ScraperTask[]>([]);
  const [selectedBangumiName, setSelectedBangumiName] = useState("");
  const [loading, setLoading] = useState(false);
  const [triggeringAll, setTriggeringAll] = useState(false);
  const [triggeringIDs, setTriggeringIDs] = useState<Set<number>>(new Set());

  const loadTasks = useCallback(async () => {
    setLoading(true);
    try {
      const result = await configAPI.listScraperTasks();
      setTasks(result ?? []);
    } catch (error) {
      const desc = extractErrorMessage(error);
      toast({
        title: "获取刮削任务列表失败",
        description: desc,
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  }, [toast]);

  useEffect(() => {
    if (open) {
      loadTasks();
      return;
    }
    setSelectedBangumiName("");
  }, [open, loadTasks]);

  const bangumiOptions = useMemo(() => getScraperTaskBangumiNames(tasks), [tasks]);
  const filteredTasks = useMemo(
    () => filterScraperTasksByBangumiName(tasks, selectedBangumiName),
    [tasks, selectedBangumiName]
  );
  const isFiltering = selectedBangumiName.trim().length > 0;

  useEffect(() => {
    if (selectedBangumiName && !bangumiOptions.includes(selectedBangumiName)) {
      setSelectedBangumiName("");
    }
  }, [bangumiOptions, selectedBangumiName]);

  const handleTriggerTask = async (id: number) => {
    setTriggeringIDs((prev) => new Set(prev).add(id));
    try {
      await configAPI.triggerScrapeTask(id);
      toast({
        title: "刮削已触发",
        description: "正在处理该刮削任务",
      });
      await loadTasks();
    } catch (error) {
      const desc = extractErrorMessage(error);
      toast({
        title: "触发刮削失败",
        description: desc,
        variant: "destructive",
      });
    } finally {
      setTriggeringIDs((prev) => {
        const next = new Set(prev);
        next.delete(id);
        return next;
      });
    }
  };

  const handleTriggerAll = async () => {
    setTriggeringAll(true);
    try {
      await configAPI.triggerScrapeAll();
      toast({
        title: "全部刮削已触发",
        description: "正在处理所有刮削任务",
      });
      await loadTasks();
    } catch (error) {
      const desc = extractErrorMessage(error);
      toast({
        title: "触发全部刮削失败",
        description: desc,
        variant: "destructive",
      });
    } finally {
      setTriggeringAll(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="w-[calc(100vw-1.5rem)] max-w-[calc(100vw-1.5rem)] max-h-[calc(100svh-1.5rem)] overflow-hidden rounded-3xl border-primary/10 bg-background/95 p-4 backdrop-blur-md sm:w-full sm:max-w-2xl sm:p-6">
        <DialogHeader className="space-y-1">
          <DialogTitle className="text-2xl font-bold anime-gradient-text flex items-center gap-2">
            <ListTodo className="h-6 w-6 text-primary" />
            待刮削任务列表
          </DialogTitle>
          <DialogDescription className="flex items-center gap-1.5">
            <Info className="h-3.5 w-3.5" />
            以下番剧正在等待补全元数据，可手动触发补全任务
          </DialogDescription>
        </DialogHeader>

        {/* 控制区：范围 + 检索 + 动作 */}
        <div className="rounded-2xl border border-primary/10 bg-muted/20 p-3">
          <div className="grid gap-2.5 sm:grid-cols-[auto_minmax(0,1fr)_auto] sm:items-center">
            <div className="order-1 min-w-0 sm:order-2">
              <BangumiFilterSelect
                activeBangumiName={selectedBangumiName}
                candidateBangumiNames={bangumiOptions}
                disabled={loading || tasks.length === 0}
                onSelectBangumiName={setSelectedBangumiName}
              />
            </div>

            <div className="order-2 flex items-center justify-between gap-3 sm:order-1 sm:block">
              <Badge
                variant="secondary"
                className={cn(
                  "h-9 w-fit rounded-full px-3 py-0 border text-sm transition-colors duration-200",
                  isFiltering
                    ? "bg-primary/10 text-primary border-primary/20"
                    : "bg-primary/5 text-primary border-primary/10"
                )}
              >
                {isFiltering ? `${filteredTasks.length} / ${tasks.length}` : tasks.length} 个任务
              </Badge>

              <div className="flex items-center gap-1.5 sm:hidden">
                <ScraperTaskToolbarActions
                  compact
                  loading={loading}
                  triggeringAll={triggeringAll}
                  hasTasks={tasks.length > 0}
                  onRefresh={loadTasks}
                  onTriggerAll={handleTriggerAll}
                />
              </div>
            </div>

            <div className="hidden items-center gap-1.5 sm:order-3 sm:flex">
              <ScraperTaskToolbarActions
                loading={loading}
                triggeringAll={triggeringAll}
                hasTasks={tasks.length > 0}
                onRefresh={loadTasks}
                onTriggerAll={handleTriggerAll}
              />
            </div>
          </div>
        </div>

        {/* 列表区域 */}
        <ScrollArea className="h-[55vh] pr-2 sm:h-[480px] sm:pr-4">
          {loading ? (
            <div className="flex flex-col items-center justify-center h-60 gap-4">
              <div className="relative">
                <div className="h-12 w-12 rounded-full border-4 border-primary/20 border-t-primary animate-spin" />
                <Loader2 className="h-6 w-6 text-primary absolute inset-0 m-auto animate-pulse" />
              </div>
              <p className="text-sm text-muted-foreground animate-pulse font-medium">
                同步任务状态中...
              </p>
            </div>
          ) : tasks.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-60 gap-4 text-muted-foreground/40">
              <div className="p-6 rounded-full bg-muted/20 border-2 border-dashed border-muted">
                <PlayCircle className="h-16 w-16 opacity-20" />
              </div>
              <div className="text-center">
                <p className="text-base font-bold text-muted-foreground/60">
                  库中没有缺失数据的番剧
                </p>
                <p className="text-xs mt-1">所有元数据均已完美补全 ✨</p>
              </div>
            </div>
          ) : filteredTasks.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-60 gap-4 text-muted-foreground/40">
              <div className="p-6 rounded-full bg-muted/20 border-2 border-dashed border-muted">
                <ListTodo className="h-16 w-16 opacity-20" />
              </div>
              <div className="text-center">
                <p className="text-base font-bold text-muted-foreground/60">
                  未找到匹配的番剧
                </p>
                <p className="text-xs mt-1">试试切换到其他番剧或查看全部任务</p>
              </div>
            </div>
          ) : (
            <div className="grid grid-cols-1 gap-3 pb-4">
              {filteredTasks.map((task) => (
                <ScraperTaskItem
                  key={task.id}
                  task={task}
                  onTrigger={handleTriggerTask}
                  triggering={triggeringIDs.has(task.id)}
                />
              ))}
            </div>
          )}
        </ScrollArea>
      </DialogContent>
    </Dialog>
  );
}
