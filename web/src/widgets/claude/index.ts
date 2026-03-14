import type { WidgetDescriptor } from '../types'
import ClaudeWidget from './ClaudeWidget.svelte'

export const claudeWidget: WidgetDescriptor = {
  id: 'claude',
  title: 'Claude AI',
  description: 'Chat with Claude, your personal AI assistant.',
  component: ClaudeWidget,
  defaultSize: 'xl',
}
