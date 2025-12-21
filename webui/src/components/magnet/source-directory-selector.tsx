import { useState, useEffect, useMemo } from "react";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Folder,
  FolderOpen,
  ChevronRight,
  Home,
  Loader2,
  Search,
  FileText,
  Subtitles,
} from "lucide-react";
import magnetAPI, { type DirInfo, type FileInfo } from "@/api/magnet";
import { cn } from "@/lib/utils";
import { useToast } from "@/hooks/useToast";
import { extractErrorMessage } from "@/utils/error";
import { formatFileSize } from "@/utils/util";
export type SourceType = "dir" | "file";

export interface SourceSelection {
  path: string;
  type: SourceType;
}

interface SourceDirectorySelectorProps {
  initialPath?: string;
  selectedSource: SourceSelection | null;
  onSelectSource: (source: SourceSelection | null) => void;
  isMobile?: boolean;
  className?: string;
}

export function SourceDirectorySelector({
  initialPath = "/",
  selectedSource,
  onSelectSource,
  isMobile = false,
  className,
}: SourceDirectorySelectorProps) {
  const { toast } = useToast();
  const [currentPath, setCurrentPath] = useState<string>(initialPath);
  const [directories, setDirectories] = useState<DirInfo[]>([]);
  const [files, setFiles] = useState<FileInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [filePathSplit, setFilePathSplit] = useState<string>("/");
  const [fileRoots, setFileRoots] = useState<string[]>([]);
  const [showRoots, setShowRoots] = useState(false);
  const [searchTerm, setSearchTerm] = useState("");

  const loadDirectories = async (path: string) => {
    setLoading(true);
    setSearchTerm("");
    try {
      const resp = await magnetAPI.listDirs(path);
      setDirectories(resp.dirs);
      setFiles(resp.files);
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
      setFiles([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (showRoots) return;
    if (initialPath) {
      setCurrentPath(initialPath);
      loadDirectories(initialPath);
    }
  }, [initialPath]);

  const handleEnterDir = async (dirPath: string) => {
    setShowRoots(false);
    setCurrentPath(dirPath);
    await loadDirectories(dirPath);
  };

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

  const handleGoBack = async () => {
    const parentPath = getParentPath(currentPath);
    setCurrentPath(parentPath);
    await loadDirectories(parentPath);
  };

  const handleGoHome = async () => {
    setSearchTerm("");
    if (fileRoots.length <= 1) {
      setShowRoots(false);
      const rootPath = fileRoots[0] || filePathSplit || "/";
      setCurrentPath(rootPath);
      await loadDirectories(rootPath);
      return;
    }
    setShowRoots(true);
    setLoading(true);
    try {
      const rootDirs = await Promise.all(
        fileRoots.map(async (rootPath) => {
          try {
            const resp = await magnetAPI.listDirs(rootPath);
            return {
              path: rootPath,
              name: rootPath,
              hasDir: resp.dirs.length > 0,
              subtitleCount: 0,
            };
          } catch {
            return {
              path: rootPath,
              name: rootPath,
              hasDir: false,
              subtitleCount: 0,
            };
          }
        })
      );
      setDirectories(rootDirs);
      setFiles([]);
    } finally {
      setLoading(false);
    }
  };

  const pathSegments = useMemo(() => {
    const sep = filePathSplit || "/";
    const roots = fileRoots.length ? fileRoots : [sep];
    const matchedRoot =
      roots.find((r) => currentPath.startsWith(r)) || roots[0];

    let relative = currentPath.slice(matchedRoot.length);
    const segments = relative.split(sep).filter(Boolean);

    return [
      { name: matchedRoot, path: matchedRoot },
      ...segments.map((seg, i) => ({
        name: seg,
        path: matchedRoot + segments.slice(0, i + 1).join(sep),
      })),
    ];
  }, [currentPath, filePathSplit, fileRoots]);

  const filteredDirs = useMemo(() => {
    if (!searchTerm.trim()) return directories;
    const lowerTerm = searchTerm.toLowerCase();
    return directories.filter((dir) =>
      dir.name.toLowerCase().includes(lowerTerm)
    );
  }, [directories, searchTerm]);

  const filteredFiles = useMemo(() => {
    if (!searchTerm.trim()) return files;
    const lowerTerm = searchTerm.toLowerCase();
    return files.filter((file) => file.name.toLowerCase().includes(lowerTerm));
  }, [files, searchTerm]);

  const toggleFile = (filePath: string) => {
    if (selectedSource?.path === filePath) {
      onSelectSource(null);
    } else {
      onSelectSource({ path: filePath, type: "file" });
    }
  };

  const handleSelectDir = (dirPath: string) => {
    onSelectSource({ path: dirPath, type: "dir" });
  };

  // 截断名称，过长时省略
  const truncateName = (name: string, maxLength: number = 32): string => {
    if (name.length <= maxLength) return name;
    return name.slice(0, maxLength) + "...";
  };

  return (
    <div className={cn("space-y-2 h-full flex flex-col min-h-0", className)}>
      <div className="flex flex-col gap-1.5">
        <div className="flex items-center gap-1 overflow-x-auto py-0.5 scrollbar-hide no-scrollbar">
          <Button
            variant="ghost"
            size="icon"
            onClick={handleGoHome}
            className={cn("flex-shrink-0", isMobile ? "h-6 w-6" : "h-7 w-7")}
          >
            <Home className={cn(isMobile ? "w-3.5 h-3.5" : "w-4 h-4")} />
          </Button>
          {pathSegments.map((seg, i) => (
            <div key={seg.path} className="flex items-center flex-shrink-0">
              <ChevronRight className="w-3 h-3 text-muted-foreground mx-0.5" />
              <Button
                variant="ghost"
                size="sm"
                className={cn(
                  "font-normal",
                  isMobile ? "h-6 px-1 text-[11px]" : "h-7 px-1.5 text-xs",
                  i === pathSegments.length - 1 && "bg-muted font-medium"
                )}
                onClick={() => handleEnterDir(seg.path)}
              >
                {seg.name}
              </Button>
            </div>
          ))}
        </div>

        <div className="relative group">
          <Search
            className={cn(
              "absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground group-focus-within:text-blue-500",
              isMobile ? "w-3.5 h-3.5" : "w-4 h-4"
            )}
          />
          <Input
            placeholder="搜索文件夹或字幕..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className={cn(
              "pl-9 rounded-xl bg-muted/30 border-none focus-visible:ring-blue-500/30",
              isMobile ? "h-8 text-xs" : "h-9"
            )}
          />
        </div>
      </div>

      <ScrollArea className="flex-1 rounded-xl border bg-muted/10">
        <div className={cn("space-y-0.5", isMobile ? "p-1.5" : "p-2")}>
          {loading ? (
            <div className="flex items-center justify-center py-10">
              <Loader2 className="animate-spin w-6 h-6 text-blue-500" />
            </div>
          ) : (
            <>
              {filteredDirs.map((dir) => {
                const canEnter = dir.hasDir || dir.subtitleCount > 0;
                const chevronBoxClass = isMobile ? "w-6 h-6" : "w-7 h-7";
                return (
                  <div
                    key={dir.path}
                    className={cn(
                      "flex items-center gap-2 rounded-lg cursor-pointer transition-colors group",
                      isMobile ? "p-1.5" : "p-2",
                      selectedSource?.path === dir.path
                        ? "bg-blue-500/10"
                        : "hover:bg-accent"
                    )}
                    onClick={() => handleSelectDir(dir.path)}
                  >
                    {canEnter ? (
                      <div
                        className={cn(
                          "flex items-center justify-center rounded-md transition-colors",
                          "hover:bg-blue-500/20",
                          chevronBoxClass
                        )}
                        onClick={(e) => {
                          e.stopPropagation();
                          handleEnterDir(dir.path);
                        }}
                      >
                        <ChevronRight
                          className={cn(
                            isMobile ? "w-3.5 h-3.5" : "w-4 h-4",
                            "text-muted-foreground group-hover:text-blue-500"
                          )}
                        />
                      </div>
                    ) : (
                      <div className={chevronBoxClass} /> // 保持缩进对齐
                    )}
                    <Folder
                      className={cn(
                        isMobile ? "w-3.5 h-3.5" : "w-4 h-4",
                        "text-blue-500 fill-blue-500/10"
                      )}
                    />
                    <span className="flex-1 text-sm truncate" title={dir.name}>
                      {truncateName(dir.name, isMobile ? 30 : 45)}
                    </span>
                    {dir.subtitleCount > 0 && (
                      <span className="text-[10px] px-1.5 py-0.5 rounded-full bg-blue-500/10 text-blue-600 font-medium">
                        {dir.subtitleCount}
                      </span>
                    )}
                  </div>
                );
              })}

              {filteredFiles.map((file) => (
                <div
                  key={file.path}
                  className={cn(
                    "flex items-center gap-2 rounded-lg cursor-pointer transition-colors",
                    isMobile ? "p-1.5" : "p-2",
                    selectedSource?.path === file.path
                      ? "bg-blue-500/10 border border-blue-500/30"
                      : "hover:bg-accent border border-transparent"
                  )}
                  onClick={() => toggleFile(file.path)}
                >
                  <Subtitles
                    className={cn(
                      isMobile ? "w-3.5 h-3.5 ml-0.5" : "w-4 h-4 ml-1",
                      selectedSource?.path === file.path
                        ? "text-blue-600 fill-blue-600/10"
                        : "text-orange-500 fill-orange-500/10"
                    )}
                  />
                  <span
                    className={cn(
                      "flex-1 text-sm truncate",
                      selectedSource?.path === file.path &&
                        "font-medium text-blue-700"
                    )}
                    title={file.name}
                  >
                    {truncateName(file.name, isMobile ? 30 : 45)}
                  </span>
                  <span className="text-[10px] text-muted-foreground">
                    {formatFileSize(file.size)}
                  </span>
                </div>
              ))}

              {filteredDirs.length === 0 && filteredFiles.length === 0 && (
                <div className="flex flex-col items-center justify-center py-10 text-muted-foreground">
                  <Folder className="w-10 h-10 mb-2 opacity-20" />
                  <p className="text-sm">空空如也</p>
                </div>
              )}
            </>
          )}
        </div>
      </ScrollArea>

      {selectedSource && (
        <div className="p-3 rounded-xl bg-blue-500/5 border border-blue-500/10">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2 min-w-0">
              {selectedSource.type === "dir" ? (
                <>
                  <FolderOpen className="w-4 h-4 text-blue-500 flex-shrink-0" />
                  <span className="text-xs font-medium text-blue-700">
                    已选目录: {selectedSource.path.split(filePathSplit).pop()}
                  </span>
                </>
              ) : (
                <>
                  <Subtitles className="w-4 h-4 text-blue-500 flex-shrink-0" />
                  <span className="text-xs font-medium text-blue-700">
                    已选文件: {selectedSource.path.split(filePathSplit).pop()}
                  </span>
                </>
              )}
            </div>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onSelectSource(null)}
              className="h-6 px-2 text-[10px] hover:bg-blue-500/10 text-blue-600"
            >
              重置
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
