import { Badge } from '@/shared/ui/badge';
import { formatCurrency } from '@/shared/lib/utils';
import type { Account } from '../model';

interface AccountInfoProps {
  account: Account;
}

export function AccountInfo({ account }: AccountInfoProps) {
  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <span className="text-sm text-muted-foreground">Owner</span>
        <span className="font-medium">{account.owner_name}</span>
      </div>
      <div className="flex items-center justify-between">
        <span className="text-sm text-muted-foreground">Balance</span>
        <span className="font-medium">{formatCurrency(account.balance)}</span>
      </div>
      <div className="flex items-center justify-between">
        <span className="text-sm text-muted-foreground">Status</span>
        <Badge variant={account.status === 'ACTIVE' ? 'default' : 'destructive'}>
          {account.status}
        </Badge>
      </div>
    </div>
  );
}
