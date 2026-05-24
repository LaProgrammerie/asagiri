'use client';

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
  { label: string; border: string; bg: string; text: string }
> = {
  note: {
    label: 'Note',
    border: 'border-blue-500/60',
    bg: 'bg-blue-500/10',
    text: 'text-blue-100',
  },
  warning: {
    label: 'Warning',
    border: 'border-amber-500/60',
    bg: 'bg-amber-500/10',
    text: 'text-amber-100',
  },
  cost: {
    label: 'Cost',
    border: 'border-emerald-500/60',
    bg: 'bg-emerald-500/10',
    text: 'text-emerald-100',
  },
  security: {
    label: 'Security',
    border: 'border-red-500/60',
    bg: 'bg-red-500/10',
    text: 'text-red-100',
  },
  'local-first': {
    label: 'Local-first',
    border: 'border-violet-500/60',
    bg: 'bg-violet-500/10',
    text: 'text-violet-100',
  },
  experimental: {
    label: 'Experimental',
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

export function Callout({ type = 'note', title, children }: CalloutProps) {
  const styles = TYPE_STYLES[type];
  const heading = title ?? styles.label;

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
