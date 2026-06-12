import { localeDocsPrefix, parseDocsSlug } from '@/lib/locale-routing';
import { isLocale, type Locale } from '@/lib/i18n';

export function docsBasePath(): string {
  return process.env.NEXT_PUBLIC_BASE_PATH ?? '';
}

/** Pathname from Next (without basePath). Returns docs home or same page in another locale. */
export function switchLocalePath(pathname: string, newLocale: string): string {
  if (!isLocale(newLocale)) {
    return localeDocsPrefix('en', docsBasePath()) + '/';
  }

  const locale = newLocale as Locale;
  const base = docsBasePath();
  const docsMarker = '/docs';
  const docsIndex = pathname.indexOf(docsMarker);

  if (docsIndex < 0) {
    return localeDocsPrefix(locale, base) + '/';
  }

  const afterDocs = pathname.slice(docsIndex + docsMarker.length);
  const slugParts = afterDocs.replace(/^\/+/, '').split('/').filter(Boolean);
  const { pageSlug } = parseDocsSlug(slugParts);
  const prefix = localeDocsPrefix(locale, base);

  if (pageSlug.length === 0) {
    return `${prefix}/`;
  }

  return `${prefix}/${pageSlug.join('/')}/`;
}
