'use client';

import { calloutLabel } from '@/components/callout-labels';
import { parseDocsSlug } from '@/lib/locale-routing';
import { usePathname } from 'next/navigation';
import type { ReactNode } from 'react';

export type CalloutType =
  | 'note'
  | 'warning'
  | 'cost'
  | 'security'
  | 'local-first'
  | 'experimental';

const TYPE_STYLES: Record<
  CalloutType,
  { border: string; bg: string; text: string }
> = {
  note: {
    border: 'border-blue-500/60',
    bg: 'bg-blue-500/10',
    text: 'text-blue-100',
  },
  warning: {
    border: 'border-amber-500/60',
    bg: 'bg-amber-500/10',
    text: 'text-amber-100',
  },
  cost: {
    border: 'border-emerald-500/60',
    bg: 'bg-emerald-500/10',
    text: 'text-emerald-100',
  },
  security: {
    border: 'border-red-500/60',
    bg: 'bg-red-500/10',
    text: 'text-red-100',
  },
  'local-first': {
    border: 'border-violet-500/60',
    bg: 'bg-violet-500/10',
    text: 'text-violet-100',
  },
  experimental: {
    border: 'border-fd-muted-foreground/40',
    bg: 'bg-fd-muted',
    text: 'text-fd-foreground',
  },
};

type CalloutProps = {
  type?: CalloutType;
  title?: string;
  children: ReactNode;
};

function useDocsLocale() {
  const pathname = usePathname();
  const docsIndex = pathname.indexOf('/docs');
  const afterDocs =
    docsIndex >= 0 ? pathname.slice(docsIndex + '/docs'.length) : '';
  const slugParts = afterDocs.replace(/^\/+/, '').split('/').filter(Boolean);
  return parseDocsSlug(slugParts).locale;
}

export function Callout({ type = 'note', title, children }: CalloutProps) {
  const locale = useDocsLocale();
  const styles = TYPE_STYLES[type];
  const heading = title ?? calloutLabel(locale, type);

  return (
    <aside
      className={`my-4 rounded-lg border border-l-4 ${styles.border} ${styles.bg} p-4`}
      data-callout-type={type}
    >
      <p className={`mb-2 text-sm font-semibold ${styles.text}`}>{heading}</p>
      <div className="text-fd-muted-foreground [&>p]:my-2 [&>ul]:my-2">
        {children}
      </div>
    </aside>
  );
}
