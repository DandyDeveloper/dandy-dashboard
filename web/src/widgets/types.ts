import type { Component } from 'svelte'

// WidgetSize controls how many grid columns the widget occupies.
export type WidgetSize = 'sm' | 'md' | 'lg' | 'xl'

// WidgetDescriptor is the contract every dashboard widget must satisfy.
// To add a new widget:
//   1. Create src/widgets/<name>/<Name>Widget.svelte
//   2. Create src/widgets/<name>/index.ts exporting a WidgetDescriptor
//   3. Add one line to src/widgets/registry.ts
export interface WidgetDescriptor {
  /** Must match the backend widget's Slug() value */
  id: string
  title: string
  description: string
  /** The Svelte component to render inside the widget card */
  component: Component
  /** Default grid column span on desktop */
  defaultSize: WidgetSize
}
