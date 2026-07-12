export function extractBangumiNameFromParseError(message: string): string {
  const match = message.match(/番剧名称:\s*(.+)$/);
  return match?.[1]?.trim() ?? "";
}

export function isTMDBMatchFailure(message: string): boolean {
  return message.includes("未搜索到番剧");
}

export function resolveDiscoveryBangumiName(options: {
  bangumiID: string;
  detail?: { mikanBangumiID: string; name: string } | null;
  bangumis?: Array<{ mikanBangumiID: string; name: string }>;
  searchBangumis?: Array<{ mikanBangumiID: string; name: string }>;
  errorMessage?: string;
}): string {
  if (options.detail?.mikanBangumiID === options.bangumiID) {
    return options.detail.name;
  }

  const fromSeason = options.bangumis?.find(
    (item) => item.mikanBangumiID === options.bangumiID
  )?.name;
  if (fromSeason) return fromSeason;

  const fromSearch = options.searchBangumis?.find(
    (item) => item.mikanBangumiID === options.bangumiID
  )?.name;
  if (fromSearch) return fromSearch;

  if (options.errorMessage) {
    return extractBangumiNameFromParseError(options.errorMessage);
  }

  return "";
}
