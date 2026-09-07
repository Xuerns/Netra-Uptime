---
name: PulseWatch Dual-Mode Design System
modes:
  - dark
  - light
tokens:
  dark:
    colors:
      surface: '#131314'
      surface-dim: '#131314'
      surface-bright: '#3a393a'
      surface-container-lowest: '#0e0e0f'
      surface-container-low: '#1c1b1c'
      surface-container: '#201f20'
      surface-container-high: '#2a2a2b'
      surface-container-highest: '#353436'
      on-surface: '#e5e2e3'
      on-surface-variant: '#bbcabf'
      inverse-surface: '#e5e2e3'
      inverse-on-surface: '#313031'
      outline: '#86948a'
      outline-variant: '#3c4a42'
      surface-tint: '#4edea3'
      primary: '#4edea3'
      on-primary: '#003824'
      primary-container: '#10b981'
      on-primary-container: '#00422b'
      inverse-primary: '#006c49'
      secondary: '#adc6ff'
      on-secondary: '#002e6a'
      secondary-container: '#0566d9'
      on-secondary-container: '#e6ecff'
      tertiary: '#d0bcff'
      on-tertiary: '#3c0091'
      tertiary-container: '#b090ff'
      on-tertiary-container: '#4600a7'
      error: '#ffb4ab'
      on-error: '#690005'
      error-container: '#93000a'
      on-error-container: '#ffdad6'
      primary-fixed: '#6ffbbe'
      primary-fixed-dim: '#4edea3'
      on-primary-fixed: '#002113'
      on-primary-fixed-variant: '#005236'
      secondary-fixed: '#d8e2ff'
      secondary-fixed-dim: '#adc6ff'
      on-secondary-fixed: '#001a42'
      on-secondary-fixed-variant: '#004395'
      tertiary-fixed: '#e9ddff'
      tertiary-fixed-dim: '#d0bcff'
      on-tertiary-fixed: '#23005c'
      on-tertiary-fixed-variant: '#5516be'
      background: '#0A0A0B'
      on-background: '#e5e2e3'
      surface-variant: '#353436'
      # Extended Glass & Border Tokens
      glass-card-bg: 'rgba(255, 255, 255, 0.03)'
      glass-card-border: 'rgba(255, 255, 255, 0.08)'
      glass-modal-bg: 'rgba(255, 255, 255, 0.06)'
      glass-modal-border: 'rgba(255, 255, 255, 0.12)'
      input-bg: 'rgba(0, 0, 0, 0.25)'
      input-border: 'rgba(255, 255, 255, 0.1)'
      chart-grid: 'rgba(255, 255, 255, 0.08)'
  light:
    colors:
      surface: '#FFFFFF'
      surface-dim: '#F8FAFC'
      surface-bright: '#FFFFFF'
      surface-container-lowest: '#FFFFFF'
      surface-container-low: '#F8FAFC'
      surface-container: '#F1F5F9'
      surface-container-high: '#E2E8F0'
      surface-container-highest: '#CBD5E1'
      on-surface: '#0F172A'
      on-surface-variant: '#475569'
      inverse-surface: '#1E293B'
      inverse-on-surface: '#F8FAFC'
      outline: '#94A3B8'
      outline-variant: '#E2E8F0'
      surface-tint: '#10B981'
      primary: '#10B981'
      on-primary: '#FFFFFF'
      primary-container: '#ECFDF5'
      on-primary-container: '#065F46'
      inverse-primary: '#4edea3'
      secondary: '#2563EB'
      on-secondary: '#FFFFFF'
      secondary-container: '#EFF6FF'
      on-secondary-container: '#1E40AF'
      tertiary: '#7C3AED'
      on-tertiary: '#FFFFFF'
      tertiary-container: '#F5F3FF'
      on-tertiary-container: '#5B21B6'
      error: '#EF4444'
      on-error: '#FFFFFF'
      error-container: '#FEF2F2'
      on-error-container: '#991B1B'
      primary-fixed: '#D1FAE5'
      primary-fixed-dim: '#A7F3D0'
      on-primary-fixed: '#064E3B'
      on-primary-fixed-variant: '#047857'
      secondary-fixed: '#DBEAFE'
      secondary-fixed-dim: '#BFDBFE'
      on-secondary-fixed: '#1E3A8A'
      on-secondary-fixed-variant: '#1D4ED8'
      tertiary-fixed: '#EDE9FE'
      tertiary-fixed-dim: '#DDD6FE'
      on-tertiary-fixed: '#4C1D95'
      on-tertiary-fixed-variant: '#6D28D9'
      background: '#F8FAFC'
      on-background: '#0F172A'
      surface-variant: '#F1F5F9'
      # Extended Card & Border Tokens
      glass-card-bg: '#FFFFFF'
      glass-card-border: '#E2E8F0'
      glass-modal-bg: '#FFFFFF'
      glass-modal-border: '#CBD5E1'
      input-bg: '#FFFFFF'
      input-border: '#E2E8F0'
      chart-grid: '#E2E8F0'
