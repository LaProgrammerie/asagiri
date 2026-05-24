import type { CalloutType } from '@/components/callout';
import type { Locale } from '@/lib/i18n';

const LABELS: Record<Locale, Record<CalloutType, string>> = {
  en: {
    note: 'Note',
    warning: 'Warning',
    cost: 'Cost',
    security: 'Security',
    'local-first': 'Local-first',
    experimental: 'Experimental',
  },
  fr: {
    note: 'Note',
    warning: 'Avertissement',
    cost: 'Coût',
    security: 'Sécurité',
    'local-first': 'Local d\'abord',
    experimental: 'Experimental',
  },
  de: {
    note: 'Hinweis',
    warning: 'Warnung',
    cost: 'Kosten',
    security: 'Sicherheit',
    'local-first': 'Local-first',
    experimental: 'Experimental',
  },
  es: {
    note: 'Nota',
    warning: 'Advertencia',
    cost: 'Coste',
    security: 'Seguridad',
    'local-first': 'Local primero',
    experimental: 'Experimental',
  },
};

export function calloutLabel(locale: Locale, type: CalloutType): string {
  return LABELS[locale][type];
}
