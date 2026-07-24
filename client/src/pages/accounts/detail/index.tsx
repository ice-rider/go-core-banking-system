import { useParams, Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useAccount } from '@/entities/account/api';
import { AccountInfo } from '@/entities/account/ui/account-info';
import { BlockAccountButton } from '@/features/block-account';
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card';
import { Loading } from '@/shared/ui/loading/loading';
import { ArrowLeft } from 'lucide-react';

export function AccountDetailPage() {
  const { id } = useParams<{ id: string }>();
  const { t } = useTranslation();
  const { data: account, isLoading, error } = useAccount(id!);

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

  if (!account) {
    return null;
  }

  return (
    <div className="container mx-auto p-6">
      <Link to="/accounts" className="inline-flex items-center text-sm text-muted-foreground hover:text-foreground mb-4">
        <ArrowLeft className="mr-2 h-4 w-4" />
        {t('common.back')}
      </Link>
      <Card className="max-w-md">
        <CardHeader>
          <CardTitle>{account.owner_name}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <AccountInfo account={account} />
          <div className="flex gap-2">
            <BlockAccountButton accountId={account.id} status={account.status} />
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
