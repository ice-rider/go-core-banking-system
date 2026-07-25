import { useTranslation } from 'react-i18next';
import { Globe } from 'lucide-react';
import { Button } from '@/shared/ui/button';

export function LanguageSwitcher() {
  const { i18n } = useTranslation();

  const toggleLanguage = () => {
    const newLang = i18n.language === 'en' ? 'ru' : 'en';
    i18n.changeLanguage(newLang);
  };

  return (
    <Button variant="ghost" size="sm" onClick={toggleLanguage}>
      <Globe className="h-4 w-4 mr-2" />
      {i18n.language === 'en' ? 'RU' : 'EN'}
    </Button>
  );
}
