'use client';

import { LanguageToggleFlags } from '@/components/language-toggle-flags';

/** Language switcher in navbar (theme toggle stays in sidebar footer only). */
export function DocsHeaderActions() {
  return <LanguageToggleFlags />;
}
