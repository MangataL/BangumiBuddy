import type { ReleaseGroupCandidate } from "@/api/discovery";

export function getDefaultSubscriptionPriority(
  group: ReleaseGroupCandidate
): number {
  return Math.max(1, group.previousMaxPriority ?? 1);
}
