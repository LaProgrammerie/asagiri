import './globals.css';
import { RootProvider } from 'fumadocs-ui/provider/next';
import { Inter } from 'next/font/google';
import type { Metadata } from 'next';
import type { ReactNode } from 'react';

const inter = Inter({ subsets: ['latin'] });

export const metadata: Metadata = {
  title: {
    default: 'Hyper Fast Builder',
    template: '%s | Hyper Fast Builder',
  },
  description: 'Documentation for Hyper Fast Builder.',
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en" className="dark" suppressHydrationWarning>
      <body className={`${inter.className} flex min-h-screen flex-col`}>
        <RootProvider
          theme={{
            enabled: true,
            defaultTheme: 'dark',
          }}
          search={{
            enabled: false,
          }}
        >
          {children}
        </RootProvider>
      </body>
    </html>
  );
}
