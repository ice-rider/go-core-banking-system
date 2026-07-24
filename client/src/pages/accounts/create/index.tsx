import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { CreateAccountForm } from '@/features/create-account';
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card';
import { ArrowLeft } from 'lucide-react';

export function AccountCreatePage() {
  const { t } = useTranslation();

  return (
    <div className="container mx-auto p-6">
      <Link to="/accounts" className="inline-flex items-center text-sm text-muted-foreground hover:text-foreground mb-4">
        <ArrowLeft className="mr-2 h-4 w-4" />
        {t('common.back')}
      </Link>
      <Card className="max-w-md">
        <CardHeader>
          <CardTitle>{t('accounts.create')}</CardTitle>
        </CardHeader>
        <CardContent>
          <CreateAccountForm />
        </CardContent>
      </Card>
    </div>
  );
}
