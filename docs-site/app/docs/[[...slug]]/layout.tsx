import { DocsHeaderActions } from '@/components/docs-header-actions';
import { source } from '@/lib/source';
import { baseOptions } from '@/lib/layout.shared';
import { parseDocsSlug } from '@/lib/locale-routing';
import { DocsLayout } from 'fumadocs-ui/layouts/docs';
import type { ReactNode } from 'react';

export default async function DocsSlugLayout({
  children,
  params,
}: {
  children: ReactNode;
  params: Promise<{ slug?: string[] }>;
}) {
  const { slug } = await params;
  const { locale } = parseDocsSlug(slug);

  const options = baseOptions(locale);

  return (
    <DocsLayout
      tree={source.getPageTree(locale)}
      {...options}
      nav={{
        ...options.nav,
        enabled: true,
        children: <DocsHeaderActions />,
      }}
    >
      {children}
    </DocsLayout>
  );
}
