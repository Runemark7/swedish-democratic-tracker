// Shared chart + visualization primitives for SDT
// All charts are inline SVG, theme-aware via CSS vars

// ─── Donut chart ───
function SDTDonut({ segments, size = 180, thickness = 28, label, sublabel, center = true }) {
  const radius = (size - thickness) / 2;
  const circ = 2 * Math.PI * radius;
  const total = segments.reduce((a, s) => a + s.value, 0);
  let offset = 0;
  return (
    <div style={{ position: 'relative', width: size, height: size }}>
      <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`} style={{ transform: 'rotate(-90deg)' }}>
        <circle cx={size/2} cy={size/2} r={radius} fill="none"
          stroke="var(--sdt-track, rgba(0,0,0,0.06))" strokeWidth={thickness} />
        {segments.map((s, i) => {
          const frac = s.value / total;
          const dash = frac * circ;
          const el = (
            <circle key={i} cx={size/2} cy={size/2} r={radius} fill="none"
              stroke={s.color} strokeWidth={thickness}
              strokeDasharray={`${dash} ${circ - dash}`}
              strokeDashoffset={-offset} />
          );
          offset += dash;
          return el;
        })}
      </svg>
      {center && (() => {
        // Size the label to fit within the donut hole (inner radius * 2, with some padding).
        const inner = (size - thickness * 2) - 18;
        const maxLabelChars = Math.max(4, String(label || '').length);
        // ~0.58em per char for numeric/mixed content; clamp so it never exceeds hole width.
        const byWidth = inner / (maxLabelChars * 0.58);
        const labelSize = Math.max(11, Math.min(size * 0.16, byWidth));
        const subSize = Math.max(9, Math.min(size * 0.075, labelSize * 0.5));
        return (
          <div style={{
            position: 'absolute', inset: 0, display: 'flex',
            flexDirection: 'column', alignItems: 'center', justifyContent: 'center',
            textAlign: 'center', pointerEvents: 'none', padding: thickness,
          }}>
            <div style={{
              fontSize: labelSize, fontWeight: 600, color: 'var(--sdt-fg)',
              lineHeight: 1, whiteSpace: 'nowrap', fontVariantNumeric: 'tabular-nums',
            }}>{label}</div>
            {sublabel && <div style={{
              fontSize: subSize, color: 'var(--sdt-fg-muted)',
              marginTop: 4, letterSpacing: 0.4, textTransform: 'uppercase',
              whiteSpace: 'nowrap',
            }}>{sublabel}</div>}
          </div>
        );
      })()}
    </div>
  );
}

// ─── Horizontal stacked bar (parliament composition) ───
function SDTStackBar({ segments, height = 14, showLabels = false, rounded = true }) {
  const total = segments.reduce((a, s) => a + s.value, 0);
  return (
    <div>
      <div style={{
        display: 'flex', width: '100%', height,
        borderRadius: rounded ? height/2 : 2, overflow: 'hidden',
        background: 'var(--sdt-track, rgba(0,0,0,0.06))',
      }}>
        {segments.map((s, i) => (
          <div key={i} style={{
            width: `${(s.value/total)*100}%`,
            background: s.color,
            transition: 'width 0.4s',
          }} title={`${s.name || s.short}: ${s.value}`} />
        ))}
      </div>
      {showLabels && (
        <div style={{ display: 'flex', marginTop: 8, fontSize: 11, color: 'var(--sdt-fg-muted)' }}>
          {segments.map((s, i) => (
            <div key={i} style={{ width: `${(s.value/total)*100}%`, paddingLeft: 2 }}>
              <span style={{ fontWeight: 600, color: s.color }}>{s.short}</span> {s.value}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

// ─── Simple horizontal bar list ───
function SDTHBars({ items, max, unit = "", valueColor = "var(--sdt-accent)", barBg = "var(--sdt-track)", height = 6, gap = 10 }) {
  const m = max || Math.max(...items.map(i => i.value));
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap }}>
      {items.map((it, i) => (
        <div key={i}>
          <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 12, marginBottom: 4, color: 'var(--sdt-fg)' }}>
            <span>{it.name}</span>
            <span style={{ color: 'var(--sdt-fg-muted)', fontVariantNumeric: 'tabular-nums' }}>
              {it.value}{unit}{it.pct != null ? ` · ${it.pct}%` : ''}
            </span>
          </div>
          <div style={{ width: '100%', height, background: barBg, borderRadius: height/2, overflow: 'hidden' }}>
            <div style={{
              width: `${(it.value/m)*100}%`, height: '100%',
              background: it.color || valueColor, transition: 'width 0.4s',
            }} />
          </div>
        </div>
      ))}
    </div>
  );
}

// ─── Sparkline (line) ───
function SDTSpark({ points, width = 120, height = 32, color = "var(--sdt-accent)", fill = true }) {
  const min = Math.min(...points);
  const max = Math.max(...points);
  const range = max - min || 1;
  const path = points.map((p, i) => {
    const x = (i/(points.length-1)) * width;
    const y = height - ((p - min)/range) * height;
    return `${i===0 ? 'M' : 'L'} ${x.toFixed(1)} ${y.toFixed(1)}`;
  }).join(' ');
  const area = fill ? `${path} L ${width} ${height} L 0 ${height} Z` : null;
  return (
    <svg width={width} height={height} viewBox={`0 0 ${width} ${height}`}>
      {area && <path d={area} fill={color} opacity={0.12} />}
      <path d={path} fill="none" stroke={color} strokeWidth={1.5} strokeLinejoin="round" strokeLinecap="round" />
    </svg>
  );
}

// ─── Parliament seating arc (hemicycle) ───
function SDTHemicycle({ groups, width = 320, height = 140 }) {
  // groups: [{color, count}] left-to-right (from opposition to government, order)
  const total = groups.reduce((a, g) => a + g.count, 0);
  const rows = 6;
  // estimate per-row seat count (inner rows have fewer)
  const rowCounts = Array.from({ length: rows }, (_, r) => {
    const f = (r+1)/rows;
    return Math.round((total/rows) * (0.7 + 0.6*f));
  });
  const sum = rowCounts.reduce((a,b)=>a+b,0);
  const scale = total/sum;
  const adjusted = rowCounts.map(c => Math.max(1, Math.round(c*scale)));
  // Flatten seats with color assignments
  const flat = [];
  groups.forEach(g => {
    for (let i = 0; i < g.count; i++) flat.push(g.color);
  });
  // arrange seats row by row, left-to-right
  const cx = width/2;
  const cy = height;
  const rMin = height * 0.35;
  const rMax = height * 0.95;
  const seats = [];
  let idx = 0;
  for (let r = 0; r < rows; r++) {
    const rowR = rMin + (rMax - rMin) * (r/(rows-1));
    const n = adjusted[r];
    for (let s = 0; s < n; s++) {
      const t = n === 1 ? 0.5 : s/(n-1);
      const angle = Math.PI * (1 - t); // pi to 0
      const x = cx + rowR * Math.cos(angle);
      const y = cy - rowR * Math.sin(angle);
      seats.push({ x, y, color: flat[idx] || '#ccc' });
      idx++;
    }
  }
  const dotR = Math.max(2, Math.min(4.5, width/90));
  return (
    <svg width={width} height={height} viewBox={`0 0 ${width} ${height}`}>
      {seats.map((s, i) => (
        <circle key={i} cx={s.x} cy={s.y} r={dotR} fill={s.color} />
      ))}
    </svg>
  );
}

// ─── Trend indicator ───
function SDTTrend({ trend, delta }) {
  const symbol = trend === 'up' ? '↑' : trend === 'down' ? '↓' : '→';
  const color = trend === 'up' ? 'var(--sdt-up)' : trend === 'down' ? 'var(--sdt-down)' : 'var(--sdt-fg-muted)';
  return (
    <span style={{ color, fontVariantNumeric: 'tabular-nums', fontSize: 12 }}>
      {symbol} {delta}
    </span>
  );
}

// ─── Status pill ───
function SDTPill({ children, tone = 'neutral' }) {
  const tones = {
    neutral: { bg: 'var(--sdt-track)', fg: 'var(--sdt-fg)' },
    pass:    { bg: 'color-mix(in oklch, var(--sdt-up) 16%, transparent)', fg: 'var(--sdt-up)' },
    fail:    { bg: 'color-mix(in oklch, var(--sdt-down) 16%, transparent)', fg: 'var(--sdt-down)' },
    pending: { bg: 'color-mix(in oklch, var(--sdt-accent) 16%, transparent)', fg: 'var(--sdt-accent)' },
  };
  const t = tones[tone] || tones.neutral;
  return (
    <span style={{
      display: 'inline-block', padding: '2px 8px', borderRadius: 3,
      fontSize: 11, fontWeight: 500, letterSpacing: 0.2,
      background: t.bg, color: t.fg,
    }}>{children}</span>
  );
}

// ─── Target bar — shows current value vs target on a scaled axis ───
// Use when a KPI has a clear goal. Renders: scale, current marker, target tick, range labels.
function SDTTargetBar({ value, target, min, max, worseHigher = false, unit = "", width = "100%", height = 30 }) {
  // Auto-compute a sensible axis if not given
  const lo = min != null ? min : Math.min(value, target) * 0.85;
  const hi = max != null ? max : Math.max(value, target) * 1.15;
  const pct = (v) => Math.max(0, Math.min(1, (v - lo) / (hi - lo))) * 100;
  // Good range: from whichever side of target is "good"
  const goodStart = worseHigher ? 0 : pct(target);
  const goodEnd = worseHigher ? pct(target) : 100;
  // Is current value in the good range?
  const isGood = worseHigher ? value <= target : value >= target;

  return (
    <div style={{ width }}>
      <div style={{ position: 'relative', height, marginTop: 4 }}>
        {/* baseline track */}
        <div style={{
          position: 'absolute', top: height/2 - 2, left: 0, right: 0, height: 4,
          background: 'var(--sdt-track)', borderRadius: 2,
        }} />
        {/* good range */}
        <div style={{
          position: 'absolute', top: height/2 - 2,
          left: `${goodStart}%`, width: `${goodEnd - goodStart}%`, height: 4,
          background: 'color-mix(in oklch, var(--sdt-up) 38%, transparent)',
          borderRadius: 2,
        }} />
        {/* target tick */}
        <div style={{
          position: 'absolute', top: 2, left: `${pct(target)}%`,
          width: 2, height: height - 4, background: 'var(--sdt-fg-muted)',
          transform: 'translateX(-1px)',
        }} />
        {/* current marker */}
        <div style={{
          position: 'absolute', top: 0, left: `${pct(value)}%`,
          transform: 'translateX(-50%)',
        }}>
          <div style={{
            width: 12, height: height, background: isGood ? 'var(--sdt-up)' : 'var(--sdt-down)',
            borderRadius: 2, border: '2px solid var(--sdt-bg)',
            boxShadow: '0 0 0 1px ' + (isGood ? 'var(--sdt-up)' : 'var(--sdt-down)'),
          }} />
        </div>
      </div>
      <div style={{
        display: 'flex', justifyContent: 'space-between', marginTop: 6,
        fontSize: 10, fontVariantNumeric: 'tabular-nums', color: 'var(--sdt-fg-muted)',
        letterSpacing: 0.3,
      }}>
        <span>{lo.toFixed(lo % 1 === 0 ? 0 : 1)}{unit}</span>
        <span style={{ color: 'var(--sdt-fg)', fontWeight: 500 }}>Mål {target}{unit}</span>
        <span>{hi.toFixed(hi % 1 === 0 ? 0 : 1)}{unit}</span>
      </div>
    </div>
  );
}

