import type { InferPageType } from 'fumadocs-core/source';
import type { MDXComponents } from 'mdx/types';
import type { ComponentType } from 'react';
import type { TableOfContents } from 'fumadocs-core/server';
import type { source } from '@/lib/source';

type BasePage = InferPageType<typeof source>;

export type DocPage = BasePage & {
  data: BasePage['data'] & {
    body: ComponentType<{ components?: MDXComponents }>;
    toc: TableOfContents;
    full?: boolean;
  };
};
