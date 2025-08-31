import { useState, useCallback } from "react";
import { RefreshCw, Plus, Sparkles } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { AddMagnetDialog } from "./add-magnet-dialog";
import { MagnetTaskList } from "./magnet-task-list";

export default function MagnetDownload() {
  const [refreshSuccess, setRefreshSuccess] = useState(false);
  const [addDialogOpen, setAddDialogOpen] = useState(false);
  const [refreshTrigger, setRefreshTrigger] = useState(false);
  const [openTaskID, setOpenTaskID] = useState<string | null>(null);

  // 触发刷新
  const handleRefresh = useCallback(() => {
    setRefreshSuccess(false);
    setRefreshTrigger(true);
  }, []);

  // 刷新完成回调
  const handleRefreshComplete = useCallback(() => {
    setRefreshSuccess(true);
    setRefreshTrigger(false);
    // 短暂显示刷新成功状态
    setTimeout(() => setRefreshSuccess(false), 500);
  }, []);

  // 处理添加任务成功
  const handleAddSuccess = useCallback(
    (taskID: string) => {
      // 先刷新列表
      handleRefresh();
      // 然后打开详情对话框
      setOpenTaskID(taskID);
      // 重置 openTaskID，以便下次可以打开同一个任务
      setTimeout(() => setOpenTaskID(null), 100);
    },
    [handleRefresh]
  );

  return (
    <div className="flex flex-col h-[calc(100dvh-4rem)] hide-scrollbar">
      <div className="flex-none space-y-6 pb-6">
        <div className="flex flex-row items-center justify-between gap-4">
          <h1 className="text-2xl xs:text-3xl font-bold anime-gradient-text flex items-center gap-2">
            <Sparkles className="h-4 w-4 xs:h-6 xs:w-6 text-primary animate-pulse" />
            <span className="flex flex-row">磁力下载</span>
          </h1>
          <div className="flex flex-wrap gap-2 justify-end">
            <Button
              className="rounded-xl anime-button bg-gradient-to-r from-primary to-blue-500 hover:opacity-90 p-2 sm:px-3 sm:py-2 aspect-square sm:aspect-auto"
              onClick={() => setAddDialogOpen(true)}
            >
              <Plus className="icon-button" />
              <span className="hidden sm:inline">添加磁力任务</span>
            </Button>
            <AddMagnetDialog
              open={addDialogOpen}
              onOpenChange={(open) => {
                setAddDialogOpen(open);
                if (!open) {
                  handleRefresh();
                }
              }}
              onSuccess={handleAddSuccess}
            />

            <Button
              variant="outline"
              className={cn(
                "rounded-xl anime-button p-2 sm:px-3 sm:py-2 aspect-square sm:aspect-auto",
                refreshSuccess &&
                  "bg-green-100 border-green-500 text-green-600 transition-colors duration-500"
              )}
              onClick={handleRefresh}
            >
              <RefreshCw
                className={cn(
                  "icon-button",
                  refreshSuccess &&
                    "text-green-500 animate-[spin_1s_ease-in-out]"
                )}
              />
              <span className="hidden sm:inline">
                {refreshSuccess ? "刷新成功" : "刷新"}
              </span>
            </Button>
          </div>
        </div>
      </div>

      {/* 磁力任务列表 */}
      <div className="flex-grow border rounded-xl bg-background/50 overflow-hidden flex flex-col">
        <div className="p-4 flex-1 flex flex-col min-h-0">
          <MagnetTaskList
            refresh={refreshTrigger}
            onRefreshComplete={handleRefreshComplete}
            pauseRefresh={addDialogOpen}
            openTaskID={openTaskID}
          />
        </div>
      </div>
    </div>
  );
}
