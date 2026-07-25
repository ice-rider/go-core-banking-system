import { Badge } from '@/shared/ui/badge';

interface AccountStatusBadgeProps {
  status: 'ACTIVE' | 'BLOCKED' | 'CLOSED';
}

const variantMap = {
  ACTIVE: 'default' as const,
  BLOCKED: 'destructive' as const,
  CLOSED: 'secondary' as const,
};

export function AccountStatusBadge({ status }: AccountStatusBadgeProps) {
  return <Badge variant={variantMap[status]}>{status}</Badge>;
}
