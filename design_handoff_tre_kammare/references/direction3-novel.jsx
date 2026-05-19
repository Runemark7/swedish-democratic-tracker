// Direction 3: "Tre Kammare" — the chosen direction.
// Dark-first spatial metaphor: Riksdag → Region → Kommun as nested/stacked chambers,
// each with its own live pulse. Refined: clearer hero, KPI target bars, consistent geometry.

const d3Theme = {
  '--sdt-bg': '#f5f3ee',
  '--sdt-surface': '#ffffff',
  '--sdt-fg': '#0e1620',
  '--sdt-fg-muted': '#5a6470',
  '--sdt-border': 'rgba(14,22,32,0.14)',
  '--sdt-track': 'rgba(14,22,32,0.06)',
  '--sdt-accent': '#0b3d7a',
  '--sdt-accent-2': '#c8a13b',
  '--sdt-pulse': '#d0533f',
  '--sdt-up': '#2d7a4a',
  '--sdt-down': '#b8391c',
  '--sdt-plane': 'rgba(14,22,32,0.04)',
};

const d3Dark = {
  '--sdt-bg': '#0a0e14',
  '--sdt-surface': '#10151d',
  '--sdt-fg': '#eef1f5',
  '--sdt-fg-muted': '#7c8896',
  '--sdt-border': 'rgba(238,241,245,0.14)',
  '--sdt-track': 'rgba(238,241,245,0.08)',
  '--sdt-accent': '#6fb0f2',
  '--sdt-accent-2': '#e0bc5e',
  '--sdt-pulse': '#e87560',
  '--sdt-up': '#5fbb7f',
  '--sdt-down': '#e07a62',
  '--sdt-plane': 'rgba(238,241,245,0.04)',
};

const d3Display = `"GT Sectra", "Tiempos Headline", Georgia, serif`;
const d3Sans = `"Söhne", "Inter", system-ui, sans-serif`;
const d3Mono = `"Berkeley Mono", "JetBrains Mono", ui-monospace, monospace`;

function D3Frame({ children, theme = 'dark', width = 1280, height = 880 }) {
  const vars = theme === 'dark' ? d3Dark : d3Theme;
  return (
    <div style={{
      width, height, overflow: 'hidden', position: 'relative',
      background: 'var(--sdt-bg)', color: 'var(--sdt-fg)',
      fontFamily: d3Sans, fontSize: 13, lineHeight: 1.5,
      ...vars,
    }}>
      <style>{`@keyframes d3pulse { 0%,100% { opacity: 1; } 50% { opacity: 0.35; } }`}</style>
      {children}
    </div>
  );
}

