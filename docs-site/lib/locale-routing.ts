import { defaultLocale, isLocale, type Locale } from '@/lib/i18n';

const nonDefaultLocales = ['fr', 'de', 'es'] as const;

export function parseDocsSlug(slug: string[] | undefined): {
  locale: Locale;
  pageSlug: string[];
} {
  const parts = slug ?? [];
  if (parts.length > 0 && isLocale(parts[0]) && parts[0] !== defaultLocale) {
    return { locale: parts[0], pageSlug: parts.slice(1) };
  }
  return { locale: defaultLocale, pageSlug: parts };
}

export function docsSlugForPage(locale: Locale, pageSlug: string[]): string[] {
  if (locale === defaultLocale) {
    return pageSlug;
  }
  return pageSlug.length === 0 ? [locale] : [locale, ...pageSlug];
}

export function localeDocsPrefix(locale: Locale, basePath: string): string {
  const docs = basePath === '' ? '/docs' : `${basePath}/docs`;
  if (locale === defaultLocale) {
    return docs;
  }
  return `${docs}/${locale}`;
}

export { nonDefaultLocales };
