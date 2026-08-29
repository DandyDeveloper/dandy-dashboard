import type { WidgetDescriptor } from './types'
import { tasksWidget } from './tasks'
import { japaneseWidget } from './japanese'
import { calendarWidget } from './calendar'

// widgetRegistry is the single source of truth for all dashboard modules.
// Widgets are rendered in the order they appear here.
export const widgetRegistry: WidgetDescriptor[] = [
  tasksWidget,
  japaneseWidget,
  calendarWidget,
]