function D3Nav({ active = 'Start' }) {
  const items = [
    { k: 'Start', l: '◆' },
    { k: 'Riksdag', l: 'I' },
    { k: 'Region', l: 'II' },
    { k: 'Kommun', l: 'III' },
    { k: 'Sök', l: '⌕' },
  ];
  return (
    <div style={{
      position: 'absolute', top: 0, left: 0, right: 0, zIndex: 10,
      display: 'flex', justifyContent: 'space-between', alignItems: 'center',
      padding: '20px 32px',
    }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
        <svg width="22" height="22" viewBox="0 0 22 22">
          <rect x="2" y="14" width="18" height="4" fill="var(--sdt-accent)" opacity="0.45" />
          <rect x="2" y="9" width="18" height="4" fill="var(--sdt-accent)" opacity="0.7" />
          <rect x="2" y="4" width="18" height="4" fill="var(--sdt-accent)" opacity="1" />
        </svg>
        <div style={{ fontFamily: d3Display, fontSize: 18, fontWeight: 500, letterSpacing: -0.3 }}>
          Tre <i style={{ color: 'var(--sdt-accent-2)' }}>Kammare</i>
        </div>
      </div>
      <div style={{ display: 'flex', gap: 6, fontFamily: d3Mono, fontSize: 11 }}>
        {items.map(it => (
          <div key={it.k} style={{
            padding: '6px 14px', borderRadius: 999,
            border: '1px solid var(--sdt-border)',
            background: it.k === active ? 'var(--sdt-fg)' : 'transparent',
            color: it.k === active ? 'var(--sdt-bg)' : 'var(--sdt-fg-muted)',
            letterSpacing: 1,
          }}>
            <span style={{ marginRight: 6 }}>{it.l}</span>{it.k.toUpperCase()}
          </div>
        ))}
      </div>
      <div style={{ fontFamily: d3Mono, fontSize: 10, color: 'var(--sdt-fg-muted)', letterSpacing: 1.5 }}>
        <span style={{ display: 'inline-block', width: 6, height: 6, borderRadius: '50%', background: 'var(--sdt-pulse)', marginRight: 6, animation: 'd3pulse 2s infinite' }} />
        LIVE · 23.04.2026
      </div>
    </div>
  );
}

// ── Concentric-chambers hero ── Riksdag contains Region contains Kommun.
// Cleaner than 3D planes; reads as diagram, not illustration.
function D3ChambersHero() {
  const d = window.SDT_DATA;
  const rings = [
    { tag: 'I',   label: 'Riksdagen',  sub: '349 mandat · nationellt',         inset: 0,   pulses: 4,  activeToday: 3, party: d.riksdag.ruling.parties },
    { tag: 'II',  label: 'Regionen',   sub: '21 regioner · vård & trafik',     inset: 52,  pulses: 3,  activeToday: 1, party: d.region.ruling.parties },
    { tag: 'III', label: 'Kommunen',   sub: '290 kommuner · skola, omsorg',    inset: 104, pulses: 5,  activeToday: 2, party: d.kommun.ruling.parties },
  ];
  return (
    <div style={{ position: 'relative', height: 320, margin: '0 40px' }}>
      {rings.map((r, i) => (
        <div key={r.tag} style={{
          position: 'absolute',
          top: r.inset, left: r.inset + 80, right: r.inset + 80, bottom: r.inset,
          border: '1px solid var(--sdt-border)',
          borderRadius: 3,
          background: i === 2 ? 'var(--sdt-plane)' : 'transparent',
          boxShadow: i === 0 ? '0 30px 60px -20px rgba(0,0,0,0.4)' : 'none',
        }}>
          {/* Label tab top-left */}
          <div style={{
            position: 'absolute', top: -1, left: 14,
            transform: 'translateY(-50%)',
            background: 'var(--sdt-bg)', padding: '0 12px',
            display: 'flex', alignItems: 'baseline', gap: 10,
          }}>
            <span style={{ fontFamily: d3Display, fontStyle: 'italic', fontSize: 14, color: 'var(--sdt-accent-2)' }}>{r.tag}</span>
            <span style={{ fontFamily: d3Display, fontSize: 18, fontWeight: 500, letterSpacing: -0.3 }}>{r.label}</span>
            <span style={{ fontFamily: d3Mono, fontSize: 10, color: 'var(--sdt-fg-muted)', letterSpacing: 1 }}>{r.sub.toUpperCase()}</span>
          </div>
          {/* Pulse dots, bottom edge */}
          {Array.from({ length: r.pulses }).map((_, j) => (
            <div key={j} style={{
              position: 'absolute', bottom: -4,
              left: `${10 + (j / r.pulses) * 78}%`,
              width: 9, height: 9, borderRadius: '50%',
              background: j < r.activeToday ? 'var(--sdt-pulse)' : 'var(--sdt-accent)',
              border: '2px solid var(--sdt-bg)',
              animation: j < r.activeToday ? 'd3pulse 2.4s infinite' : 'none',
            }} />
          ))}
          {/* Ruling-party strip, top-right */}
          <div style={{
            position: 'absolute', top: 14, right: 14,
            display: 'flex', gap: 3,
          }}>
            {r.party.map(p => (
              <div key={p.short} style={{ width: 18, height: 4, background: p.color, borderRadius: 1 }} title={p.short} />
            ))}
          </div>
        </div>
      ))}
      {/* Center legend */}
      <div style={{
        position: 'absolute', top: '50%', left: '50%', transform: 'translate(-50%, -50%)',
        textAlign: 'center', pointerEvents: 'none',
      }}>
        <div style={{ fontFamily: d3Mono, fontSize: 10, letterSpacing: 2, color: 'var(--sdt-fg-muted)', marginBottom: 6 }}>
          12 PÅGÅENDE · 6 BESLUT IDAG
        </div>
        <div style={{ fontFamily: d3Mono, fontSize: 10, color: 'var(--sdt-fg-muted)', display: 'flex', gap: 14, justifyContent: 'center' }}>
          <span><span style={{ display: 'inline-block', width: 7, height: 7, borderRadius: '50%', background: 'var(--sdt-accent)', verticalAlign: 'middle', marginRight: 5 }} />PÅGÅR</span>
          <span><span style={{ display: 'inline-block', width: 7, height: 7, borderRadius: '50%', background: 'var(--sdt-pulse)', verticalAlign: 'middle', marginRight: 5 }} />BESLUT IDAG</span>
        </div>
      </div>
    </div>
  );
}

// ── Homepage ──
function D3Home() {
  const d = window.SDT_DATA;
  return (
    <>
      <D3Nav active="Start" />

      <div style={{ padding: '80px 40px 24px' }}>
        <div style={{ fontFamily: d3Mono, fontSize: 10, letterSpacing: 2.5, color: 'var(--sdt-fg-muted)', marginBottom: 16 }}>
          VEM BESTÄMMER · VAR · JUST NU
        </div>
        <div style={{ fontFamily: d3Display, fontSize: 60, fontWeight: 400, letterSpacing: -1.8, lineHeight: 0.98, maxWidth: 820 }}>
          Tre <i style={{ color: 'var(--sdt-accent-2)' }}>kammare</i>, ett samhälle.
        </div>
      </div>

      <D3ChambersHero />

      {/* Three mini status blocks */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 1, background: 'var(--sdt-border)', margin: '24px 32px 0', borderTop: '1px solid var(--sdt-border)', borderBottom: '1px solid var(--sdt-border)' }}>
        {[
          { tag: 'I', key: 'riksdag', title: 'Riksdagen', data: d.riksdag, todayCount: 3 },
          { tag: 'II', key: 'region', title: 'Region Stockholm', data: d.region, todayCount: 1 },
          { tag: 'III', key: 'kommun', title: 'Malmö kommun', data: d.kommun, todayCount: 2 },
        ].map((lvl, i) => {
          const ruling = [...lvl.data.ruling.parties, ...(lvl.data.ruling.support || [])];
          const rulingSeats = ruling.reduce((s, p) => s + p.seats, 0);
          const totalSeats = rulingSeats + lvl.data.ruling.opposition.reduce((s, p) => s + p.seats, 0);
          return (
            <div key={lvl.key} style={{ background: 'var(--sdt-bg)', padding: '22px 22px 18px' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline', marginBottom: 12 }}>
                <div style={{ display: 'flex', alignItems: 'baseline', gap: 10 }}>
                  <span style={{ fontFamily: d3Display, fontSize: 20, color: 'var(--sdt-accent-2)', fontStyle: 'italic' }}>{lvl.tag}</span>
                  <span style={{ fontFamily: d3Display, fontSize: 22, fontWeight: 500, letterSpacing: -0.3 }}>{lvl.title}</span>
                </div>
                <div style={{ fontFamily: d3Mono, fontSize: 10, color: 'var(--sdt-pulse)', letterSpacing: 1 }}>
                  ● {lvl.todayCount} IDAG
                </div>
              </div>

              <div style={{ marginBottom: 10 }}>
                <SDTStackBar segments={[
                  ...lvl.data.ruling.parties.map(p => ({ ...p, value: p.seats })),
                  ...(lvl.data.ruling.support || []).map(p => ({ ...p, value: p.seats })),
                  ...lvl.data.ruling.opposition.map(p => ({ ...p, value: p.seats })),
                ]} height={8} rounded={false} />
              </div>

              <div style={{ fontFamily: d3Mono, fontSize: 10, color: 'var(--sdt-fg-muted)', letterSpacing: 1, marginBottom: 4 }}>
                STYRE
              </div>
              <div style={{ fontSize: 13, marginBottom: 6 }}>{lvl.data.ruling.type}</div>
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: '4px 6px', marginBottom: 14 }}>
                {ruling.map(p => (
                  <span key={p.short} style={{
                    fontFamily: d3Mono, fontSize: 10, padding: '2px 7px',
                    background: 'var(--sdt-track)', borderRadius: 2,
                    display: 'inline-flex', alignItems: 'center', gap: 5,
                  }}>
                    <span style={{ width: 7, height: 7, background: p.color, borderRadius: '50%' }} />
                    {p.short} {p.seats}
                  </span>
                ))}
              </div>
              <div style={{ fontFamily: d3Mono, fontSize: 10, color: 'var(--sdt-fg-muted)', letterSpacing: 0.8, marginBottom: 12 }}>
                MAJORITET {rulingSeats}/{totalSeats}
              </div>

              {lvl.data.liveVotes.slice(0, 2).map((v, j) => (
                <div key={j} style={{
                  padding: '10px 0', borderTop: '1px solid var(--sdt-border)',
                  display: 'flex', justifyContent: 'space-between', gap: 12, alignItems: 'baseline',
                }}>
                  <div>
                    <div style={{ fontSize: 12.5 }}>{v.title}</div>
                    <div style={{ fontFamily: d3Mono, fontSize: 10, color: 'var(--sdt-fg-muted)', letterSpacing: 1, marginTop: 2 }}>{v.time.toUpperCase()}</div>
                  </div>
                  <SDTPill tone={v.status === 'Bifall' ? 'pass' : v.status === 'Återremiss' ? 'pending' : 'fail'}>{v.status}</SDTPill>
                </div>
              ))}

              <div style={{ marginTop: 14, fontFamily: d3Mono, fontSize: 10, letterSpacing: 1.5, color: 'var(--sdt-accent)' }}>
                GÅ TILL KAMMARE →
              </div>
            </div>
          );
        })}
      </div>

      {/* Bottom: 24h pulse strip */}
      <div style={{ padding: '22px 32px 28px' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline', marginBottom: 12 }}>
          <div style={{ fontFamily: d3Mono, fontSize: 10, letterSpacing: 2, color: 'var(--sdt-fg-muted)' }}>
            DYGNETS PULS — 9 BESLUT I 3 KAMMARE
          </div>
          <div style={{ fontFamily: d3Mono, fontSize: 10, color: 'var(--sdt-fg-muted)' }}>
            <span style={{ display: 'inline-block', width: 8, height: 8, background: 'var(--sdt-pulse)', marginRight: 6 }} />I
            <span style={{ display: 'inline-block', width: 8, height: 8, background: 'var(--sdt-accent)', margin: '0 6px 0 14px' }} />II
            <span style={{ display: 'inline-block', width: 8, height: 8, background: 'var(--sdt-accent-2)', margin: '0 6px 0 14px' }} />III
          </div>
        </div>
        <div style={{ display: 'flex', gap: 2 }}>
          {Array.from({ length: 48 }).map((_, i) => {
            const hits = [5, 12, 19, 26, 31, 38, 42];
            const v = hits.includes(i);
            const level = v ? (i === 12 || i === 31 ? 0 : i === 19 || i === 38 ? 1 : 2) : 0;
            return (
              <div key={i} style={{
                flex: 1, height: 28,
                background: v
                  ? (level === 0 ? 'var(--sdt-pulse)' : level === 1 ? 'var(--sdt-accent)' : 'var(--sdt-accent-2)')
                  : 'var(--sdt-track)',
                borderRadius: 1,
              }} />
            );
          })}
        </div>
        <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: 6, fontFamily: d3Mono, fontSize: 10, color: 'var(--sdt-fg-muted)' }}>
          <span>00:00</span><span>06:00</span><span>12:00</span><span>18:00</span><span>NU</span>
        </div>
      </div>
    </>
  );
}

