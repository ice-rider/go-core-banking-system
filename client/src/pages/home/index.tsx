import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Button } from '@/shared/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card';
import { Plus } from 'lucide-react';

export function HomePage() {
  const { t } = useTranslation();

  return (
    <div className="container mx-auto p-6">
      <h1 className="text-3xl font-bold mb-6">{t('app.title')}</h1>
      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>{t('nav.accounts')}</CardTitle>
          </CardHeader>
          <CardContent>
            <Link to="/accounts">
              <Button>
                <Plus className="mr-2 h-4 w-4" />
                {t('accounts.list')}
              </Button>
            </Link>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>{t('nav.transfers')}</CardTitle>
          </CardHeader>
          <CardContent>
            <Link to="/transfers/new">
              <Button>
                <Plus className="mr-2 h-4 w-4" />
                {t('transfers.create')}
              </Button>
            </Link>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
