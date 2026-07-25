import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useAccounts } from '@/entities/account/api';
import { AccountsTable } from '@/widgets/accounts-table';
import { Button } from '@/shared/ui/button';
import { Loading } from '@/shared/ui/loading/loading';
import { Plus } from 'lucide-react';

export function AccountsListPage() {
  const { t } = useTranslation();
  const { data: accounts, isLoading, error } = useAccounts();

  if (isLoading) {
    return <Loading className="min-h-screen" />;
  }

  if (error) {
    return (
      <div className="container mx-auto p-6">
        <p className="text-destructive">{t('common.error')}: {error.message}</p>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-3xl font-bold">{t('accounts.list')}</h1>
        <Link to="/accounts/new">
          <Button>
            <Plus className="mr-2 h-4 w-4" />
            {t('accounts.create')}
          </Button>
        </Link>
      </div>
      {accounts && <AccountsTable accounts={accounts} />}
    </div>
  );
}
