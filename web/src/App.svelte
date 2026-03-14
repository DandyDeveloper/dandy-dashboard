<script lang="ts">
  import { widgetRegistry } from './widgets/registry'
  import WidgetCard from './components/WidgetCard.svelte'
</script>

<div class="min-h-screen bg-surface-0">
  <!-- Header -->
  <header class="border-b border-white/[0.06] bg-surface-1/50 backdrop-blur-md sticky top-0 z-10">
    <div class="max-w-7xl mx-auto px-6 h-14 flex items-center justify-between">
      <div class="flex items-center gap-3">
        <div class="w-7 h-7 rounded-lg bg-accent flex items-center justify-center">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
            <rect x="3" y="3" width="7" height="7" rx="1"/>
            <rect x="14" y="3" width="7" height="7" rx="1"/>
            <rect x="3" y="14" width="7" height="7" rx="1"/>
            <rect x="14" y="14" width="7" height="7" rx="1"/>
          </svg>
        </div>
        <span class="font-semibold text-gray-100 tracking-tight">Dandy Dashboard</span>
      </div>

      <time class="text-sm text-gray-500 tabular-nums">
        {new Date().toLocaleDateString('en-US', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })}
      </time>
    </div>
  </header>

  <!-- Widget grid -->
  <main class="max-w-7xl mx-auto px-6 py-8">
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 auto-rows-min">
      {#each widgetRegistry as descriptor (descriptor.id)}
        <WidgetCard widget={descriptor}>
          <descriptor.component />
        </WidgetCard>
      {/each}
    </div>
  </main>
</div>
