import type { WidgetDescriptor } from '../types'
import TasksWidget from './TasksWidget.svelte'

export const tasksWidget: WidgetDescriptor = {
  id: 'tasks',
  title: 'Tasks',
  description: 'Lightweight task tracker with status, priority, and due dates.',
  component: TasksWidget,
  defaultSize: 'lg',
}
