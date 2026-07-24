import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Landmark } from 'lucide-react';
import { Button } from '@/shared/ui/button';
import { LanguageSwitcher } from '@/shared/ui/language-switcher/language-switcher';

export function Navbar() {
  const { t } = useTranslation();

  return (
    <nav className="border-b bg-background">
      <div className="container mx-auto flex h-14 items-center justify-between px-4">
        <Link to="/" className="flex items-center gap-2 font-semibold">
          <Landmark className="h-6 w-6" />
          <span>{t('app.title')}</span>
        </Link>
        <div className="flex items-center gap-4">
          <Link to="/accounts">
            <Button variant="ghost" size="sm">{t('nav.accounts')}</Button>
          </Link>
          <Link to="/transfers/new">
            <Button variant="ghost" size="sm">{t('nav.transfers')}</Button>
          </Link>
          <LanguageSwitcher />
        </div>
      </div>
    </nav>
  );
}
