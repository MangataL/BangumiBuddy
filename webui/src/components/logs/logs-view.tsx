import { cn } from "@/lib/utils";

import { useState, useEffect, useCallback, useRef } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  RefreshCw,
  Search,
  AlertCircle,
  Info,
  Sparkles,
  Loader2,
} from "lucide-react";
import { LogService, LogEntry } from "@/api/log";
import { format } from "date-fns";

const LOG_FETCH_LIMIT = 50;

// 将Unix时间戳转换为可读时间格式
const formatTimestamp = (ts: number): string => {
  try {
    // 检查是否为秒级时间戳
    const date = new Date(ts * 1000);
    return format(date, "yyyy-MM-dd HH:mm:ss");
  } catch (e) {
    return "未知时间";
  }
};

export default function LogsView() {
  const [logLevel, setLogLevel] = useState("all");
  const [searchQuery, setSearchQuery] = useState("");
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [offset, setOffset] = useState(0);
  const [hasMore, setHasMore] = useState(true);
  const requestIdRef = useRef(0);

  const fetchLogs = useCallback(
    async (nextOffset: number, reset = false) => {
      const requestId = requestIdRef.current + 1;
      requestIdRef.current = requestId;
      setLoading(true);
      try {
        const logData = await LogService.getLogs({
          level: logLevel !== "all" ? logLevel : undefined,
          keyword: searchQuery || undefined,
          limit: LOG_FETCH_LIMIT,
          offset: nextOffset,
        });
        if (requestId === requestIdRef.current) {
          setLogs((prev) => (reset ? logData : [...prev, ...logData]));
          setOffset(nextOffset + logData.length);
          setHasMore(logData.length === LOG_FETCH_LIMIT);
        }
      } catch (error) {
        if (requestId === requestIdRef.current) {
          console.error("加载日志失败:", error);
        }
      } finally {
        if (requestId === requestIdRef.current) {
          setLoading(false);
        }
      }
    },
    [logLevel, searchQuery]
  );

  // 首次加载和日志级别或搜索查询变化时获取日志
  useEffect(() => {
    setLogs([]);
    setOffset(0);
    setHasMore(true);
    fetchLogs(0, true);
  }, [logLevel, searchQuery, fetchLogs]);

  // 处理刷新按钮点击
  const handleRefresh = () => {
    setLogs([]);
    setOffset(0);
    setHasMore(true);
    fetchLogs(0, true);
  };

  const handleScroll = (event: React.UIEvent<HTMLDivElement>) => {
    if (loading || !hasMore) {
      return;
    }
    const { scrollTop, scrollHeight, clientHeight } = event.currentTarget;
    if (scrollHeight - scrollTop - clientHeight < 80) {
      fetchLogs(offset);
    }
  };

  const getLogIcon = (level: string) => {
    switch (level.toLowerCase()) {
      case "error":
        return <AlertCircle className="h-4 w-4 text-destructive" />;
      case "success":
      case "info":
        return <Info className="h-4 w-4 text-blue-500" />;
      case "debug":
        return <Info className="h-4 w-4 text-purple-500" />;
      case "warn":
      case "warning":
        return <AlertCircle className="h-4 w-4 text-yellow-500" />;
      default:
        return <Info className="h-4 w-4 text-blue-500" />;
    }
  };

  const getLogStyle = (level: string) => {
    switch (level.toLowerCase()) {
      case "error":
        return "border-destructive/50 bg-destructive/10";
      case "warn":
      case "warning":
        return "border-yellow-500/50 bg-yellow-500/10";
      case "success":
        return "border-green-500/50 bg-green-500/10";
      case "debug":
        return "border-purple-500/50 bg-purple-500/10";
      case "info":
      default:
        return "border-blue-500/50 bg-blue-500/10";
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold flex items-center gap-2">
          <Sparkles className="h-6 w-6 text-primary animate-pulse anime-gradient-text" />
          <span className="anime-gradient-text">系统日志</span>
        </h1>
        <p className="text-muted-foreground">查看系统运行日志和错误信息</p>
      </div>

      <Card className="border-primary/10 rounded-xl overflow-hidden">
        <CardHeader className="pb-3 bg-gradient-to-r from-primary/5 to-blue-500/5">
          <CardTitle className="text-xl anime-gradient-text">
            日志记录
          </CardTitle>
          <CardDescription>查看和筛选系统日志</CardDescription>
          <div className="flex items-center gap-2 sm:gap-4 pt-4">
            <div className="flex-1 min-w-0">
              <div className="relative">
                <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                  type="search"
                  placeholder="搜索日志..."
                  className="pl-8 rounded-xl"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                />
              </div>
            </div>
            <Select value={logLevel} onValueChange={setLogLevel}>
              <SelectTrigger className="w-[100px] sm:w-[180px] rounded-xl flex-shrink-0">
                <SelectValue placeholder="选择日志级别" />
              </SelectTrigger>
              <SelectContent className="rounded-xl">
                <SelectItem value="all">全部日志</SelectItem>
                <SelectItem value="info">信息</SelectItem>
                <SelectItem value="debug">调试</SelectItem>
                <SelectItem value="warn">警告</SelectItem>
                <SelectItem value="error">错误</SelectItem>
              </SelectContent>
            </Select>
            <Button
              variant="outline"
              size="icon"
              className="rounded-xl anime-button flex-shrink-0"
              onClick={handleRefresh}
              disabled={loading}
            >
              <RefreshCw className={cn("h-4 w-4", loading && "animate-spin")} />
            </Button>
          </div>
        </CardHeader>
        <CardContent className="p-4">
          <div
            className="h-[500px] overflow-y-auto pr-4"
            onScroll={handleScroll}
          >
            {logs.length === 0 && loading ? (
              <div className="flex flex-col items-center justify-center h-[400px] text-muted-foreground">
                <Loader2 className="h-8 w-8 animate-spin mb-2" />
                <p>日志加载中...</p>
              </div>
            ) : logs.length === 0 ? (
              <div className="flex flex-col items-center justify-center h-[400px] text-muted-foreground">
                <AlertCircle className="h-8 w-8 mb-2" />
                <p>暂无日志数据</p>
              </div>
            ) : (
              <div className="space-y-2">
                {logs.map((log, index) => (
                  <div
                    key={index}
                    className={cn(
                      "flex items-start gap-2 rounded-xl border p-3 transition-all duration-200 hover:shadow-md",
                      getLogStyle(log.level)
                    )}
                  >
                    <div className="mt-0.5">{getLogIcon(log.level)}</div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center justify-between">
                        <span className="font-medium">
                          {formatTimestamp(log.ts)}
                        </span>
                        <span className="text-xs px-2 py-0.5 rounded-full bg-background">
                          {log.level.toUpperCase()}
                        </span>
                      </div>
                      <p className="mt-1 text-sm break-all whitespace-pre-wrap">
                        {log.message}
                      </p>
                    </div>
                  </div>
                ))}
                <div className="flex items-center justify-center py-3 text-sm text-muted-foreground">
                  {loading ? (
                    <div className="flex items-center gap-2">
                      <Loader2 className="h-4 w-4 animate-spin" />
                      <span>加载更多日志...</span>
                    </div>
                  ) : hasMore ? (
                    <span>向下滚动以加载更多日志</span>
                  ) : (
                    <span>没有更多日志了</span>
                  )}
                </div>
              </div>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
