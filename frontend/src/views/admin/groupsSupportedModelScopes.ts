export const normalizeSupportedModelScopesForPlatform = (
  platform: string,
  scopes: string[] | undefined,
): string[] => {
  if (platform === "antigravity") return scopes ?? [];
  // Video groups may carry the "video" scope to enable the Videos API on a
  // non-video-platform group; for video-platform groups it is implied.
  if (platform === "video") {
    return (scopes ?? []).filter((s) => s === "video");
  }
  return [];
};
