import { useParams, Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useTransfer } from '@/entities/transfer/api';
import { TransferInfo } from '@/entities/transfer/ui/transfer-info';
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card';
import { Loading } from '@/shared/ui/loading/loading';
import { ArrowLeft } from 'lucide-react';

export function TransferDetailPage() {
  const { id } = useParams<{ id: string }>();
  const { t } = useTranslation();
  const { data: transfer, isLoading, error } = useTransfer(id!);

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

  if (!transfer) {
    return null;
  }

  return (
    <div className="container mx-auto p-6">
      <Link to="/" className="inline-flex items-center text-sm text-muted-foreground hover:text-foreground mb-4">
        <ArrowLeft className="mr-2 h-4 w-4" />
        {t('common.back')}
      </Link>
      <Card className="max-w-md">
        <CardHeader>
          <CardTitle>Transfer Details</CardTitle>
        </CardHeader>
        <CardContent>
          <TransferInfo transfer={transfer} />
        </CardContent>
      </Card>
    </div>
  );
}
