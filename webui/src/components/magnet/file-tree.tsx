import { useState, useMemo } from "react";
import {
  ChevronRight,
  ChevronDown,
  File,
  Folder,
  FolderOpen,
  FileVideo,
  CheckCircle2,
  Eye,
  Minus,
  CloudDownload,
  CloudOff,
  Layers,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { TorrentFile, type DownloadType } from "@/api/magnet";
import { cn } from "@/lib/utils";
import { useMobile } from "@/hooks/useMobile";
import { EpisodeEditDialog } from "./episode-edit-dialog";
import { BatchEpisodeEditDialog } from "./batch-episode-edit-dialog";
import { SubtitleTransferDialog } from "./subtitle-transfer-dialog";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { TooltipProvider } from "@/components/ui/tooltip";
import {
  Tooltip,
  TooltipTrigger,
  TooltipContent,
} from "@/components/ui/tooltip";

interface TreeNode {
  name: string;
  path: string;
  isDirectory: boolean;
  children: TreeNode[];
  file?: TorrentFile;
}

interface FileTreeProps {
  files: TorrentFile[];
  downloadType: DownloadType;
  taskID: string;
  onFileChange: (fileName: string, updates: Partial<TorrentFile>) => void;
  onFilesChange: (
    updates: { fileName: string; updates: Partial<TorrentFile> }[]
  ) => void;
  onSubtitleTransferSuccess?: () => void;
  defaultExpandLevel?: number; // 默认展开层级，0 表示不展开
}

// 构建目录树
function buildTree(files: TorrentFile[]): TreeNode {
  const root: TreeNode = {
    name: "root",
    path: "",
    isDirectory: true,
    children: [],
  };

  files.forEach((file) => {
    const parts = file.fileName.split("/");
    let currentNode = root;

    parts.forEach((part, index) => {
      const isLastPart = index === parts.length - 1;
      const path = parts.slice(0, index + 1).join("/");

      let childNode = currentNode.children.find((child) => child.name === part);

      if (!childNode) {
        childNode = {
          name: part,
          path,
          isDirectory: !isLastPart,
          children: [],
          file: isLastPart ? file : undefined,
        };
        currentNode.children.push(childNode);
      }

      currentNode = childNode;
    });
  });

  return root;
}

// 树节点组件
function TreeNodeComponent({
  node,
  allFiles,
  downloadType,
  taskID,
  onFileChange,
  onFilesChange,
  onSubtitleTransferSuccess,
  level = 0,
  defaultExpandLevel = 0,
  isMobile = false,
}: {
  node: TreeNode;
  allFiles: TorrentFile[];
  downloadType: DownloadType;
  taskID: string;
  onFileChange: (fileName: string, updates: Partial<TorrentFile>) => void;
  onFilesChange: (
    updates: { fileName: string; updates: Partial<TorrentFile> }[]
  ) => void;
  onSubtitleTransferSuccess?: () => void;
  level?: number;
  defaultExpandLevel?: number;
  isMobile?: boolean;
}) {
  const [expanded, setExpanded] = useState(level < defaultExpandLevel);
  const [isHovered, setIsHovered] = useState(false);
  const [showEpisodeDialog, setShowEpisodeDialog] = useState(false);
  const [showBatchDialog, setShowBatchDialog] = useState(false);

  // 智能截断文件夹名称 - 仅在移动端对文件夹生效
  const truncateFolderName = (name: string): string => {
    if (!isMobile || name.length <= 25) {
      return name;
    }

    // 尝试在合适的位置截断（优先在空格、符号处）
    const breakPoints = [" ", "-", "_", "[", "(", "."];
    let bestPos = -1;

    for (let i = 20; i <= 25 && i < name.length; i++) {
      if (breakPoints.includes(name[i])) {
        bestPos = i;
        break;
      }
    }

    // 如果找到合适的断点，在那里截断
    if (bestPos > 0) {
      return name.substring(0, bestPos) + "...";
    }

    // 否则直接在25个字符处截断
    return name.substring(0, 25) + "...";
  };

  // 获取该节点下所有的文件
  const getAllFiles = (currentNode: TreeNode): TorrentFile[] => {
    let results: TorrentFile[] = [];
    if (currentNode.file) {
      results.push(currentNode.file);
    }
    currentNode.children.forEach((child) => {
      results = results.concat(getAllFiles(child));
    });
    return results;
  };

  // 目录节点
  if (node.isDirectory) {
    const folderColors = [
      "from-blue-500/5 to-transparent border-blue-500/20",
      "from-purple-500/5 to-transparent border-purple-500/20",
      "from-pink-500/5 to-transparent border-pink-500/20",
      "from-green-500/5 to-transparent border-green-500/20",
    ];
    const colorIndex = level % folderColors.length;

    const childrenFiles = getAllFiles(node);
    const allSelected = childrenFiles.every((f) => f.download);
    const someSelected = childrenFiles.some((f) => f.download);
    const isIndeterminate = someSelected && !allSelected;

    const toggleFolderDownload = (e: React.MouseEvent) => {
      e.stopPropagation();
      const targetStatus = !allSelected;
      const updates = childrenFiles.map((f) => ({
        fileName: f.fileName,
        updates: { download: targetStatus },
      }));
      onFilesChange(updates);
    };

    return (
      <div className="select-none my-1">
        <div
          className={cn(
            "flex items-center gap-2 py-2 px-3 rounded-lg cursor-pointer transition-all duration-200",
            "hover:scale-[1.02] anime-card",
            expanded
              ? `bg-gradient-to-r ${folderColors[colorIndex]} border`
              : "hover:bg-accent border border-transparent"
          )}
          style={{ marginLeft: `${level * 16}px` }}
          onClick={() => setExpanded(!expanded)}
          onMouseEnter={() => setIsHovered(true)}
          onMouseLeave={() => setIsHovered(false)}
        >
          <div className="flex items-center gap-1.5 flex-shrink-0">
            <div
              className="p-1 hover:bg-primary/10 rounded transition-colors group/dl"
              onClick={toggleFolderDownload}
            >
              {allSelected ? (
                <CloudDownload className="w-4 h-4 text-blue-500 animate-pulse" />
              ) : isIndeterminate ? (
                <div className="relative">
                  <CloudDownload className="w-4 h-4 text-blue-500/50" />
                  <Minus className="absolute -bottom-1 -right-1 w-2.5 h-2.5 text-primary bg-background rounded-full" />
                </div>
              ) : (
                <CloudOff className="w-4 h-4 text-muted-foreground/50 group-hover/dl:text-blue-400 transition-colors" />
              )}
            </div>
            {expanded ? (
              <ChevronDown className="w-4 h-4 text-primary transition-transform" />
            ) : (
              <ChevronRight className="w-4 h-4 text-muted-foreground transition-transform" />
            )}
            {expanded ? (
              <FolderOpen className="w-4 h-4 text-primary" />
            ) : (
              <Folder
                className={cn(
                  "w-4 h-4 transition-colors",
                  isHovered ? "text-primary" : "text-muted-foreground"
                )}
              />
            )}
          </div>
          <span
            className={cn(
              "text-sm font-medium transition-colors",
              expanded ? "text-foreground" : "text-muted-foreground"
            )}
            title={isMobile && node.name.length > 25 ? node.name : undefined}
          >
            {truncateFolderName(node.name)}
          </span>
          <Badge
            variant="secondary"
            className="ml-auto text-xs bg-primary/10 border-0"
          >
            {node.children.length}
          </Badge>
        </div>
        {expanded && (
          <div className="mt-1 space-y-1">
            {node.children.map((child) => (
              <TreeNodeComponent
                key={child.path}
                node={child}
                allFiles={allFiles}
                downloadType={downloadType}
                taskID={taskID}
                onFileChange={onFileChange}
                onFilesChange={onFilesChange}
                onSubtitleTransferSuccess={onSubtitleTransferSuccess}
                level={level + 1}
                defaultExpandLevel={defaultExpandLevel}
                isMobile={isMobile}
              />
            ))}
          </div>
        )}
      </div>
    );
  }

  // 文件节点
  const isMediaFile = node.file?.media;
  const isDownload = node.file?.download;
  const hasLinkFile = !!node.file?.linkFile;
  const isLibraryFile = hasLinkFile; // 已入库
  const isPendingFile = isMediaFile && !hasLinkFile; // 待入库

  const hasEpisodeInfo = downloadType === "tv" && isMediaFile;

  // 获取文件图标和样式
  const getFileIconAndStyle = (
    isMedia: boolean,
    isLibrary: boolean,
    isPending: boolean,
    download: boolean
  ) => {
    const IconComponent = isMedia ? FileVideo : File;

    let className = "w-3.5 h-3.5 ";
    if (!download) {
      className += "text-muted-foreground/40";
      return { IconComponent, className };
    }

    if (isMedia) {
      if (isLibrary) {
        className += "text-green-600 dark:text-green-400";
      } else if (isPending) {
        className += "text-amber-600 dark:text-amber-400";
      } else {
        className += "text-muted-foreground";
      }
    } else {
      className += "text-muted-foreground";
    }

    return { IconComponent, className };
  };

  const [showTransferDialog, setShowTransferDialog] = useState(false);

  // 格式化季集号码为两位数
  const formatEpisodeNumber = (num: number | undefined): string => {
    return (num || 0).toString().padStart(2, "0");
  };

  // 切换下载状态
  const toggleDownloadStatus = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (node.file) {
      onFileChange(node.file.fileName, {
        download: !node.file.download,
      });
    }
  };

  // 切换待入库状态
  const toggleMediaStatus = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (node.file) {
      onFileChange(node.file.fileName, {
        media: !node.file.media,
      });
    }
  };

  // 处理季集标签点击
  const handleEpisodeBadgeClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    setShowEpisodeDialog(true);
  };

  // 处理批量修改点击
  const handleBatchEditClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    setShowBatchDialog(true);
  };

  // 处理查看转移详情
  const handleViewTransferDetail = (e: React.MouseEvent) => {
    e.stopPropagation();
    setShowTransferDialog(true);
  };

  return (
    <>
      <style>
        {`
          @keyframes cloud-float {
            0%, 100% { transform: translateY(0); }
            50% { transform: translateY(2px); }
          }
          .animate-cloud-float {
            animation: cloud-float 2s ease-in-out infinite;
          }
        `}
      </style>
      <Card
        className={cn(
          "transition-all duration-300 anime-card my-1 group",
          isDownload
            ? "border-primary/20 bg-card/50 backdrop-blur-sm"
            : "border-transparent bg-transparent opacity-60 grayscale-[0.5]",
          isLibraryFile && isDownload && "border-green-500/30 bg-green-500/5",
          isPendingFile && isDownload && "border-amber-500/30 bg-amber-500/5"
        )}
        style={{ marginLeft: `${level * 16}px` }}
      >
        <CardContent className="p-3">
          <div className="space-y-3">
            {/* 文件名行 */}
            <div className="flex items-start gap-2">
              <div
                className="pt-1 flex-shrink-0 cursor-pointer group/dl"
                onClick={toggleDownloadStatus}
                title={isDownload ? "点击取消下载" : "点击加入下载"}
              >
                {isDownload ? (
                  <div className="p-1 rounded-full bg-blue-500/10">
                    <CloudDownload className="w-4 h-4 text-blue-500 animate-cloud-float" />
                  </div>
                ) : (
                  <div className="p-1 rounded-full hover:bg-muted transition-colors">
                    <CloudOff className="w-4 h-4 text-muted-foreground/40 group-hover/dl:text-blue-400 transition-colors" />
                  </div>
                )}
              </div>

              {/* 可点击的图标区域 */}
              <div
                className={cn(
                  "p-1.5 rounded-md flex-shrink-0 cursor-pointer transition-all",
                  "hover:scale-110 active:scale-95",
                  !isDownload
                    ? "bg-muted/5 opacity-50"
                    : isLibraryFile
                    ? "bg-green-500/20 hover:bg-green-500/30"
                    : isPendingFile
                    ? "bg-amber-500/20 hover:bg-amber-500/30"
                    : "bg-primary/10 hover:bg-primary/20"
                )}
                onClick={toggleMediaStatus}
                title={
                  !isDownload
                    ? "未选择下载 - 点击标记为媒体文件将自动开启下载"
                    : isLibraryFile
                    ? "已入库 - 点击取消标记"
                    : isPendingFile
                    ? "待入库 - 点击取消标记"
                    : "点击标记为待入库"
                }
              >
                {(() => {
                  const { IconComponent, className } = getFileIconAndStyle(
                    !!isMediaFile,
                    isLibraryFile,
                    !!isPendingFile,
                    !!isDownload
                  );
                  return <IconComponent className={className} />;
                })()}
              </div>

              <div className="flex-1 min-w-0">
                <p
                  className={cn(
                    "text-sm break-all leading-tight transition-colors",
                    !isDownload ? "text-muted-foreground/60" : "text-foreground"
                  )}
                >
                  {node.name}
                </p>
                <div className="flex items-center gap-2 mt-1.5 flex-wrap">
                  {hasEpisodeInfo && (
                    <Badge
                      variant="outline"
                      className={cn(
                        "text-xs cursor-pointer transition-all duration-200",
                        !isDownload && "opacity-40",
                        isLibraryFile
                          ? "border-green-500/30 text-green-700 dark:text-green-300 hover:bg-green-500/10 hover:border-green-500/50"
                          : "border-amber-500/30 text-amber-700 dark:text-amber-300 hover:bg-amber-500/10 hover:border-amber-500/50"
                      )}
                      onClick={handleEpisodeBadgeClick}
                      title="点击修改季集信息"
                    >
                      S{formatEpisodeNumber(node.file?.season)}E
                      {formatEpisodeNumber(node.file?.episode)}
                    </Badge>
                  )}
                  {hasEpisodeInfo && (
                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-6 px-1.5 text-xs text-blue-600 hover:text-blue-700 hover:bg-blue-500/10"
                            onClick={handleBatchEditClick}
                          >
                            <Layers className="w-3 h-3" />
                          </Button>
                        </TooltipTrigger>
                        <TooltipContent>批量设置相似文件</TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                  )}
                  {isLibraryFile && (
                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-6 px-2 text-xs text-green-600 hover:text-green-700 hover:bg-green-500/10"
                            onClick={handleViewTransferDetail}
                          >
                            <Eye className="w-3 h-3 mr-1" />
                          </Button>
                        </TooltipTrigger>
                        <TooltipContent>点击查看文件转移详情</TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                  )}
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* 季集编辑弹窗 */}
      {node.file && (
        <EpisodeEditDialog
          open={showEpisodeDialog}
          onOpenChange={setShowEpisodeDialog}
          file={node.file}
          onSave={onFileChange}
        />
      )}

      {/* 批量季集编辑弹窗 */}
      {node.file && (
        <BatchEpisodeEditDialog
          open={showBatchDialog}
          onOpenChange={setShowBatchDialog}
          taskID={taskID}
          triggerFile={node.file}
          allFiles={allFiles}
          onSave={onFilesChange}
        />
      )}

      {/* 文件转移详情弹窗 */}
      {node.file && isLibraryFile && (
        <Dialog open={showTransferDialog} onOpenChange={setShowTransferDialog}>
          <DialogContent className="max-w-2xl">
            <DialogHeader>
              <DialogTitle className="flex items-center gap-2">
                <FileVideo className="w-5 h-5 text-green-600" />
                文件转移详情
              </DialogTitle>
              <DialogDescription>查看文件的转移链接信息</DialogDescription>
            </DialogHeader>
            <div className="space-y-4">
              {/* 源文件 */}
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <File className="w-4 h-4 text-muted-foreground" />
                  <span className="text-sm font-medium text-muted-foreground">
                    源文件：
                  </span>
                </div>
                <div className="pl-6 p-3 bg-muted/50 rounded-lg">
                  <code className="text-xs break-all">
                    {node.file.fileName}
                  </code>
                </div>
              </div>

              {/* 链接文件 */}
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <CheckCircle2 className="w-4 h-4 text-green-600" />
                  <span className="text-sm font-medium text-green-600">
                    转移至：
                  </span>
                </div>
                <div className="pl-6 p-3 bg-green-500/5 border border-green-500/20 rounded-lg">
                  <code className="text-xs break-all text-green-700 dark:text-green-300">
                    {node.file.linkFile}
                  </code>
                </div>
              </div>

              {/* 季集信息 */}
              {hasEpisodeInfo && (
                <div className="space-y-2">
                  <div className="flex items-center gap-2">
                    <FileVideo className="w-4 h-4 text-muted-foreground" />
                    <span className="text-sm font-medium text-muted-foreground">
                      季集信息：
                    </span>
                  </div>
                  <div className="pl-6 flex gap-2">
                    <Badge variant="outline" className="text-xs">
                      第 {node.file.season} 季
                    </Badge>
                    <Badge variant="outline" className="text-xs">
                      第 {node.file.episode} 集
                    </Badge>
                  </div>
                </div>
              )}
            </div>
          </DialogContent>
        </Dialog>
      )}
    </>
  );
}

