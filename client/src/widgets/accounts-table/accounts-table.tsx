import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/ui/table';
import { AccountStatusBadge } from '@/entities/account/ui/account-status-badge';
import { formatCurrency } from '@/shared/lib/utils';
import { Link } from 'react-router-dom';
import type { Account } from '@/entities/account';

interface AccountsTableProps {
  accounts: Account[];
}

export function AccountsTable({ accounts }: AccountsTableProps) {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Owner</TableHead>
          <TableHead>Balance</TableHead>
          <TableHead>Status</TableHead>
          <TableHead>Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {accounts.map((account) => (
          <TableRow key={account.id}>
            <TableCell>{account.owner_name}</TableCell>
            <TableCell>{formatCurrency(account.balance)}</TableCell>
            <TableCell>
              <AccountStatusBadge status={account.status} />
            </TableCell>
            <TableCell>
              <Link
                to={`/accounts/${account.id}`}
                className="text-primary hover:underline"
              >
                View
              </Link>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