// ── Level page — shared layout ──
function D3LevelPage({ tag, title, subtitle, data, showKpis, switcherLabel }) {
  const allParties = [
    ...data.ruling.parties,
    ...(data.ruling.support || []),
    ...data.ruling.opposition,
  ];
  const ruling = [...data.ruling.parties, ...(data.ruling.support || [])];
  const rulingSeats = ruling.reduce((s, p) => s + p.seats, 0);
  const totalSeats = rulingSeats + data.ruling.opposition.reduce((s, p) => s + p.seats, 0);

  const budgetColors = ['#0b3d7a', '#2d6fa8', '#5a9fd0', '#8bc0e0', '#c8a13b', '#d0533f', '#7a8390', '#b3bcc5'];
  const budgetWithColor = data.budget.areas.map((a, i) => ({ ...a, color: budgetColors[i % budgetColors.length] }));

  return (
    <>
      <D3Nav active={tag === 'I' ? 'Riksdag' : tag === 'II' ? 'Region' : 'Kommun'} />

      <div style={{ padding: '70px 32px 0' }}>
        <div style={{ display: 'flex', alignItems: 'baseline', gap: 20, marginBottom: 10 }}>
          <div style={{ fontFamily: d3Display, fontSize: 88, fontWeight: 400, color: 'var(--sdt-accent-2)', fontStyle: 'italic', letterSpacing: -3, lineHeight: 0.85 }}>{tag}</div>
          <div>
            <div style={{ fontFamily: d3Mono, fontSize: 10, letterSpacing: 2, color: 'var(--sdt-fg-muted)', marginBottom: 6 }}>KAMMARE {tag === 'I' ? 'ETT' : tag === 'II' ? 'TVÅ' : 'TRE'}</div>
            <div style={{ fontFamily: d3Display, fontSize: 52, fontWeight: 400, letterSpacing: -1.5, lineHeight: 0.95 }}>{title}</div>
            <div style={{ fontSize: 13, color: 'var(--sdt-fg-muted)', marginTop: 10, maxWidth: 600 }}>
              {subtitle}
              {switcherLabel && <span style={{ marginLeft: 12, fontFamily: d3Mono, fontSize: 11, color: 'var(--sdt-accent)', letterSpacing: 1 }}>↓ BYT {switcherLabel}</span>}
            </div>
          </div>
        </div>
      </div>

      {/* KPIs with target bars (Region/Kommun) */}
      {showKpis && data.kpis && (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', margin: '22px 32px 0', gap: 1, background: 'var(--sdt-border)', border: '1px solid var(--sdt-border)' }}>
          {data.kpis.map((k, i) => (
            <div key={i} style={{ background: 'var(--sdt-bg)', padding: '16px 20px' }}>
              <div style={{ fontFamily: d3Mono, fontSize: 10, letterSpacing: 1.5, color: 'var(--sdt-fg-muted)', marginBottom: 6 }}>{k.label.toUpperCase()}</div>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
                <div style={{ fontFamily: d3Display, fontSize: 30, fontWeight: 500, letterSpacing: -0.8, fontVariantNumeric: 'tabular-nums' }}>{k.value}</div>
                <SDTTrend trend={k.trend} delta={k.delta} />
              </div>
              {k.target != null && (
                <div style={{ marginTop: 10 }}>
                  <SDTTargetBar value={k.raw} target={k.target} worseHigher={k.worseHigher} unit={k.unit} />
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {/* Main grid: chamber composition + budget */}
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 1, background: 'var(--sdt-border)', margin: '22px 32px 0', border: '1px solid var(--sdt-border)' }}>
        <div style={{ background: 'var(--sdt-bg)', padding: '20px 22px' }}>
          <div style={{ fontFamily: d3Mono, fontSize: 10, letterSpacing: 2, color: 'var(--sdt-fg-muted)', marginBottom: 14 }}>MANDAT · KAMMARENS SAMMANSÄTTNING</div>
          <SDTHemicycle groups={[
            ...data.ruling.opposition.map(p => ({ color: p.color, count: p.seats })),
            ...(data.ruling.support || []).map(p => ({ color: p.color, count: p.seats })),
            ...data.ruling.parties.map(p => ({ color: p.color, count: p.seats })),
          ]} width={440} height={150} />
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px 14px', marginTop: 14 }}>
            {allParties.map(p => (
              <div key={p.short} style={{ display: 'flex', alignItems: 'center', gap: 5, fontSize: 11 }}>
                <span style={{ width: 9, height: 9, background: p.color, borderRadius: '50%' }} />
                <span style={{ fontFamily: d3Mono, fontWeight: 600 }}>{p.short}</span>
                <span style={{ color: 'var(--sdt-fg-muted)', fontFamily: d3Mono }}>{p.seats}</span>
              </div>
            ))}
          </div>
          <div style={{ marginTop: 14, paddingTop: 12, borderTop: '1px solid var(--sdt-border)', display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
            <div style={{ fontFamily: d3Display, fontSize: 18, fontStyle: 'italic' }}>{data.ruling.type}</div>
            <div style={{ fontFamily: d3Mono, fontSize: 10, color: 'var(--sdt-fg-muted)', letterSpacing: 0.8 }}>
              MAJORITET {rulingSeats}/{totalSeats}
            </div>
          </div>
        </div>

        <div style={{ background: 'var(--sdt-bg)', padding: '20px 22px' }}>
          <div style={{ fontFamily: d3Mono, fontSize: 10, letterSpacing: 2, color: 'var(--sdt-fg-muted)', marginBottom: 14 }}>BUDGET {data.budget.year} · {data.budget.total}</div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 18 }}>
            <SDTDonut segments={budgetWithColor} size={150} thickness={18} label={data.budget.total.split(' ')[0]} sublabel={data.budget.total.split(' ')[1] || ''} />
            <div style={{ flex: 1 }}>
              <SDTHBars items={budgetWithColor.slice(0, 5)} unit=" mdkr" />
            </div>
          </div>
        </div>
      </div>

      {/* Bottom: pulse (live votes) + agenda */}
      <div style={{ display: 'grid', gridTemplateColumns: '1.3fr 1fr', gap: 1, background: 'var(--sdt-border)', margin: '1px 32px 28px 32px', border: '1px solid var(--sdt-border)', borderTop: 'none' }}>
        <div style={{ background: 'var(--sdt-bg)', padding: '20px 22px' }}>
          <div style={{ fontFamily: d3Mono, fontSize: 10, letterSpacing: 2, color: 'var(--sdt-fg-muted)', marginBottom: 12 }}>
            <span style={{ display: 'inline-block', width: 6, height: 6, borderRadius: '50%', background: 'var(--sdt-pulse)', marginRight: 8, animation: 'd3pulse 2s infinite' }} />
            PULS · AKTUELLA BESLUT
          </div>
          {data.liveVotes.map((v, i) => (
            <div key={i} style={{
              display: 'grid', gridTemplateColumns: '18px 70px 1fr auto 90px',
              gap: 12, padding: '11px 0', alignItems: 'baseline',
              borderTop: i === 0 ? 'none' : '1px solid var(--sdt-border)',
            }}>
              <span style={{ width: 7, height: 7, borderRadius: '50%', background: i < 2 ? 'var(--sdt-pulse)' : 'var(--sdt-accent)', display: 'inline-block' }} />
              <span style={{ fontFamily: d3Mono, fontSize: 10, color: 'var(--sdt-fg-muted)', letterSpacing: 0.8 }}>{v.time.toUpperCase()}</span>
              <div>
                <div style={{ fontSize: 13.5 }}>{v.title}</div>
                <div style={{ fontFamily: d3Mono, fontSize: 10, color: 'var(--sdt-fg-muted)', letterSpacing: 1, marginTop: 2 }}>{v.tag.toUpperCase()}</div>
              </div>
              <SDTPill tone={v.status === 'Bifall' ? 'pass' : v.status === 'Återremiss' ? 'pending' : 'fail'}>{v.status}</SDTPill>
              <span style={{ fontFamily: d3Mono, fontSize: 11, textAlign: 'right', color: 'var(--sdt-fg-muted)' }}>{v.margin || '—'}</span>
            </div>
          ))}
        </div>
        <div style={{ background: 'var(--sdt-bg)', padding: '20px 22px' }}>
          <div style={{ fontFamily: d3Mono, fontSize: 10, letterSpacing: 2, color: 'var(--sdt-fg-muted)', marginBottom: 14 }}>AGENDA · STYRETS PRIORITERINGAR</div>
          {data.agenda.map((a, i) => (
            <div key={i} style={{ padding: '10px 0', borderTop: i === 0 ? 'none' : '1px solid var(--sdt-border)', display: 'flex', gap: 14, alignItems: 'baseline' }}>
              <span style={{ fontFamily: d3Display, fontSize: 22, fontStyle: 'italic', color: 'var(--sdt-accent-2)', minWidth: 24, lineHeight: 1 }}>{i+1}</span>
              <span style={{ fontSize: 13.5, lineHeight: 1.4 }}>{a}</span>
            </div>
          ))}
        </div>
      </div>
    </>
  );
}

function D3Riksdag() {
  const d = window.SDT_DATA.riksdag;
  return <D3LevelPage tag="I" title="Riksdagen" subtitle={d.subtitle + ' · Mandatperiod 2022–2026'} data={d} />;
}
function D3Region() {
  const d = window.SDT_DATA.region;
  return <D3LevelPage tag="II" title="Region Stockholm" subtitle={d.subtitle + ' · ' + d.population + ' invånare'} data={d} showKpis switcherLabel="REGION" />;
}
function D3Kommun() {
  const d = window.SDT_DATA.kommun;
  return <D3LevelPage tag="III" title="Malmö kommun" subtitle={d.subtitle + ' · ' + d.population + ' invånare'} data={d} showKpis switcherLabel="KOMMUN" />;
}

// ── Search/explore page ── jump across all three levels by topic or party
function D3Search() {
  const topics = [
    { t: 'Skola & utbildning', n: 24, lvl: ['I', 'III'] },
    { t: 'Vård & omsorg', n: 31, lvl: ['I', 'II'] },
    { t: 'Kollektivtrafik', n: 12, lvl: ['II', 'III'] },
    { t: 'Bostad & planering', n: 18, lvl: ['I', 'III'] },
    { t: 'Skatt & ekonomi', n: 22, lvl: ['I'] },
    { t: 'Miljö & klimat', n: 15, lvl: ['I', 'II', 'III'] },
    { t: 'Arbetsmarknad', n: 9, lvl: ['I'] },
    { t: 'Trygghet & brott', n: 14, lvl: ['I', 'III'] },
  ];
  const parties = ['S', 'M', 'SD', 'V', 'C', 'KD', 'L', 'MP'];
  return (
    <>
      <D3Nav active="Sök" />
      <div style={{ padding: '70px 32px 0' }}>
        <div style={{ fontFamily: d3Mono, fontSize: 10, letterSpacing: 2, color: 'var(--sdt-fg-muted)', marginBottom: 10 }}>SÖK · UTFORSKA</div>
        <div style={{ fontFamily: d3Display, fontSize: 48, fontWeight: 400, letterSpacing: -1.2, lineHeight: 0.98 }}>
          Vad <i style={{ color: 'var(--sdt-accent-2)' }}>händer</i> om…
        </div>

        <div style={{
          marginTop: 22, padding: '14px 18px',
          border: '1px solid var(--sdt-border)', borderRadius: 3,
          display: 'flex', alignItems: 'center', gap: 12,
          background: 'var(--sdt-surface)',
        }}>
          <span style={{ fontSize: 16, color: 'var(--sdt-fg-muted)' }}>⌕</span>
          <span style={{ fontFamily: d3Mono, fontSize: 13, color: 'var(--sdt-fg-muted)' }}>
            t.ex. <span style={{ color: 'var(--sdt-fg)' }}>förskola</span>, <span style={{ color: 'var(--sdt-fg)' }}>pensioner</span>, <span style={{ color: 'var(--sdt-fg)' }}>vårdcentral Norrtälje</span>…
          </span>
        </div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 1, background: 'var(--sdt-border)', margin: '22px 32px 0', border: '1px solid var(--sdt-border)' }}>
        <div style={{ background: 'var(--sdt-bg)', padding: '20px 22px' }}>
          <div style={{ fontFamily: d3Mono, fontSize: 10, letterSpacing: 2, color: 'var(--sdt-fg-muted)', marginBottom: 14 }}>ÄMNEN</div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px 16px' }}>
            {topics.map(t => (
              <div key={t.t} style={{
                padding: '10px 14px', border: '1px solid var(--sdt-border)', borderRadius: 2,
                background: 'var(--sdt-surface)',
                display: 'flex', justifyContent: 'space-between', alignItems: 'baseline', gap: 10,
              }}>
                <span style={{ fontSize: 13 }}>{t.t}</span>
                <div style={{ display: 'flex', alignItems: 'baseline', gap: 8 }}>
                  <div style={{ display: 'flex', gap: 2 }}>
                    {t.lvl.map(l => (
                      <span key={l} style={{ fontFamily: d3Display, fontStyle: 'italic', fontSize: 11, color: 'var(--sdt-accent-2)' }}>{l}</span>
                    ))}
                  </div>
                  <span style={{ fontFamily: d3Mono, fontSize: 11, color: 'var(--sdt-fg-muted)' }}>{t.n}</span>
                </div>
              </div>
            ))}
          </div>
        </div>
        <div style={{ background: 'var(--sdt-bg)', padding: '20px 22px' }}>
          <div style={{ fontFamily: d3Mono, fontSize: 10, letterSpacing: 2, color: 'var(--sdt-fg-muted)', marginBottom: 14 }}>PARTIER</div>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
            {parties.map(p => {
              const party = window.SDT_DATA.riksdag.ruling.parties.concat(window.SDT_DATA.riksdag.ruling.support, window.SDT_DATA.riksdag.ruling.opposition).find(x => x.short === p);
              return (
                <div key={p} style={{
                  padding: '8px 14px', border: '1px solid var(--sdt-border)', borderRadius: 2,
                  display: 'flex', alignItems: 'center', gap: 8, background: 'var(--sdt-surface)',
                }}>
                  <span style={{ width: 10, height: 10, background: party ? party.color : '#999', borderRadius: '50%' }} />
                  <span style={{ fontFamily: d3Mono, fontWeight: 600, fontSize: 13 }}>{p}</span>
                  <span style={{ fontFamily: d3Mono, fontSize: 10, color: 'var(--sdt-fg-muted)' }}>{party ? party.seats : '–'}</span>
                </div>
              );
            })}
          </div>
          <div style={{ fontFamily: d3Mono, fontSize: 10, letterSpacing: 2, color: 'var(--sdt-fg-muted)', marginTop: 22, marginBottom: 14 }}>SENASTE FRÅN DIN KOMMUN</div>
          {window.SDT_DATA.kommun.liveVotes.slice(0, 3).map((v, i) => (
            <div key={i} style={{ padding: '8px 0', borderTop: i === 0 ? 'none' : '1px solid var(--sdt-border)', display: 'flex', justifyContent: 'space-between', gap: 12, alignItems: 'baseline' }}>
              <span style={{ fontSize: 12.5 }}>{v.title}</span>
              <SDTPill tone={v.status === 'Bifall' ? 'pass' : v.status === 'Avslag' ? 'fail' : 'pending'}>{v.status}</SDTPill>
            </div>
          ))}
        </div>
      </div>
    </>
  );
}

Object.assign(window, { D3Frame, D3Home, D3Riksdag, D3Region, D3Kommun, D3Search });
