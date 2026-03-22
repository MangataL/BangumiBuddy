import type { ScraperTask } from "@/api/config";

function normalizeBangumiName(bangumiName: string) {
  return bangumiName.trim();
}

export function getScraperTaskBangumiNames(tasks: ScraperTask[]) {
  const bangumiNames = new Set<string>();

  for (const task of tasks) {
    const bangumiName = normalizeBangumiName(task.bangumiName);
    if (!bangumiName) {
      continue;
    }
    bangumiNames.add(bangumiName);
  }

  return Array.from(bangumiNames);
}

export function filterScraperTasksByBangumiName(
  tasks: ScraperTask[],
  bangumiName: string
) {
  const normalizedBangumiName = normalizeBangumiName(bangumiName);
  if (!normalizedBangumiName) {
    return tasks;
  }

  return tasks.filter(
    (task) => normalizeBangumiName(task.bangumiName) === normalizedBangumiName
  );
}
