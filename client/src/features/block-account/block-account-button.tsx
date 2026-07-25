import { useTranslation } from 'react-i18next';
import { Button } from '@/shared/ui/button';
import { useBlockAccount, useUnblockAccount } from '@/entities/account/api';

interface BlockAccountButtonProps {
  accountId: string;
  status: 'ACTIVE' | 'BLOCKED' | 'CLOSED';
}

export function BlockAccountButton({ accountId, status }: BlockAccountButtonProps) {
  const { t } = useTranslation();
  const blockAccount = useBlockAccount();
  const unblockAccount = useUnblockAccount();

  const handleClick = () => {
    if (status === 'ACTIVE') {
      blockAccount.mutate(accountId);
    } else if (status === 'BLOCKED') {
      unblockAccount.mutate(accountId);
    }
  };

  const isLoading = blockAccount.isPending || unblockAccount.isPending;

  if (status === 'CLOSED') {
    return null;
  }

  return (
    <Button
      variant={status === 'ACTIVE' ? 'destructive' : 'default'}
      onClick={handleClick}
      disabled={isLoading}
    >
      {status === 'ACTIVE' ? t('accounts.block') : t('accounts.unblock')}
    </Button>
  );
}
