# Design System Specification: Nordic Transparency & Democratic Authority

## 1. Overview & Creative North Star: "The Digital Atrium"
This design system is built upon the concept of **The Digital Atrium**. In classical architecture, an atrium is a central, open space that provides light, ventilation, and a sense of shared community. For a Swedish democratic platform, this translates to a UI that feels open, unshakeable, and illuminated.

We reject the "boxed-in" nature of traditional government portals. Instead of rigid grids and heavy borders, we utilize **Intentional Asymmetry** and **Tonal Depth**. The aesthetic merges "Nordic Minimalism" (functional, airy, honest) with "Data Transparency" (precise, layered, informative). We move away from flat UI into a world of "Editorial Precision," where typography does the heavy lifting and space is used as a functional tool rather than a void.

---

## 2. Color Philosophy: Refined Heritage
The palette evolves the Swedish flag into a sophisticated, professional spectrum. We move from primary yellow to a muted "Metallic Gold" and from bright blue to a "Midnight Navy."

### The "No-Line" Rule
**Explicit Instruction:** 1px solid borders are strictly prohibited for sectioning or defining layout boundaries. Boundaries must be defined solely through background color shifts.
* *Implementation:* Place a `surface_container_low` section directly against a `surface` background. The shift in tone is the divider.

### Surface Hierarchy & Nesting
Treat the UI as a series of physical layers—like stacked sheets of fine Nordic paper.
* **Base Layer:** `surface` (#f8f9fb)
* **Secondary Content Areas:** `surface_container_low` (#f2f4f6)
* **Interactive Cards:** `surface_container_lowest` (#ffffff)
* **Information Callouts:** `surface_container_high` (#e6e8ea)

### The Glass & Gradient Rule
To prevent a "template" feel, use **Glassmorphism** for floating navigation or overlays.
* *Style:* Apply `surface_variant` at 70% opacity with a `24px` backdrop blur.
* **Signature Textures:** For primary CTAs or Hero sections, use a subtle linear gradient: `primary` (#003461) to `primary_container` (#004b87) at a 135-degree angle. This adds a "soul" and depth that flat color cannot replicate.

---

## 3. Typography: The Editorial Voice
We pair the structural precision of **Manrope** for high-level communication with the neutral clarity of **Inter** for data and body content.

* **Display (Manrope):** Used for large-scale data points or high-impact democratic statements. The scale (up to `3.5rem`) creates an authoritative, editorial feel.
* **Headline (Manrope):** Set with tight letter-spacing (-0.02em) to feel "heavy" and trustworthy.
* **Body & Labels (Inter):** The workhorse. Used for policy text and data labels. Inter’s high x-height ensures readability in complex democratic documents.
* **Hierarchy Note:** Use `on_surface_variant` (#424750) for secondary metadata to create a clear visual distinction from primary headers in `on_surface` (#191c1e).

---

## 4. Elevation & Depth: Tonal Layering
We convey importance through "lift" rather than "lines."

* **The Layering Principle:** Depth is achieved by "stacking." A `surface_container_lowest` card placed on a `surface_container_low` background creates a soft, natural lift without the need for a shadow.
* **Ambient Shadows:** If a floating element (like a modal) is required, use an extra-diffused shadow:
* *Values:* `0px 20px 40px`
* *Color:* `on_surface` at 6% opacity. Never use pure black or grey; always tint the shadow with the surface color.
* **The Ghost Border Fallback:** If a border is required for accessibility (e.g., input fields), use the `outline_variant` token at 20% opacity. **Forbid 100% opaque borders.**
* **Glassmorphism Depth:** Use backdrop blurs on `surface_container_lowest` with 80% opacity to allow democratic data to feel integrated into the environment.

---

## 5. Components: The Building Blocks of Trust

### Buttons & Interaction
* **Primary:** A gradient of `primary` to `primary_container`. No border. Roundedness: `md` (0.375rem).
* **Secondary:** `primary_fixed` background with `on_primary_fixed` text.
* **Tertiary:** No background. `on_surface` text with an underline that appears only on hover.

### Input Fields & Forms
* **Structure:** No heavy boxes. Use `surface_container_highest` as a subtle background fill.
* **States:** On focus, the background shifts to `surface_container_lowest` with a `surface_tint` "Ghost Border" (20% opacity).

### Cards & Lists (The "No-Divider" Rule)
* **Forbid Divider Lines:** Separate list items using the Spacing Scale (e.g., `spacing.4` or `1rem` vertical gaps).
* **Cards:** Use `surface_container_low` for the card body. Distinguish content types by shifting the background of the header or footer of the card to `surface_container_high`.

### Data-Specific Components
* **Metric Tiles:** Large `display-sm` numbers in `primary`. Pair with `label-md` text in `on_tertiary_fixed_variant` for a "Data Transparency" look.
* **Progress Bars:** Use `secondary_container` (#fecc00) for the track and `secondary` (#735c00) for the fill to reference the Swedish gold in a functional, non-decorative way.

---

## 6. Do's and Don'ts

### Do
* **Do** use asymmetrical margins (e.g., a wider left margin for headlines) to create an editorial, bespoke feel.
* **Do** prioritize "Breathing Room." Use `spacing.16` (4rem) between major sections to emphasize transparency and calm.
* **Do** use `tertiary` colors for deep-dive technical data or "expert-level" information.

### Don't
* **Don't** use 1px solid black or grey borders. They feel "cheap" and bureaucratic.
* **Don't** use standard Material Design drop shadows. They lack the Nordic "soft-light" quality.
* **Don't** center-align long blocks of text. Stick to left-aligned editorial layouts for maximum authority.
* **Don't** use fully opaque backgrounds for navigation bars; always use the Glassmorphism rule to maintain the "Atrium" feel.

---

## 7. Spacing & Rhythm
Use the geometric spacing scale to ensure mathematical harmony.
* **Containers:** Use `spacing.6` (1.5rem) internal padding.
* **Section Gaps:** Use `spacing.20` (5rem) to separate major democratic modules.
* **Typography Lead:** Use `title-sm` with `spacing.2` (0.5rem) margin-bottom to anchor body text.
