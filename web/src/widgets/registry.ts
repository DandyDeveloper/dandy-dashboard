import type { WidgetDescriptor } from './types'
import { claudeWidget } from './claude'
import { japaneseWidget } from './japanese'
import { calendarWidget } from './calendar'

// widgetRegistry is the single source of truth for all dashboard modules.
// Widgets are rendered in the order they appear here.
export const widgetRegistry: WidgetDescriptor[] = [
  japaneseWidget,
  calendarWidget,
  claudeWidget,
]
