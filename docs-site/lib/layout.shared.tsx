import { i18n } from '@/lib/i18n';
import { localeDocsPrefix } from '@/lib/locale-routing';
import type { Locale } from '@/lib/i18n';
import { defineI18nUI } from 'fumadocs-ui/i18n';
import type { BaseLayoutProps } from 'fumadocs-ui/layouts/shared';

const basePath = process.env.GITHUB_PAGES === 'true' ? '/hyper-fast-builder' : '';

export const i18nUI = defineI18nUI(i18n, {
  translations: {
    en: { displayName: 'English', chooseLanguage: 'Choose a language' },
    fr: {
      displayName: 'Français',
      search: 'Rechercher',
      chooseLanguage: 'Choisir une langue',
    },
    de: {
      displayName: 'Deutsch',
      search: 'Suchen',
      chooseLanguage: 'Sprache wählen',
    },
    es: {
      displayName: 'Español',
      search: 'Buscar',
      chooseLanguage: 'Elegir idioma',
    },
  },
});

export function baseOptions(locale: Locale): BaseLayoutProps {
  const docsHome = `${localeDocsPrefix(locale, basePath)}/`;
  return {
    i18n: true,
    nav: {
      title: 'AgentFlow',
      url: docsHome,
    },
  };
}
