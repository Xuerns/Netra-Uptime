---
name: Pulsewatch Design System
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
  background: '#131314'
  on-background: '#e5e2e3'
  surface-variant: '#353436'
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

## Brand & Style

The design system is engineered for a high-performance uptime and status monitoring platform. The brand personality is **vigilant, technical, and premium**. It evokes a sense of "mission control" through a sophisticated dark-mode aesthetic that prioritizes clarity and immediate information hierarchy.

The visual style is **Glassmorphic Modernism**. It leverages deep charcoal surfaces, translucent frosted layers, and vibrant neon accents to simulate a multi-layered digital cockpit. High-contrast typography ensures that critical status data is readable at a glance, while subtle background blurs and micro-gradients provide a sense of depth and luxury characteristic of high-end SaaS tools.

## Colors

The palette is anchored by a deep obsidian base to reduce eye strain during long-term monitoring. 

- **Primary (Uptime Green):** #10B981 is used exclusively for "Healthy" states, active monitoring, and primary calls to action.
- **Secondary (Monitoring Blue):** #3B82F6 represents data flow, active pings, and configuration settings.
- **Surface Tones:** The background is #0A0A0B. Surfaces use varying levels of transparency rather than flat greys to maintain the glassmorphic effect.
- **Contrast Accents:** Use high-vibrancy tints of the primary and secondary colors for glows and "pulse" animations to draw attention to real-time events.

## Typography

This design system utilizes **Inter** for its neutral, systematic character and exceptional legibility in data-dense environments. For technical metrics (latency, IP addresses, timestamps), **JetBrains Mono** is introduced to provide a distinct "developer-first" feel.

Typography should be treated with high contrast: use pure white (`#FFFFFF`) for headlines and a secondary silver-grey (`#94A3B8`) for body text to establish a clear hierarchy. Large display type should use tighter letter spacing for a more aggressive, modern appearance.

## Layout & Spacing

The layout follows a **12-column fluid grid** for the main dashboard view, allowing monitoring widgets to reflow based on screen real estate. 

- **Desktop:** 24px margins with 24px gutters. Content is housed in "Glass Containers" that can span 3, 4, 6, or 12 columns.
- **Tablet:** 16px margins. Sidebar collapses into a compact icon-only rail.
- **Mobile:** 16px margins. Grid collapses to a single column. Vertical spacing is increased to 32px between widgets to maintain clarity.

The spacing rhythm is based on a 4px baseline, ensuring that all elements align to a consistent mathematical grid.

## Elevation & Depth

Depth is not achieved through traditional shadows, but through **Backdrop Blurs and Tonal Stacking**.

1.  **Level 0 (Floor):** Pure #0A0A0B.
2.  **Level 1 (Card/Surface):** `rgba(255, 255, 255, 0.03)` background with a 12px Backdrop Blur. A 1px solid border at `rgba(255, 255, 255, 0.08)` defines the edge.
3.  **Level 2 (Modals/Popovers):** `rgba(255, 255, 255, 0.06)` background with a 20px Backdrop Blur. Add a subtle outer glow using the primary color at 5% opacity to indicate focus.

Always use a subtle top-down linear gradient (Transparent to `rgba(255, 255, 255, 0.02)`) on glass surfaces to simulate light hitting the "edge" of the pane.

## Shapes

The design system uses **Rounded (Level 2)** geometry to soften the technical nature of the data. 

- **Standard Cards:** 1rem (16px) corner radius.
- **Buttons & Inputs:** 0.5rem (8px) corner radius.
- **Status Chips:** Full pill-shape (999px) to differentiate them from interactive buttons.

This balance of rounded corners with sharp, high-contrast data visualization creates a friendly yet professional aesthetic.

## Components

### Buttons
Primary buttons use a solid #10B981 fill with black text for maximum contrast. Secondary buttons use a "ghost" style: a 1px border of `rgba(255, 255, 255, 0.1)` and a subtle hover state that increases the background opacity.

### Status Indicators
"Live" status is represented by a 8px circle with a `box-shadow` glow. For "Healthy" states, the glow should pulse slowly using the primary green.

### Glass Cards
All dashboard widgets must be contained in glass cards. These cards should have a subtle inner stroke (1px) to give the appearance of beveled glass. Title areas within cards should be separated by a 1px horizontal line with 10% opacity.

### Inputs
Search and configuration fields use a darker, recessed background (`rgba(0, 0, 0, 0.2)`) with a 1px border. On focus, the border color transitions to the Secondary Blue (#3B82F6) with a 4px soft outer glow.

### Uptime Graphs
Line charts should use a "Stepped" interpolation rather than smooth curves to emphasize the binary nature of uptime (Up/Down). Use a subtle vertical gradient fill below the line.