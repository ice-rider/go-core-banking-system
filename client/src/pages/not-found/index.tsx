import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Button } from '@/shared/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card';
import { Home } from 'lucide-react';

export function NotFoundPage() {
  const { t } = useTranslation();

  return (
    <div className="container mx-auto p-6 flex items-center justify-center min-h-screen">
      <Card className="max-w-md text-center">
        <CardHeader>
          <CardTitle className="text-6xl">404</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-muted-foreground">Page not found</p>
          <Link to="/">
            <Button>
              <Home className="mr-2 h-4 w-4" />
              {t('nav.home')}
            </Button>
          </Link>
        </CardContent>
      </Card>
    </div>
  );
}
