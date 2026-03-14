<script lang="ts">
  import { Calendar, MapPin, Clock, ChevronRight } from 'lucide-svelte'

  interface CalEvent {
    id: string
    title: string
    start: string
    end: string
    all_day: boolean
    location?: string
    description?: string
    color?: string
  }

  let events = $state<CalEvent[]>([])
  let loading = $state(true)
  let error = $state<string | null>(null)

  async function fetchEvents() {
    loading = true
    error = null
    try {
      const resp = await fetch('/api/widgets/calendar/events?days=7')
      if (!resp.ok) throw new Error(`HTTP ${resp.status}`)
      const data = await resp.json()
      events = data.events ?? []
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load events'
    } finally {
      loading = false
    }
  }

  $effect(() => {
    fetchEvents()
  })

  function formatDate(iso: string, allDay: boolean): string {
    const d = new Date(iso)
    if (allDay) {
      return d.toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' })
    }
    return d.toLocaleString('en-US', {
      weekday: 'short',
      month: 'short',
      day: 'numeric',
      hour: 'numeric',
      minute: '2-digit',
    })
  }

  function formatTime(iso: string): string {
    return new Date(iso).toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' })
  }

  function isToday(iso: string): boolean {
    const d = new Date(iso)
    const now = new Date()
    return d.getFullYear() === now.getFullYear() &&
      d.getMonth() === now.getMonth() &&
      d.getDate() === now.getDate()
  }

  function isTomorrow(iso: string): boolean {
    const d = new Date(iso)
    const tomorrow = new Date()
    tomorrow.setDate(tomorrow.getDate() + 1)
    return d.getFullYear() === tomorrow.getFullYear() &&
      d.getMonth() === tomorrow.getMonth() &&
      d.getDate() === tomorrow.getDate()
  }

  function dayLabel(iso: string, allDay: boolean): string {
    if (isToday(iso)) return 'Today'
    if (isTomorrow(iso)) return 'Tomorrow'
    return formatDate(iso, allDay)
  }

  // Color mapping for Google Calendar color IDs
  const colorMap: Record<string, string> = {
    '1': '#ac725e', '2': '#d06b64', '3': '#f83a22', '4': '#fa573c',
    '5': '#ff7537', '6': '#ffad46', '7': '#42d692', '8': '#16a765',
    '9': '#7bd148', '10': '#b3dc6c', '11': '#fbe983', 'default': '#6366f1',
  }

  function eventColor(id?: string): string {
    return colorMap[id ?? ''] ?? colorMap['default']
  }
</script>

{#if loading}
  <div class="space-y-3 animate-pulse">
    {#each [1,2,3] as _}
      <div class="flex gap-3">
        <div class="w-1 rounded-full bg-white/10 self-stretch"></div>
        <div class="flex-1 space-y-1.5">
          <div class="h-4 w-3/4 bg-white/5 rounded"></div>
          <div class="h-3 w-1/2 bg-white/5 rounded"></div>
        </div>
      </div>
    {/each}
  </div>
{:else if error}
  <p class="text-sm text-red-400">{error}</p>
{:else if events.length === 0}
  <div class="flex flex-col items-center py-8 gap-2 text-gray-500">
    <Calendar size={28} class="opacity-40" />
    <p class="text-sm">No upcoming events</p>
  </div>
{:else}
  <div class="space-y-2">
    {#each events as event (event.id)}
      <div class="group flex gap-3 p-3 rounded-xl hover:bg-white/[0.04] transition-colors animate-slide-up">
        <!-- Colour stripe -->
        <div
          class="w-1 rounded-full shrink-0 self-stretch"
          style="background-color: {eventColor(event.color)}"
        ></div>

        <!-- Content -->
        <div class="flex-1 min-w-0">
          <div class="flex items-start justify-between gap-2">
            <p class="text-sm font-medium text-gray-100 leading-snug truncate">{event.title}</p>
            <ChevronRight size={14} class="text-gray-600 shrink-0 mt-0.5 opacity-0 group-hover:opacity-100 transition-opacity" />
          </div>

          <div class="mt-1 flex flex-wrap items-center gap-x-3 gap-y-0.5">
            <span class="flex items-center gap-1 text-xs text-gray-500">
              <Clock size={11} />
              {#if event.all_day}
                {dayLabel(event.start, true)} · All day
              {:else}
                {dayLabel(event.start, false)} · {formatTime(event.start)}–{formatTime(event.end)}
              {/if}
            </span>
            {#if event.location}
              <span class="flex items-center gap-1 text-xs text-gray-500 truncate max-w-[200px]">
                <MapPin size={11} />
                {event.location}
              </span>
            {/if}
          </div>
        </div>
      </div>
    {/each}
  </div>
{/if}
