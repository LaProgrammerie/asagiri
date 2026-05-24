import Link from 'next/link';

const basePath = process.env.GITHUB_PAGES === 'true' ? '/hyper-fast-builder' : '';

const locales = [
  { code: 'en', label: 'English', href: `${basePath}/docs/` },
  { code: 'fr', label: 'Français', href: `${basePath}/docs/fr/` },
  { code: 'de', label: 'Deutsch', href: `${basePath}/docs/de/` },
  { code: 'es', label: 'Español', href: `${basePath}/docs/es/` },
] as const;

export default function HomePage() {
  return (
    <main className="mx-auto flex min-h-screen max-w-lg flex-col justify-center gap-8 px-6 py-16">
      <div>
        <h1 className="text-3xl font-semibold tracking-tight">AgentFlow</h1>
        <p className="mt-2 text-fd-muted-foreground">
          Deterministic orchestration for AI coding workflows.
        </p>
      </div>
      <section>
        <h2 className="mb-3 text-sm font-medium text-fd-muted-foreground">
          Documentation
        </h2>
        <ul className="grid gap-2">
          {locales.map(({ code, label, href }) => (
            <li key={code}>
              <Link
                href={href}
                className="block rounded-lg border border-fd-border px-4 py-3 transition-colors hover:bg-fd-accent"
              >
                <span className="font-medium">{label}</span>
                <span className="ml-2 text-sm text-fd-muted-foreground">
                  ({code})
                </span>
              </Link>
            </li>
          ))}
        </ul>
      </section>
    </main>
  );
}
