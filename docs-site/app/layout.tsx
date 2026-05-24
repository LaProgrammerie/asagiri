import './globals.css';
import { LocaleProvider } from '@/components/locale-provider';
import { Inter } from 'next/font/google';
import type { Metadata } from 'next';
import type { ReactNode } from 'react';

const inter = Inter({ subsets: ['latin'] });

export const metadata: Metadata = {
  title: {
    default: 'Asagiri',
    template: '%s | Asagiri',
  },
  description: 'Documentation for Asagiri.',
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en" className="dark" suppressHydrationWarning>
      <body className={`${inter.className} flex min-h-screen flex-col`}>
        <LocaleProvider>{children}</LocaleProvider>
      </body>
    </html>
  );
}
