import { docs } from 'collections/server';
import { loader, type InferPageType } from 'fumadocs-core/source';
import { i18n } from '@/lib/i18n';

const basePath = process.env.GITHUB_PAGES === 'true' ? '/hyper-fast-builder' : '';
const docsBaseUrl = basePath === '' ? '/docs' : `${basePath}/docs`;

function docsUrl(slugs: string[], locale?: string): string {
  const lang = locale ?? i18n.defaultLanguage;
  const segments = [...slugs];
  if (lang !== i18n.defaultLanguage) {
    segments.unshift(lang);
  }
  const path = segments.length > 0 ? `/${segments.join('/')}` : '';
  return `${docsBaseUrl}${path}`;
}

export const source = loader({
  i18n,
  baseUrl: docsBaseUrl,
  url: docsUrl,
  source: docs.toFumadocsSource(),
});

export type Page = InferPageType<typeof source>;
