<script lang="ts">
  import type { WidgetDescriptor, WidgetSize } from '../widgets/types'
  import type { Snippet } from 'svelte'

  interface Props {
    widget: WidgetDescriptor
    children: Snippet
  }

  let { widget, children }: Props = $props()

  // Icon mapping by widget id
  import { Bot, BookOpen, Calendar } from 'lucide-svelte'

  const iconMap: Record<string, typeof Bot> = {
    claude: Bot,
    japanese: BookOpen,
    calendar: Calendar,
  }

  const Icon = $derived(iconMap[widget.id] ?? BookOpen)

  const sizeClasses: Record<WidgetSize, string> = {
    sm: 'col-span-1',
    md: 'col-span-1 lg:col-span-1',
    lg: 'col-span-1 lg:col-span-2',
    xl: 'col-span-1 lg:col-span-2',
  }
</script>

<div class="widget-card flex flex-col h-full {sizeClasses[widget.defaultSize]}">
  <div class="widget-header">
    <div class="w-7 h-7 rounded-lg bg-accent/15 flex items-center justify-center text-accent">
      <Icon size={15} />
    </div>
    <span class="widget-title">{widget.title}</span>
  </div>
  <div class="widget-body flex-1">
    {@render children()}
  </div>
</div>
