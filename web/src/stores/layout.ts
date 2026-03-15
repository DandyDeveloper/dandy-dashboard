import type { WidgetDescriptor, WidgetSize } from '../widgets/types'

export interface WidgetLayout {
  colSpan: number // 1–12
  rowSpan: number // 1–n  (each row = ROW_HEIGHT px + gap)
}

export type LayoutMap = Record<string, WidgetLayout>

const STORAGE_KEY = 'dashboard-layout-v1'

const SIZE_DEFAULTS: Record<WidgetSize, WidgetLayout> = {
  sm: { colSpan: 4, rowSpan: 3 },
  md: { colSpan: 6, rowSpan: 4 },
  lg: { colSpan: 8, rowSpan: 5 },
  xl: { colSpan: 12, rowSpan: 6 },
}

export function defaultLayout(widgets: WidgetDescriptor[]): LayoutMap {
  return Object.fromEntries(
    widgets.map(w => [w.id, SIZE_DEFAULTS[w.defaultSize] ?? SIZE_DEFAULTS.md])
  )
}

export function loadLayout(widgets: WidgetDescriptor[]): LayoutMap {
  const defaults = defaultLayout(widgets)
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return defaults
    // Merge stored values with defaults so newly-added widgets always appear.
    return { ...defaults, ...JSON.parse(raw) }
  } catch {
    return defaults
  }
}

export function saveLayout(layout: LayoutMap): void {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(layout))
}

export function resetLayout(widgets: WidgetDescriptor[]): LayoutMap {
  localStorage.removeItem(STORAGE_KEY)
  return defaultLayout(widgets)
}
