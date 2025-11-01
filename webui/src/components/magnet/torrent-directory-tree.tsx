import { useState } from "react";
import { Folder, FolderOpen, ChevronRight, ChevronDown } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { TorrentFile } from "@/api/magnet";
import { ScrollArea } from "@/components/ui/scroll-area";

interface TreeNode {
  name: string;
  path: string;
  isDirectory: boolean;
  children: TreeNode[];
}

interface TorrentDirectoryTreeProps {
  files: TorrentFile[];
  selectedPath: string;
  onSelect: (path: string) => void;
  className?: string;
}

// 构建目录树（只包含目录）
function buildDirectoryTree(files: TorrentFile[]): TreeNode {
  const root: TreeNode = {
    name: "根目录",
    path: "",
    isDirectory: true,
    children: [],
  };

  // 收集所有目录路径
  const dirPaths = new Set<string>();

  files.forEach((file) => {
    const parts = file.fileName.split("/");
    // 添加所有父目录
    for (let i = 1; i < parts.length; i++) {
      const dirPath = parts.slice(0, i).join("/");
      dirPaths.add(dirPath);
    }
  });

  // 构建目录树
  const sortedPaths = Array.from(dirPaths).sort();
  sortedPaths.forEach((dirPath) => {
    const parts = dirPath.split("/");
    let currentNode = root;

    parts.forEach((part, index) => {
      const path = parts.slice(0, index + 1).join("/");
      let childNode = currentNode.children.find((child) => child.name === part);

      if (!childNode) {
        childNode = {
          name: part,
          path,
          isDirectory: true,
          children: [],
        };
        currentNode.children.push(childNode);
      }

      currentNode = childNode;
    });
  });

  return root;
}

// 目录节点组件
function DirectoryNodeComponent({
  node,
  level = 0,
  selectedPath,
  onSelect,
}: {
  node: TreeNode;
  level?: number;
  selectedPath: string;
  onSelect: (path: string) => void;
}) {
  const [expanded, setExpanded] = useState(level === 0); // 根节点默认展开
  const isSelected = selectedPath === node.path;

  return (
    <div className="select-none">
      <div
        className={cn(
          "flex items-center gap-2 py-1.5 px-2 rounded-lg cursor-pointer transition-all",
          "hover:bg-accent",
          isSelected && "bg-blue-500/10 border border-blue-500/30"
        )}
        style={{ marginLeft: `${level * 12}px` }}
        onClick={() => onSelect(node.path)}
      >
        <div
          className="flex items-center gap-1 flex-shrink-0"
          onClick={(e) => {
            if (node.children.length > 0) {
              e.stopPropagation();
              setExpanded(!expanded);
            }
          }}
        >
          {node.children.length > 0 ? (
            expanded ? (
              <ChevronDown className="w-3 h-3 text-muted-foreground" />
            ) : (
              <ChevronRight className="w-3 h-3 text-muted-foreground" />
            )
          ) : (
            <div className="w-3" /> // 占位
          )}
          {expanded ? (
            <FolderOpen className="w-3.5 h-3.5 text-blue-600" />
          ) : (
            <Folder className="w-3.5 h-3.5 text-blue-600" />
          )}
        </div>
        <span
          className={cn("text-sm flex-1 truncate", isSelected && "font-medium")}
        >
          {node.name}
        </span>
        {node.children.length > 0 && (
          <Badge variant="secondary" className="text-xs h-4 px-1">
            {node.children.length}
          </Badge>
        )}
      </div>
      {expanded && node.children.length > 0 && (
        <div className="mt-0.5">
          {node.children.map((child) => (
            <DirectoryNodeComponent
              key={child.path}
              node={child}
              level={level + 1}
              selectedPath={selectedPath}
              onSelect={onSelect}
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
}: TorrentDirectoryTreeProps) {
  const tree = buildDirectoryTree(files);

  return (
    <ScrollArea
      className={cn("rounded-lg border bg-muted/20", className ?? "h-64")}
    >
      <div className="p-2">
        <DirectoryNodeComponent
          node={tree}
          selectedPath={selectedPath}
          onSelect={onSelect}
        />
      </div>
    </ScrollArea>
  );
}
