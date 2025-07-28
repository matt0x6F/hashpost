// Community type configuration
export const COMMUNITY_CONFIG = {
  t: { name: 'Topical', color: 'bg-blue-100 text-blue-800' },
  g: { name: 'Geographic', color: 'bg-green-100 text-green-800' },
  b: { name: 'Branded', color: 'bg-purple-100 text-purple-800' },
  c: { name: 'Creator', color: 'bg-orange-100 text-orange-800' },
  h: { name: 'General', color: 'bg-gray-100 text-gray-800' },
} as const;

export type CommunityType = keyof typeof COMMUNITY_CONFIG;

export function getCommunityConfig(communityType: string) {
  return COMMUNITY_CONFIG[communityType as CommunityType] || COMMUNITY_CONFIG.h;
}

export function parseSubforumPath(fullPath: string): { communityType: CommunityType; subforumName: string } | null {
  const parts = fullPath.split('/');
  if (parts.length !== 2) return null;
  
  const [communityType, subforumName] = parts;
  if (!COMMUNITY_CONFIG[communityType as CommunityType]) return null;
  
  return {
    communityType: communityType as CommunityType,
    subforumName
  };
} 