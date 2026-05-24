'use client';

import { i18nUI } from '@/lib/layout.shared';
import { parseDocsSlug } from '@/lib/locale-routing';
import { switchLocalePath } from '@/lib/switch-locale-path';
import { RootProvider } from 'fumadocs-ui/provider/next';
import { usePathname, useRouter } from 'next/navigation';
import type { ReactNode } from 'react';

type LocaleProviderProps = {
  children: ReactNode;
};

export function LocaleProvider({ children }: LocaleProviderProps) {
  const pathname = usePathname();
  const router = useRouter();
  const docsIndex = pathname.indexOf('/docs');
  const afterDocs =
    docsIndex >= 0 ? pathname.slice(docsIndex + '/docs'.length) : '';
  const slugParts = afterDocs.replace(/^\/+/, '').split('/').filter(Boolean);
  const { locale } = parseDocsSlug(slugParts);

  const provider = i18nUI.provider(locale);

  return (
    <RootProvider
      theme={{
        enabled: true,
        defaultTheme: 'dark',
      }}
      search={{
        enabled: false,
      }}
      i18n={{
        ...provider,
        onLocaleChange: (newLocale) => {
          router.push(switchLocalePath(pathname, newLocale));
        },
      }}
    >
      {children}
    </RootProvider>
  );
}