typography:
  display-lg:
    fontFamily: Inter
    fontSize: 48px
    fontWeight: '700'
    lineHeight: '1.1'
    letterSpacing: -0.02em
  headline-lg:
    fontFamily: Inter
    fontSize: 32px
    fontWeight: '600'
    lineHeight: '1.2'
    letterSpacing: -0.01em
  headline-lg-mobile:
    fontFamily: Inter
    fontSize: 24px
    fontWeight: '600'
    lineHeight: '1.2'
  headline-md:
    fontFamily: Inter
    fontSize: 24px
    fontWeight: '600'
    lineHeight: '1.3'
  body-lg:
    fontFamily: Inter
    fontSize: 18px
    fontWeight: '400'
    lineHeight: '1.6'
  body-md:
    fontFamily: Inter
    fontSize: 16px
    fontWeight: '400'
    lineHeight: '1.6'
  label-md:
    fontFamily: Inter
    fontSize: 14px
    fontWeight: '500'
    lineHeight: '1.4'
    letterSpacing: 0.01em
  mono-data:
    fontFamily: JetBrains Mono
    fontSize: 14px
    fontWeight: '450'
    lineHeight: '1.5'
rounded:
  sm: 0.25rem
  DEFAULT: 0.5rem
  md: 0.75rem
  lg: 1rem
  xl: 1.5rem
  full: 9999px
spacing:
  base: 4px
  xs: 8px
  sm: 16px
  md: 24px
  lg: 40px
  xl: 64px
  container-max: 1440px
  gutter: 24px
---

# PulseWatch Dual-Mode Design System (`design.md`)

## 1. Brand & Style Philosophy
PulseWatch is a high-performance cloud infrastructure and uptime monitoring workspace. The brand personality is **vigilant, technical, and premium**, evoking a sense of "mission control" across both operating modes:

- **Dark Mode ("Mission Control Glassmorphism")**: Engineered to reduce eye strain during prolonged monitoring sessions. Features deep obsidian surfaces (`#0A0A0B`), translucent frosted layers, and glowing neon telemetry lines to simulate a multi-layered digital cockpit.
- **Light Mode ("Precision Engineering")**: Optimized for daylight operation, executive reporting, and rapid incident triage. Delivers crisp contrast with an airy slate canvas (`#F8FAFC`), pure white elevated cards (`#FFFFFF`), subtle dot-matrix micro-textures, and sharp border delineation.

---

## 2. Color System & Dual-Mode Palette

### Mode Comparison Matrix

