import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Folder, FolderOpen, ChevronRight, Home, Loader2 } from "lucide-react";
import magnetAPI, { type FileDir } from "@/api/magnet";
import { cn } from "@/lib/utils";
import { useToast } from "@/hooks/useToast";
import { extractErrorMessage } from "@/utils/error";
interface SourceDirectorySelectorProps {
  /** 初始路径 */
  initialPath?: string;
  /** 选中的目录路径 */
  selectedDir: string;
  /** 目录选中时的回调 */
  onSelectDir: (path: string) => void;
  /** 是否为移动端 */
  isMobile?: boolean;
  className?: string;
}

export function SourceDirectorySelector({
  initialPath = "/",
  selectedDir,
  onSelectDir,
  isMobile = false,
  className,
}: SourceDirectorySelectorProps) {
  const { toast } = useToast();
  const [currentPath, setCurrentPath] = useState<string>(initialPath);
  const [directories, setDirectories] = useState<FileDir[]>([]);
  const [loading, setLoading] = useState(false);
  const [filePathSplit, setFilePathSplit] = useState<string>("/");
  const [fileRoots, setFileRoots] = useState<string[]>([]);
  const [showRoots, setShowRoots] = useState(false);

  // 加载指定路径下的目录
  const loadDirectories = async (path: string) => {
    setLoading(true);
    try {
      const resp = await magnetAPI.listDirs(path);
      setDirectories(resp.dirs);
      setFilePathSplit(resp.filePathSplit);
      setFileRoots(resp.fileRoots);
    } catch (error) {
      const description = extractErrorMessage(error);
      toast({
        title: "加载目录失败",
        description: description,
        variant: "destructive",
      });
      setDirectories([]);
    } finally {
      setLoading(false);
    }
  };

  // 初始加载和初始路径变化时加载目录
  useEffect(() => {
    if (showRoots) return;
    if (initialPath) {
      setCurrentPath(initialPath);
      loadDirectories(initialPath);
    }
  }, [initialPath]);

  // 进入子目录
  const handleEnterDir = async (dirPath: string) => {
    setShowRoots(false);
    setCurrentPath(dirPath);
    await loadDirectories(dirPath);
  };

  // 计算上级目录
  const getParentPath = (path: string): string => {
    const sep = filePathSplit || "/";
    const roots = fileRoots.length ? fileRoots : [sep];
    const matchedRoot = roots.find((r) => path.startsWith(r)) || roots[0];

    if (path === matchedRoot) return matchedRoot;

    let stripped = path;
    if (stripped !== matchedRoot) {
      while (stripped.endsWith(sep)) stripped = stripped.slice(0, -sep.length);
    }
    const idx = stripped.lastIndexOf(sep);
    if (idx < matchedRoot.length) return matchedRoot;
    return stripped.slice(0, idx);
  };

  // 返回上级目录
  const handleGoBack = async () => {
    const parentPath = getParentPath(currentPath);
    setCurrentPath(parentPath);
    await loadDirectories(parentPath);
  };

  // 回到根目录
  const handleGoHome = async () => {
    if (fileRoots.length <= 1) {
      setShowRoots(false);
      const rootPath = fileRoots[0] || filePathSplit || "/";
      setCurrentPath(rootPath);
      await loadDirectories(rootPath);
      return;
    }
    // 多根盘场景：直接展示 roots
    setShowRoots(true);
    // 将 fileRoots 转换为 FileDir 格式，并检查每个根盘
    setLoading(true);
    try {
      const rootDirs = await Promise.all(
        fileRoots.map(async (rootPath) => {
          try {
            const resp = await magnetAPI.listDirs(rootPath);
            return {
              path: rootPath,
              hasDir: resp.dirs.length > 0,
              subtitleCount: 0,
            };
          } catch {
            return { path: rootPath, hasDir: false, subtitleCount: 0 };
          }
        })
      );
      setDirectories(rootDirs);
    } finally {
      setLoading(false);
    }
  };

  // 获取目录名称（路径的最后一部分）
  const getDirName = (path: string): string => {
    const sep = filePathSplit || "/";
    // 保持根盘完整显示，例如 C:\
    if (fileRoots.includes(path)) return path;
    const parts = path.split(sep).filter(Boolean);
    return parts[parts.length - 1] || sep;
  };

  // 截断文件夹名称，过长时省略
  const truncateDirName = (name: string, maxLength: number = 30): string => {
    if (name.length <= maxLength) return name;
    return name.slice(0, maxLength) + "...";
  };

  return (
    <div className={cn("space-y-3 h-full flex flex-col min-h-0", className)}>
      {/* 当前路径导航 */}
      <div className="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          onClick={handleGoHome}
          disabled={
            (fileRoots.length <= 1
              ? currentPath === (fileRoots[0] || filePathSplit || "/")
              : showRoots) || loading
          }
          className={cn(isMobile ? "h-7 px-2" : "h-8 px-2")}
        >
          <Home className={cn(isMobile ? "w-3 h-3" : "w-3.5 h-3.5")} />
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={handleGoBack}
          disabled={
            (fileRoots.length <= 1
              ? currentPath === (fileRoots[0] || filePathSplit || "/")
              : showRoots) || loading
          }
          className={cn(isMobile ? "h-7 px-2" : "h-8 px-2")}
        >
          <ChevronRight
            className={cn("rotate-180", isMobile ? "w-3 h-3" : "w-3.5 h-3.5")}
          />
        </Button>
        <div
          className={cn(
            "flex-1 rounded-lg bg-muted/50 border overflow-x-auto whitespace-nowrap scrollbar-hide",
            isMobile ? "p-1.5 text-xs" : "p-2 text-xs"
          )}
        >
          <code>{currentPath}</code>
        </div>
      </div>

      {/* 已选择的源目录 */}
      {selectedDir && (
        <div className="p-3 rounded-lg bg-blue-500/5 border border-blue-500/20">
          <div className="flex items-center justify-between gap-2">
            <div className="flex items-center gap-2 flex-1 min-w-0">
              <FolderOpen
                className={cn(
                  "text-blue-600 flex-shrink-0",
                  isMobile ? "w-3.5 h-3.5" : "w-4 h-4"
                )}
              />
              <code
                className={cn(
                  "break-all text-blue-700 dark:text-blue-300",
                  isMobile ? "text-xs" : "text-xs"
                )}
              >
                {selectedDir}
              </code>
            </div>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onSelectDir("")}
              className={cn(
                "text-xs flex-shrink-0",
                isMobile ? "h-5 px-1.5" : "h-6 px-2"
              )}
            >
              清除
            </Button>
          </div>
        </div>
      )}

      {/* 目录列表 */}
      <ScrollArea className="flex-1 min-h-0 rounded-lg border bg-muted/20">
        <div className="p-2 space-y-1">
          {loading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2
                className={cn(
                  "animate-spin text-muted-foreground",
                  isMobile ? "w-5 h-5" : "w-6 h-6"
                )}
              />
            </div>
          ) : directories.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-8 text-muted-foreground">
              <Folder
                className={cn(isMobile ? "w-6 h-6 mb-2" : "w-8 h-8 mb-2")}
              />
              <p className={cn(isMobile ? "text-xs" : "text-sm")}>此目录为空</p>
            </div>
          ) : (
            directories.map((dir, index) => {
              return (
                <div
                  key={index}
                  className={cn(
                    "flex items-center gap-2 rounded-lg transition-all hover:bg-accent cursor-pointer",
                    isMobile ? "p-2" : "p-2",
                    selectedDir === dir.path &&
                      "bg-blue-500/10 border border-blue-500/30"
                  )}
                  onClick={() => onSelectDir(dir.path)}
                >
                  {/* 只有当目录有子目录时才显示箭头 */}
                  {dir.hasDir && (
                    <div
                      className="p-1 hover:bg-blue-500/20 rounded transition-colors cursor-pointer flex-shrink-0"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleEnterDir(dir.path);
                      }}
                    >
                      <ChevronRight
                        className={cn(
                          "text-muted-foreground",
                          isMobile ? "w-4 h-4" : "w-4 h-4"
                        )}
                      />
                    </div>
                  )}
                  {/* 没有子目录时用占位符保持布局对齐 */}
                  {!dir.hasDir && (
                    <div
                      className={cn("flex-shrink-0", isMobile ? "w-6" : "w-6")}
                    />
                  )}
                  <div
                    className={cn(
                      "rounded-md flex-shrink-0 bg-blue-500/10 hover:bg-blue-500/20",
                      isMobile ? "p-1.5" : "p-1.5"
                    )}
                  >
                    <Folder
                      className={cn(
                        "text-blue-600",
                        isMobile ? "w-3.5 h-3.5" : "w-3.5 h-3.5"
                      )}
                    />
                  </div>
                  <span
                    className={cn(
                      "flex-1 min-w-0 text-foreground",
                      isMobile ? "text-sm" : "text-sm"
                    )}
                    title={getDirName(dir.path)}
                  >
                    {truncateDirName(getDirName(dir.path), isMobile ? 20 : 30)}
                  </span>
                  <span
                    className={cn(
                      "text-xs px-2 py-0.5 rounded-full bg-muted text-muted-foreground flex-shrink-0",
                      isMobile ? "ml-1" : "ml-2"
                    )}
                  >
                    {dir.subtitleCount ?? 0}
                  </span>
                </div>
              );
            })
          )}
        </div>
      </ScrollArea>

      {/* 操作提示 */}
      {!isMobile && (
        <div
          className={cn(
            "flex items-start gap-2 p-2 rounded-lg bg-muted/50 text-muted-foreground",
            "text-xs"
          )}
        >
          <div className="flex-shrink-0 mt-0.5">💡</div>
          <div className="space-y-1">
            <p>
              • 点击选择目录，点击左侧箭头进入，数字是该目录下的字幕文件数量
              {fileRoots.length > 1 ? "（根目录直接展示盘符）" : ""}
            </p>
            <p>• 选择包含字幕文件的目录</p>
          </div>
        </div>
      )}
    </div>
  );
}
