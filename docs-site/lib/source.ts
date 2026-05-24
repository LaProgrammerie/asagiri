import { docs } from 'collections/server';
import { loader, type InferPageType } from 'fumadocs-core/source';

const basePath = process.env.GITHUB_PAGES === 'true' ? '/hyper-fast-builder' : '';
const docsBaseUrl = basePath === '' ? '/docs' : `${basePath}/docs`;

export const source = loader({
  baseUrl: docsBaseUrl,
  source: docs.toFumadocsSource(),
});

export type Page = InferPageType<typeof source>;
