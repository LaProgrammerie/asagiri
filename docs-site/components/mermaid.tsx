'use client';

import { useEffect, useId, useRef, useState } from 'react';

type MermaidProps = {
  chart: string;
};

export function Mermaid({ chart }: MermaidProps) {
  const id = useId().replace(/:/g, '');
  const containerRef = useRef<HTMLDivElement>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    void (async () => {
      try {
        const mod = await import('mermaid');
        const mermaid = mod.default;
        mermaid.initialize({
          startOnLoad: false,
          theme: 'dark',
          securityLevel: 'strict',
        });
        if (!containerRef.current || cancelled) {
          return;
        }
        const { svg } = await mermaid.render(`mermaid-${id}`, chart);
        if (cancelled || !containerRef.current) {
          return;
        }
        containerRef.current.innerHTML = svg;
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to render diagram');
        }
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [chart, id]);

  if (error !== null) {
    return (
      <pre className="overflow-x-auto rounded-lg border border-red-500/40 bg-red-500/10 p-4 text-sm text-red-200">
        {error}
      </pre>
    );
  }

  return (
    <div
      ref={containerRef}
      className="my-6 overflow-x-auto rounded-lg border border-fd-border bg-fd-card p-4 [&_svg]:mx-auto"
      data-mermaid-id={id}
    />
  );
}
