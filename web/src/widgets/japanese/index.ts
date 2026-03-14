import type { WidgetDescriptor } from '../types'
import JapaneseWidget from './JapaneseWidget.svelte'

export const japaneseWidget: WidgetDescriptor = {
  id: 'japanese',
  title: 'Word of the Day',
  description: 'Daily Japanese vocabulary with readings and example sentences.',
  component: JapaneseWidget,
  defaultSize: 'md',
}
