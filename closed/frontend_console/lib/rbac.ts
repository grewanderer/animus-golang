export type EffectiveRole = 'viewer' | 'editor' | 'admin' | 'unknown';

const roleRank: Record<string, number> = {
  viewer: 1,
  editor: 2,
  admin: 3,
};

export function deriveEffectiveRole(roles: string[] | undefined): EffectiveRole {
  if (!roles || roles.length === 0) {
    return 'unknown';
  }
  let best: EffectiveRole = 'unknown';
  let score = 0;
  for (const raw of roles) {
    const role = raw.trim().toLowerCase();
    const rank = roleRank[role] ?? 0;
    if (rank > score) {
      score = rank;
      best = role as EffectiveRole;
    }
  }
  return best;
}

export function roleLabel(role: EffectiveRole): string {
  switch (role) {
    case 'admin':
      return 'Администратор';
    case 'editor':
      return 'Оператор';
    case 'viewer':
      return 'Наблюдатель';
    default:
      return 'Не определена';
  }
}
