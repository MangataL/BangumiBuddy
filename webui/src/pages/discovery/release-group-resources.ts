import type {
  ReleaseGroupCandidate,
  ResourceCandidate,
} from "@/api/discovery";

export function getResourcesForReleaseGroup(
  group: ReleaseGroupCandidate,
  resources: ResourceCandidate[]
): ResourceCandidate[] {
  const groupName = normalizeReleaseGroupName(group.name);
  return resources.filter((resource) => {
    if (resource.mikanSubgroupID) {
      return resource.mikanSubgroupID === group.mikanSubgroupID;
    }
    const resourceGroup = normalizeReleaseGroupName(resource.releaseGroup);
    return Boolean(
      resourceGroup &&
        groupName &&
        (resourceGroup === groupName ||
          resourceGroup.includes(groupName) ||
          groupName.includes(resourceGroup))
    );
  });
}

function normalizeReleaseGroupName(name: string) {
  return name
    .trim()
    .toLocaleLowerCase()
    .replace(/[\s【】（）·・&＆+]/g, "")
    .replaceAll("[", "")
    .replaceAll("]", "")
    .replaceAll("(", "")
    .replaceAll(")", "");
}
