import { Badge } from '@/shared/ui/badge';

interface TransferStatusBadgeProps {
  status: 'PENDING' | 'COMPLETED' | 'FAILED';
}

const variantMap = {
  PENDING: 'secondary' as const,
  COMPLETED: 'default' as const,
  FAILED: 'destructive' as const,
};

export function TransferStatusBadge({ status }: TransferStatusBadgeProps) {
  return <Badge variant={variantMap[status]}>{status}</Badge>;
}
