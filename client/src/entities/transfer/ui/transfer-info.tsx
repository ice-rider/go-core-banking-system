import { Badge } from '@/shared/ui/badge';
import { formatCurrency, formatDate } from '@/shared/lib/utils';
import type { Transfer } from '../model';

interface TransferInfoProps {
  transfer: Transfer;
}

export function TransferInfo({ transfer }: TransferInfoProps) {
  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <span className="text-sm text-muted-foreground">From</span>
        <span className="font-medium">{transfer.from_account_id}</span>
      </div>
      <div className="flex items-center justify-between">
        <span className="text-sm text-muted-foreground">To</span>
        <span className="font-medium">{transfer.to_account_id}</span>
      </div>
      <div className="flex items-center justify-between">
        <span className="text-sm text-muted-foreground">Amount</span>
        <span className="font-medium">{formatCurrency(transfer.amount)}</span>
      </div>
      <div className="flex items-center justify-between">
        <span className="text-sm text-muted-foreground">Status</span>
        <Badge variant={transfer.status === 'COMPLETED' ? 'default' : 'destructive'}>
          {transfer.status}
        </Badge>
      </div>
      <div className="flex items-center justify-between">
        <span className="text-sm text-muted-foreground">Created</span>
        <span className="font-medium">{formatDate(transfer.created_at)}</span>
      </div>
    </div>
  );
}
