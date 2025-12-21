import { useState } from "react";
import {
  Folder,
  FolderOpen,
  ChevronRight,
  ChevronDown,
  Video,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { TorrentFile } from "@/api/magnet";
import { ScrollArea } from "@/components/ui/scroll-area";

interface TreeNode {
  name: string;
  path: string;
  isDirectory: boolean;
  isMedia?: boolean;
  children: TreeNode[];
}

interface TorrentDirectoryTreeProps {
  files: TorrentFile[];
  selectedPath: string;
  onSelect: (path: string) => void;
  className?: string;
  isMobile?: boolean;
}

// 构建目录树（包含目录和媒体文件）
function buildTree(files: TorrentFile[]): TreeNode {
  const root: TreeNode = {
    name: "根目录",
    path: "",
    isDirectory: true,
    children: [],
  };

  files.forEach((file) => {
    const parts = file.fileName.split("/");
    let currentNode = root;

    parts.forEach((part, index) => {
      const isLast = index === parts.length - 1;
      const path = parts.slice(0, index + 1).join("/");

      let childNode = currentNode.children.find((child) => child.name === part);

      if (!childNode) {
        if (isLast) {
          if (file.media) {
            childNode = {
              name: part,
              path,
              isDirectory: false,
              isMedia: true,
              children: [],
            };
            currentNode.children.push(childNode);
          }
        } else {
          childNode = {
            name: part,
            path,
            isDirectory: true,
            children: [],
          };
          currentNode.children.push(childNode);
        }
      }

      if (childNode) {
        currentNode = childNode;
      }
    });
  });

  return root;
}

// 节点组件
function NodeComponent({
  node,
  level = 0,
  selectedPath,
  onSelect,
  isMobile = false,
}: {
  node: TreeNode;
  level?: number;
  selectedPath: string;
  onSelect: (path: string) => void;
  isMobile?: boolean;
}) {
  const [expanded, setExpanded] = useState(level === 0);
  const isSelected = selectedPath === node.path;

  const hasChildren = node.children.length > 0;

  return (
    <div className="select-none">
      <div
        className={cn(
          "flex items-center gap-2 rounded-lg cursor-pointer transition-all",
          isMobile ? "py-1 px-1.5" : "py-1.5 px-2",
          "hover:bg-accent",
          isSelected && "bg-blue-500/10 border border-blue-500/30"
        )}
        style={{ marginLeft: `${level * (isMobile ? 8 : 12)}px` }}
        onClick={() => onSelect(node.path)}
      >
        <div
          className="flex items-center gap-1 flex-shrink-0"
          onClick={(e) => {
            if (hasChildren) {
              e.stopPropagation();
              setExpanded(!expanded);
            }
          }}
        >
          {hasChildren ? (
            expanded ? (
              <ChevronDown
                className={cn(
                  isMobile ? "w-2.5 h-2.5" : "w-3 h-3",
                  "text-muted-foreground"
                )}
              />
            ) : (
              <ChevronRight
                className={cn(
                  isMobile ? "w-2.5 h-2.5" : "w-3 h-3",
                  "text-muted-foreground"
                )}
              />
            )
          ) : (
            <div className={isMobile ? "w-2.5" : "w-3"} />
          )}
          {node.isDirectory ? (
            expanded ? (
              <FolderOpen
                className={cn(
                  isMobile ? "w-3 h-3" : "w-3.5 h-3.5",
                  "text-blue-600 fill-blue-600/10"
                )}
              />
            ) : (
              <Folder
                className={cn(
                  isMobile ? "w-3 h-3" : "w-3.5 h-3.5",
                  "text-blue-600 fill-blue-600/10"
                )}
              />
            )
          ) : (
            <Video
              className={cn(
                isMobile ? "w-3 h-3" : "w-3.5 h-3.5",
                "text-purple-600 fill-purple-600/10"
              )}
            />
          )}
        </div>
        <span
          className={cn(
            "text-sm flex-1 truncate",
            isSelected && "font-medium text-blue-600",
            !node.isDirectory && "text-muted-foreground"
          )}
        >
          {node.name}
        </span>
        {node.isDirectory && hasChildren && (
          <Badge
            variant="secondary"
            className="text-[10px] h-4 px-1 opacity-50"
          >
            {node.children.length}
          </Badge>
        )}
      </div>
      {expanded && hasChildren && (
        <div className="mt-0.5">
          {node.children
            .sort((a, b) => {
              if (a.isDirectory === b.isDirectory)
                return a.name.localeCompare(b.name);
              return a.isDirectory ? -1 : 1;
            })
            .map((child) => (
              <NodeComponent
                key={child.path}
                node={child}
                level={level + 1}
                selectedPath={selectedPath}
                onSelect={onSelect}
                isMobile={isMobile}
              />
            ))}
        </div>
      )}
    </div>
  );
}

export function TorrentDirectoryTree({
  files,
  selectedPath,
  onSelect,
  className,
  isMobile = false,
}: TorrentDirectoryTreeProps) {
  const tree = buildTree(files);

  return (
    <ScrollArea
      className={cn("rounded-xl border bg-muted/5", className ?? "h-64")}
    >
      <div className="p-2">
        <NodeComponent
          node={tree}
          selectedPath={selectedPath}
          onSelect={onSelect}
          isMobile={isMobile}
        />
      </div>
    </ScrollArea>
  );
}
