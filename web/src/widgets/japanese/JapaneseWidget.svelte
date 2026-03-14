<script lang="ts">
  import { BookOpen, Volume2, RefreshCw } from 'lucide-svelte'

  interface Example {
    japanese: string
    english: string
  }

  interface WordEntry {
    word: string
    reading: string
    meanings: string[]
    level: string
    examples: Example[]
    date: string
    source: 'wanikani' | 'wordlist'
  }

  let entry = $state<WordEntry | null>(null)
  let loading = $state(true)
  let error = $state<string | null>(null)

  async function fetchWord() {
    loading = true
    error = null
    try {
      const resp = await fetch('/api/widgets/japanese/word-of-day')
      if (!resp.ok) throw new Error(`HTTP ${resp.status}`)
      entry = await resp.json()
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load word'
    } finally {
      loading = false
    }
  }

  $effect(() => { fetchWord() })

  const sourceLabel: Record<string, string> = {
    wanikani: 'WaniKani',
    wordlist: 'Word List',
  }
  const sourceBadgeClass: Record<string, string> = {
    wanikani: 'bg-pink-500/20 text-pink-300',
    wordlist: 'bg-blue-500/20 text-blue-300',
  }
</script>

{#if loading}
  <div class="space-y-3 animate-pulse">
    <div class="h-12 w-32 bg-white/5 rounded-lg"></div>
    <div class="h-4 w-48 bg-white/5 rounded"></div>
    <div class="h-4 w-64 bg-white/5 rounded"></div>
    <div class="mt-4 space-y-2">
      <div class="h-3 w-full bg-white/5 rounded"></div>
      <div class="h-3 w-5/6 bg-white/5 rounded"></div>
    </div>
  </div>
{:else if error}
  <div class="flex flex-col items-center gap-3 py-4 text-center">
    <p class="text-sm text-red-400">{error}</p>
    <button onclick={fetchWord} class="btn-ghost text-xs flex items-center gap-1.5">
      <RefreshCw size={12} /> Retry
    </button>
  </div>
{:else if entry}
  <div class="space-y-5 animate-fade-in">
    <!-- Word + reading -->
    <div class="flex items-start justify-between gap-4">
      <div>
        <p class="font-japanese text-5xl font-bold text-gray-50 leading-none mb-2">{entry.word}</p>
        <p class="font-japanese text-lg text-accent">{entry.reading}</p>
      </div>
      <div class="flex flex-col items-end gap-2 shrink-0">
        <div class="flex items-center gap-2">
          {#if entry.level}
            <span class="tag bg-accent/20 text-accent font-semibold">{entry.level}</span>
          {/if}
          {#if entry.source}
            <span class="tag {sourceBadgeClass[entry.source] ?? 'bg-white/10 text-gray-400'} font-medium">
              {sourceLabel[entry.source] ?? entry.source}
            </span>
          {/if}
        </div>
        <button onclick={fetchWord} class="btn-ghost p-1.5" title="Refresh">
          <RefreshCw size={13} />
        </button>
      </div>
    </div>

    <!-- Meanings -->
    <div>
      <div class="flex items-center gap-2 mb-2">
        <BookOpen size={13} class="text-gray-500" />
        <span class="text-xs font-semibold text-gray-500 uppercase tracking-wider">Meanings</span>
      </div>
      <ul class="space-y-1">
        {#each entry.meanings.slice(0, 4) as meaning, i}
          <li class="text-sm text-gray-300 flex gap-2">
            <span class="text-gray-600 font-mono shrink-0">{i + 1}.</span>
            {meaning}
          </li>
        {/each}
      </ul>
    </div>

    <!-- Example sentences -->
    {#if entry.examples && entry.examples.length > 0}
      <div>
        <div class="flex items-center gap-2 mb-2">
          <Volume2 size={13} class="text-gray-500" />
          <span class="text-xs font-semibold text-gray-500 uppercase tracking-wider">Example Sentences</span>
        </div>
        <div class="space-y-3">
          {#each entry.examples as ex}
            <div class="pl-3 border-l-2 border-accent/30">
              <p class="font-japanese text-sm text-gray-200 leading-relaxed">{ex.japanese}</p>
              <p class="text-xs text-gray-500 mt-0.5 italic">{ex.english}</p>
            </div>
          {/each}
        </div>
      </div>
    {/if}

    <p class="text-xs text-gray-600">Word of the day · {entry.date}</p>
  </div>
{/if}
