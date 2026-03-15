<script lang="ts">
  import type { WidgetDescriptor } from '../widgets/types'
  import WidgetCard from './WidgetCard.svelte'
  import { loadLayout, saveLayout, resetLayout } from '../stores/layout'
  import type { LayoutMap } from '../stores/layout'
  import { RotateCcw } from 'lucide-svelte'

  interface Props {
    widgets: WidgetDescriptor[]
  }

  const { widgets }: Props = $props()

  // Grid constants
  const COLS = 12
  const ROW_HEIGHT = 80  // px per row unit
  const GAP = 16         // px — must match the CSS gap below

  let container: HTMLElement
  // widgets prop is static (registry never changes at runtime), so
  // capturing the initial value for loadLayout is intentional.
  let layout = $state<LayoutMap>(loadLayout($state.snapshot(widgets) as WidgetDescriptor[]))

  // ── Resize state ──────────────────────────────────────────────────────────
  interface ResizeState {
    id: string
    startX: number
    startY: number
    startColSpan: number
    startRowSpan: number
  }
  let resizing = $state<ResizeState | null>(null)

  function startResize(e: PointerEvent, id: string) {
    e.preventDefault()
    ;(e.currentTarget as Element).setPointerCapture(e.pointerId)
    resizing = {
      id,
      startX: e.clientX,
      startY: e.clientY,
      startColSpan: layout[id].colSpan,
      startRowSpan: layout[id].rowSpan,
    }
  }

  function onPointerMove(e: PointerEvent) {
    if (!resizing) return
    const containerWidth = container.getBoundingClientRect().width
    const cellW = (containerWidth - (COLS - 1) * GAP) / COLS
    const cellH = ROW_HEIGHT + GAP

    const deltaCol = Math.round((e.clientX - resizing.startX) / (cellW + GAP))
    const deltaRow = Math.round((e.clientY - resizing.startY) / cellH)

    layout[resizing.id] = {
      colSpan: Math.max(2, Math.min(COLS, resizing.startColSpan + deltaCol)),
      rowSpan: Math.max(2, Math.min(10, resizing.startRowSpan + deltaRow)),
    }
  }

  function stopResize() {
    if (!resizing) return
    resizing = null
    saveLayout(layout)
  }

  function handleReset() {
    layout = resetLayout(widgets)
  }
</script>

<!-- Grid shell — pointer events bubble up here from resize handles -->
<div
  bind:this={container}
  role="region"
  aria-label="Dashboard widget grid"
  class="grid"
  style="grid-template-columns: repeat({COLS}, 1fr); grid-auto-rows: {ROW_HEIGHT}px; gap: {GAP}px; grid-auto-flow: dense;"
  onpointermove={onPointerMove}
  onpointerup={stopResize}
  onpointercancel={stopResize}
>
  {#each widgets as descriptor (descriptor.id)}
    {@const l = layout[descriptor.id] ?? { colSpan: 6, rowSpan: 4 }}

    <!-- Grid item wrapper — carries the span values and hosts the resize handle -->
    <div
      class="relative min-h-0 group"
      style="grid-column: auto / span {l.colSpan}; grid-row: auto / span {l.rowSpan};"
    >
      <WidgetCard widget={descriptor}>
        <descriptor.component />
      </WidgetCard>

      <!-- Resize handle — bottom-right corner, appears on hover -->
      <div
        role="separator"
        aria-label="Resize {descriptor.title}"
        class="absolute bottom-1.5 right-1.5 z-10 w-4 h-4 rounded-sm cursor-nwse-resize
               opacity-0 group-hover:opacity-100 transition-opacity duration-150
               flex items-end justify-end pb-0.5 pr-0.5"
        class:opacity-100={resizing?.id === descriptor.id}
        onpointerdown={(e) => startResize(e, descriptor.id)}
      >
        <!-- Three-dot grip icon -->
        <svg width="10" height="10" viewBox="0 0 10 10" fill="currentColor" class="text-gray-500">
          <circle cx="8" cy="8" r="1.5"/>
          <circle cx="4" cy="8" r="1.5"/>
          <circle cx="8" cy="4" r="1.5"/>
        </svg>
      </div>
    </div>
  {/each}
</div>

<!-- Reset button -->
<div class="mt-6 flex justify-end">
  <button onclick={handleReset} class="btn-ghost flex items-center gap-2 text-xs text-gray-500">
    <RotateCcw size={12} />
    Reset layout
  </button>
</div>