| Token Role | Dark Mode (Glassmorphic) | Light Mode (Clean Slate) | Usage & Context |
| :--- | :--- | :--- | :--- |
| **Canvas / Background** | `#0A0A0B` (Deep Obsidian) | `#F8FAFC` (Slate 50) + Dot Grid | Viewport & master frame |
| **Primary Surface / Cards** | `rgba(255, 255, 255, 0.03)` / `#201f20` | `#FFFFFF` (Pure White) | Monitor cards, incident containers |
| **Elevated / Modals** | `rgba(255, 255, 255, 0.06)` / `#2a2a2b` | `#FFFFFF` with soft shadow | Command room drawers, popovers |
| **Borders & Outlines** | `rgba(255, 255, 255, 0.08)` / `#3c4a42` | `#E2E8F0` (Slate 200, 1px) | Card strokes, dividing rules |
| **Primary (Uptime Green)** | `#4edea3` (Neon) / `#10B981` (Emerald) | `#10B981` / `#059669` (Deep Emerald) | Healthy states, active sparklines, CTAs |
| **Secondary (Data Blue)** | `#adc6ff` / `#3B82F6` | `#2563EB` (Royal Blue) / `#0284C7` | Latency pings, network metrics, focus rings |
| **Warning / Degraded** | `#F59E0B` / `#FCD34D` | `#D97706` / `#F59E0B` (`#FFFBEB` bg) | Latency spikes, warning flags |
| **Critical / Outage** | `#ffb4ab` / `#EF4444` | `#DC2626` / `#EF4444` (`#FEF2F2` bg) | Downtime, error codes, failed checks |
| **Primary Text** | `#FFFFFF` / `#E5E2E3` | `#0F172A` (Slate 900) | Main headings, monitor names |
| **Secondary Text** | `#94A3B8` / `#BBCABF` | `#475569` (Slate 600) | Subtitles, secondary labels |
| **Monospace / Metrics** | `#4edea3` / `#E2E8F0` | `#0F172A` / `#065F46` | Latency numbers, IPs, commit hashes |

---

## 3. Typography Hierarchy

Consistent typography scales bridge both dark and light modes, relying on **Inter** for clean user interfaces and **JetBrains Mono** for telemetry data.

- **Primary Font Family**: `Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif`
- **Monospace Family**: `"JetBrains Mono", ui-monospace, SFMono-Regular, Menlo, monospace`

### Typographic Scales
- **Display Large**: `48px` / `1.1`, font-weight `700`, tracking `-0.02em`
- **Headline Large**: `32px` / `1.2`, font-weight `600`, tracking `-0.01em` (mobile: `24px`)
- **Headline Medium**: `24px` / `1.3`, font-weight `600`
- **Card Title / Section Header**: `16px` / `1.4`, font-weight `600`
- **Body Large**: `18px` / `1.6`, font-weight `400`
- **Body Medium**: `16px` / `1.6`, font-weight `400` / `500`
- **Label / Metadata**: `14px` / `1.4`, font-weight `500`, tracking `0.01em`
- **Mono Data & Telemetry**: `14px` / `1.5`, font-weight `450` (tabular figures)
- **Micro Badges**: `11px` / `1.2`, font-weight `600`, uppercase, tracking `0.05em`

---

## 4. Layout, Grid & Spatial Rhythm

- **Spatial Rhythm**: Multiples of `4px` (`4px`, `8px`, `16px`, `24px`, `40px`, `64px`)
- **Fluid Grid**: 12-column responsive layout:
  - **Desktop**: 24px margins, 24px gutters, max-width `1440px` (or `1600px` for wide monitoring cockpits). Sidebar width: `260px` fixed.
  - **Tablet**: 16px margins, compact icon-only collapsible sidebar.
  - **Mobile**: 16px margins, single-column stacked card flow with `32px` vertical spacing.
- **Corner Radii (`rounded`)**:
  - `sm`: `0.25rem` (4px) — micro badges, tags
  - `DEFAULT`: `0.5rem` (8px) — buttons, inputs, segmented controls
  - `md`: `0.75rem` (12px) — standard widgets, sub-cards
  - `lg`: `1rem` (16px) — glass containers & primary dashboard cards
  - `full`: `9999px` — status pills, avatar circles, filter counters

---

## 5. Elevation, Depth & Mode Treatment

### Dark Mode Elevation (Backdrop Blurs & Tonal Stacking)
1. **Level 0 (Base Canvas)**: Solid `#0A0A0B`.
2. **Level 1 (Glass Cards)**: `background: rgba(255, 255, 255, 0.03)` with `backdrop-filter: blur(12px)`. Border: `1px solid rgba(255, 255, 255, 0.08)`. Subtle top-to-bottom sheen gradient: `linear-gradient(180deg, rgba(255, 255, 255, 0.02) 0%, transparent 100%)`.
3. **Level 2 (Modals / Focus / Tooltips)**: `background: rgba(255, 255, 255, 0.06)` with `backdrop-filter: blur(20px)` and subtle outer glow `rgba(16, 185, 129, 0.08)`.

