import type { Locale } from '@/lib/i18n';
import DE from 'country-flag-icons/react/3x2/DE';
import ES from 'country-flag-icons/react/3x2/ES';
import FR from 'country-flag-icons/react/3x2/FR';
import GB from 'country-flag-icons/react/3x2/GB';
import type { FC, SVGProps } from 'react';

type FlagIcon = FC<SVGProps<SVGSVGElement>>;

const localeFlags: Record<Locale, FlagIcon> = {
  en: GB as FlagIcon,
  fr: FR as FlagIcon,
  de: DE as FlagIcon,
  es: ES as FlagIcon,
};

export function LocaleFlag({
  locale,
  className,
}: {
  locale: string;
  className?: string;
}) {
  const Flag = localeFlags[locale as Locale] ?? localeFlags.en;
  return (
    <Flag
      className={className}
      aria-hidden
      style={{ display: 'block' }}
    />
  );
}