export function FileTree({
  files,
  downloadType,
  taskID,
  onFileChange,
  onFilesChange,
  onSubtitleTransferSuccess,
  defaultExpandLevel = 1,
}: FileTreeProps) {
  const isMobile = useMobile();
  const tree = buildTree(files);
  const [showSubtitleDialog, setShowSubtitleDialog] = useState(false);

  if (!files || files.length === 0) {
    return (
      <Card className="border-primary/20 bg-gradient-to-br from-primary/5 to-transparent">
        <CardContent className="p-8">
          <div className="flex flex-col items-center justify-center gap-3">
            <div className="p-3 rounded-full bg-primary/10">
              <Folder className="w-8 h-8 text-primary" />
            </div>
            <p className="text-sm text-muted-foreground">暂无文件信息</p>
          </div>
        </CardContent>
      </Card>
    );
  }

  const totalFilesCount = files.length;

  // 已入库：有 linkFile
  const libraryFilesCount = files.filter((f) => f.linkFile).length;

  // 待入库：media=true 但没有 linkFile
  const pendingFilesCount = files.filter(
    (f) => f.media && !f.linkFile && f.download
  ).length;

  // 按季度统计媒体文件总数
  const seasonStats = useMemo(() => {
    if (downloadType !== "tv") return [];

    const stats: Record<number, number> = {};
    files.forEach((f) => {
      if (f.media && f.download) {
        stats[f.season] = (stats[f.season] || 0) + 1;
      }
    });

    return Object.entries(stats)
      .map(([season, count]) => ({ season: parseInt(season), count }))
      .sort((a, b) => a.season - b.season);
  }, [files, downloadType]);

  const mediaLabel = downloadType === "tv" ? "番剧" : "剧场版";

  return (
    <div className="space-y-3">
      {/* 统计信息 */}
      <div className="flex items-center gap-2 flex-wrap justify-between">
        <div className="flex items-center gap-2 flex-wrap">
          <Badge
            variant="secondary"
            className="bg-primary/10 text-primary border-0 flex items-center gap-1.5"
          >
            <File className="w-3 h-3" />
            {!isMobile && "总文件: "}
            {totalFilesCount}
          </Badge>
          <Badge
            variant="secondary"
            className="bg-amber-500/10 text-amber-700 dark:text-amber-300 border-0 flex items-center gap-1.5"
          >
            <FileVideo className="w-3 h-3" />
            {!isMobile && `待入库${mediaLabel}: `}
            {pendingFilesCount}
          </Badge>
          <Badge
            variant="secondary"
            className="bg-green-500/10 text-green-700 dark:text-green-300 border-0 flex items-center gap-1.5"
          >
            <FileVideo className="w-3 h-3" />
            {!isMobile && `已入库${mediaLabel}: `}
            {libraryFilesCount}
          </Badge>
          {seasonStats.map(({ season, count }) => (
            <Badge
              key={season}
              variant="secondary"
              className="bg-blue-500/10 text-blue-700 dark:text-blue-300 border-0 flex items-center gap-1.5"
            >
              <FileVideo className="w-3 h-3" />S
              {String(season).padStart(2, "0")}: {count}
            </Badge>
          ))}
        </div>

        {/* 转移字幕按钮 */}
        <Button
          variant="outline"
          size="sm"
          className={cn(
            "h-7 border-blue-500/30 hover:border-blue-500/50",
            "text-blue-600 hover:text-blue-700 hover:bg-blue-500/10",
            "flex items-center gap-1.5 transition-all",
            isMobile ? "px-2" : "px-3"
          )}
          onClick={() => setShowSubtitleDialog(true)}
        >
          <FileVideo className="w-3 h-3" />
          {!isMobile && "转移字幕"}
        </Button>
      </div>

      {/* 文件树 */}
      <div className="border border-primary/20 rounded-xl p-3 max-h-96 overflow-auto scrollbar-hide bg-gradient-to-br from-primary/5 to-transparent">
        <div className="space-y-1">
          {tree.children.map((child) => (
            <TreeNodeComponent
              key={child.path}
              node={child}
              allFiles={files}
              downloadType={downloadType}
              taskID={taskID}
              onFileChange={onFileChange}
              onFilesChange={onFilesChange}
              onSubtitleTransferSuccess={onSubtitleTransferSuccess}
              defaultExpandLevel={defaultExpandLevel}
              isMobile={isMobile}
            />
          ))}
        </div>
      </div>

      {/* 字幕转移对话框 */}
      <SubtitleTransferDialog
        open={showSubtitleDialog}
        onOpenChange={setShowSubtitleDialog}
        taskID={taskID}
        torrentFiles={files}
        downloadType={downloadType}
        onSuccess={onSubtitleTransferSuccess}
      />
    </div>
  );
}
