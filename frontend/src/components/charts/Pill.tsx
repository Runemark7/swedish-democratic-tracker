import type { CSSProperties, ReactNode } from 'react';

type PillTone = 'neutral' | 'pass' | 'fail' | 'pending';

interface PillProps {
  tone?: PillTone;
  children: ReactNode;
}

const toneStyles: Record<PillTone, CSSProperties> = {
  neutral: {
    background: 'var(--color-track)',
    color: 'var(--color-fg)',
  },
  pass: {
    background: 'color-mix(in oklch, var(--color-up) 16%, transparent)',
    color: 'var(--color-up)',
  },
  fail: {
    background: 'color-mix(in oklch, var(--color-down) 16%, transparent)',
    color: 'var(--color-down)',
  },
  pending: {
    background: 'color-mix(in oklch, var(--color-accent) 16%, transparent)',
    color: 'var(--color-accent)',
  },
};

const baseStyle: CSSProperties = {
  display: 'inline-block',
  padding: '2px 8px',
  borderRadius: '3px',
  fontSize: '11px',
  fontWeight: 500,
  letterSpacing: '0.2px',
};

export function Pill({ tone = 'neutral', children }: PillProps) {
  return (
    <span style={{ ...baseStyle, ...toneStyles[tone] }}>
      {children}
    </span>
  );
}
