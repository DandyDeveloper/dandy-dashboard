import type { WidgetDescriptor } from '../types'
import CalendarWidget from './CalendarWidget.svelte'

export const calendarWidget: WidgetDescriptor = {
  id: 'calendar',
  title: 'Upcoming Events',
  description: 'Your next 7 days from Google Calendar.',
  component: CalendarWidget,
  defaultSize: 'md',
}
