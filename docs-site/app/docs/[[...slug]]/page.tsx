import { getMDXComponents } from '@/components/mdx';
import { source } from '@/lib/source';
import type { DocPage } from '@/lib/doc-page';
import { docsSlugForPage, parseDocsSlug } from '@/lib/locale-routing';
import { i18n } from '@/lib/i18n';
import {
  DocsBody,
  DocsDescription,
  DocsPage,
  DocsTitle,
} from 'fumadocs-ui/page';
import { createRelativeLink } from 'fumadocs-ui/mdx';
import type { Metadata } from 'next';
import { notFound } from 'next/navigation';

type PageParams = { slug?: string[] };

export default async function Page(props: { params: Promise<PageParams> }) {
  const params = await props.params;
  const { locale, pageSlug } = parseDocsSlug(params.slug);
  const page = source.getPage(pageSlug, locale) as DocPage | undefined;
  if (!page) {
    notFound();
  }

  const MDX = page.data.body;

  return (
    <DocsPage toc={page.data.toc} full={page.data.full}>
      <DocsTitle>{page.data.title}</DocsTitle>
      <DocsDescription>{page.data.description}</DocsDescription>
      <DocsBody>
        <MDX
          components={getMDXComponents({
            a: createRelativeLink(source, page),
          })}
        />
      </DocsBody>
    </DocsPage>
  );
}

export function generateStaticParams() {
  return i18n.languages.flatMap((locale) =>
    source.getPages(locale).map((page) => ({
      slug: docsSlugForPage(locale, page.slugs),
    })),
  );
}

export async function generateMetadata(props: {
  params: Promise<PageParams>;
}): Promise<Metadata> {
  const params = await props.params;
  const { locale, pageSlug } = parseDocsSlug(params.slug);
  const page = source.getPage(pageSlug, locale);
  if (!page) {
    notFound();
  }

  return {
    title: page.data.title,
    description: page.data.description,
  };
}