### Light Mode Elevation (Natural Contrast & Subtle Shadows)
1. **Level 0 (Base Canvas)**: Solid `#F8FAFC` with subtle dot-matrix texture (`radial-gradient(#CBD5E1 1px, transparent 1px)` with `size: 20px 20px`).
2. **Level 1 (Cards)**: Pure `#FFFFFF` with `1px solid #E2E8F0` and subtle elevation shadow: `0 1px 3px 0 rgba(0, 0, 0, 0.04), 0 1px 2px -1px rgba(0, 0, 0, 0.04)`.
3. **Level 2 (Modals / Popovers)**: Pure `#FFFFFF` with `0 20px 25px -5px rgba(0, 0, 0, 0.08)` and border `1px solid #CBD5E1`.

---

## 6. Component Patterns & Specifications

### 1. Primary & Secondary Buttons
- **Dark Mode**:
  - Primary: `#10B981` background, `#003824` / `#000000` text, subtle green halo on hover.
  - Secondary / Ghost: 1px border `rgba(255, 255, 255, 0.1)`, transparent background, hover brings `rgba(255, 255, 255, 0.06)`.
- **Light Mode**:
  - Primary: `#10B981` background, `#FFFFFF` text, hover `#059669`.
  - Secondary: `#FFFFFF` background, 1px border `#E2E8F0`, text `#334155`, hover `#F8FAFC`.

### 2. Status Indicators & Badges
- **Live Pulse**:
  - Dark Mode: 8px circle with slow-pulsing `box-shadow: 0 0 12px rgba(78, 222, 163, 0.6)`.
  - Light Mode: 8px circle in solid `#10B981` with soft emerald halo `rgba(16, 185, 129, 0.25)`.
- **Status Pills (`UP`, `DOWN`, `PAUSED`)**:
  - Operational: Full pill radius (`rounded-full`), text `#10B981` with matching tinted background (`rgba(16, 185, 129, 0.12)` in Dark, `#ECFDF5` in Light) and border (`rgba(16, 185, 129, 0.25)` in Dark, `#A7F3D0` in Light).
  - Degraded / Warning: Amber tint (`#FFFBEB` in Light / `rgba(245, 158, 11, 0.12)` in Dark).
  - Critical Outage: Rose tint (`#FEF2F2` in Light / `rgba(239, 68, 68, 0.15)` in Dark).

### 3. Inputs & Search Fields
- **Dark Mode**: Background `rgba(0, 0, 0, 0.25)`, border `1px solid rgba(255, 255, 255, 0.1)`, placeholder `#64748B`. Focus transitions border to `#3B82F6` with 4px soft outer glow.
- **Light Mode**: Background `#FFFFFF`, border `1px solid #E2E8F0`, placeholder `#94A3B8`. Focus transitions border to `#10B981` or `#3B82F6` with soft tint shadow.

### 4. Uptime Graphs & Sparklines
- **Interpolation**: Stepped or crisp multi-segment line rendering to capture discrete latency and status intervals.
- **Stroke & Area Fill**:
  - Dark Mode: Stroke `#4edea3` (2px), gradient area fill `linear-gradient(180deg, rgba(78, 222, 163, 0.2) 0%, rgba(78, 222, 163, 0.0) 100%)`. Grid lines in `rgba(255, 255, 255, 0.06)`.
  - Light Mode: Stroke `#10B981` (2px), gradient area fill `linear-gradient(180deg, rgba(16, 185, 129, 0.15) 0%, rgba(16, 185, 129, 0.0) 100%)`. Grid lines in `#E2E8F0` dashed (`4 4`).

### 5. Navigation Sidebar
- Fixed `260px` left rail with brand logo mark, grouped navigation (`MONITORING`, `ORGANIZATION`), and user profile trigger.
- **Active State**:
  - Dark Mode: `background: rgba(78, 222, 163, 0.12)`, text `#4edea3`, font-weight `600`.
  - Light Mode: `background: #ECFDF5`, text `#065F46`, font-weight `600`.
