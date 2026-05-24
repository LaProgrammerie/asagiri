'use client';

import { LocaleFlag } from '@/components/locale-flags';
import { useI18n } from 'fumadocs-ui/contexts/i18n';
import { buttonVariants } from 'fumadocs-ui/components/ui/button';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from 'fumadocs-ui/components/ui/popover';
import { cn } from 'fumadocs-ui/utils/cn';

/** Language switcher with flag for current locale (navbar). */
export function LanguageToggleFlags() {
  const context = useI18n();
  if (!context.locales) {
    throw new Error('Missing `<I18nProvider />`');
  }

  return (
    <Popover>
      <PopoverTrigger
        aria-label={context.text.chooseLanguage}
        className={cn(
          buttonVariants({
            color: 'ghost',
            className: 'gap-2 px-2 py-1.5',
          }),
        )}
      >
        <LocaleFlag
          locale={context.locale ?? 'en'}
          className="h-4 w-6 rounded-sm object-cover"
        />
        <span className="hidden text-sm sm:inline">
          {context.locales.find((item) => item.locale === context.locale)?.name}
        </span>
      </PopoverTrigger>
      <PopoverContent className="flex flex-col overflow-hidden p-0">
        <p className="mb-1 p-2 text-xs font-medium text-fd-muted-foreground">
          {context.text.chooseLanguage}
        </p>
        {context.locales.map((item) => (
          <button
            key={item.locale}
            type="button"
            className={cn(
              'flex items-center gap-2 p-2 text-start text-sm',
              item.locale === context.locale
                ? 'bg-fd-primary/10 font-medium text-fd-primary'
                : 'hover:bg-fd-accent hover:text-fd-accent-foreground',
            )}
            onClick={() => context.onChange?.(item.locale)}
          >
            <LocaleFlag
              locale={item.locale}
              className="h-3.5 w-5 shrink-0 rounded-sm object-cover"
            />
            {item.name}
          </button>
        ))}
      </PopoverContent>
    </Popover>
  );
}
