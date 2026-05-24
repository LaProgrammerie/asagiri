import { defineI18n } from 'fumadocs-core/i18n';

export const i18n = defineI18n({
  defaultLanguage: 'en',
  languages: ['en', 'fr', 'de', 'es'] as const,
  parser: 'dir',
  fallbackLanguage: null,
});

export type Locale = (typeof i18n.languages)[number];

export const locales = i18n.languages;

export const defaultLocale: Locale = i18n.defaultLanguage;

export function isLocale(value: string): value is Locale {
  return (locales as readonly string[]).includes(value);
}
