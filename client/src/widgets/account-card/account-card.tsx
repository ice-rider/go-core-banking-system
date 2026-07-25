import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card';
import { AccountStatusBadge } from '@/entities/account/ui/account-status-badge';
import { formatCurrency } from '@/shared/lib/utils';
import type { Account } from '@/entities/account';

interface AccountCardProps {
  account: Account;
}

export function AccountCard({ account }: AccountCardProps) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">{account.owner_name}</CardTitle>
        <AccountStatusBadge status={account.status} />
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{formatCurrency(account.balance)}</div>
        <p className="text-xs text-muted-foreground">{account.id}</p>
      </CardContent>
    </Card>
  );
}
