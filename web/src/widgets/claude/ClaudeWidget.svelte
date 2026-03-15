<script lang="ts">
  import { Send, Bot, User, RotateCcw } from 'lucide-svelte'

  interface Message {
    role: 'user' | 'assistant'
    content: string
  }

  // Stable session ID for the lifetime of this component instance.
  // The backend uses this to look up and persist conversation history.
  const sessionId = crypto.randomUUID()

  let messages = $state<Message[]>([])
  let input = $state('')
  let isStreaming = $state(false)
  let error = $state<string | null>(null)
  let scrollEl: HTMLDivElement

  async function sendMessage() {
    const text = input.trim()
    if (!text || isStreaming) return

    input = ''
    error = null
    messages = [...messages, { role: 'user', content: text }]

    // Add empty assistant message that we'll stream into.
    messages = [...messages, { role: 'assistant', content: '' }]
    const assistantIdx = messages.length - 1
    isStreaming = true

    try {
      const resp = await fetch('/api/widgets/claude/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          session_id: sessionId,
          message: text,
        }),
      })

      if (!resp.ok) {
        throw new Error(`HTTP ${resp.status}`)
      }

      const reader = resp.body!.getReader()
      const decoder = new TextDecoder()
      let buffer = ''

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        buffer += decoder.decode(value, { stream: true })
        const lines = buffer.split('\n')
        buffer = lines.pop() ?? ''

        for (const line of lines) {
          if (line.startsWith('data:')) {
            const raw = line.slice(5).trim()
            try {
              const parsed = JSON.parse(raw)
              if (parsed.text) {
                messages = messages.map((m, i) =>
                  i === assistantIdx ? { ...m, content: m.content + parsed.text } : m
                )
                scrollToBottom()
              }
              if (parsed.error) {
                error = parsed.error
              }
            } catch {
              // Partial JSON chunk — skip
            }
          }
        }
      }
    } catch (e) {
      error = e instanceof Error ? e.message : 'Unknown error'
      messages = messages.slice(0, -1) // remove empty assistant message
    } finally {
      isStreaming = false
    }
  }

  function scrollToBottom() {
    setTimeout(() => scrollEl?.scrollTo({ top: scrollEl.scrollHeight, behavior: 'smooth' }), 0)
  }

  function reset() {
    messages = []
    error = null
    input = ''
    fetch(`/api/widgets/claude/chat/${sessionId}`, { method: 'DELETE' })
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      sendMessage()
    }
  }
</script>

<div class="flex flex-col h-full min-h-0">
  <!-- Message list -->
  <div
    bind:this={scrollEl}
    class="flex-1 overflow-y-auto space-y-4 pr-1 min-h-0 max-h-[420px]"
  >
    {#if messages.length === 0}
      <div class="flex flex-col items-center justify-center h-32 text-gray-500 gap-2">
        <Bot size={32} class="opacity-40" />
        <p class="text-sm">Ask me anything…</p>
      </div>
    {/if}

    {#each messages as msg, i (i)}
      <div class="flex gap-3 animate-slide-up" class:flex-row-reverse={msg.role === 'user'}>
        <div class="shrink-0 w-7 h-7 rounded-full flex items-center justify-center
          {msg.role === 'assistant' ? 'bg-accent/20 text-accent' : 'bg-white/10 text-gray-300'}">
          {#if msg.role === 'assistant'}
            <Bot size={14} />
          {:else}
            <User size={14} />
          {/if}
        </div>
        <div class="max-w-[80%] rounded-2xl px-4 py-2.5 text-sm leading-relaxed
          {msg.role === 'assistant'
            ? 'bg-surface-2 text-gray-200 rounded-tl-sm'
            : 'bg-accent/20 text-gray-100 rounded-tr-sm'}">
          {#if msg.content}
            {msg.content}
          {:else if isStreaming && i === messages.length - 1}
            <span class="inline-flex gap-1">
              <span class="w-1 h-1 rounded-full bg-gray-400 animate-bounce" style="animation-delay:0ms"></span>
              <span class="w-1 h-1 rounded-full bg-gray-400 animate-bounce" style="animation-delay:150ms"></span>
              <span class="w-1 h-1 rounded-full bg-gray-400 animate-bounce" style="animation-delay:300ms"></span>
            </span>
          {/if}
        </div>
      </div>
    {/each}
  </div>

  {#if error}
    <p class="mt-2 text-xs text-red-400 px-1">{error}</p>
  {/if}

  <!-- Input area -->
  <div class="mt-4 flex gap-2 items-end">
    <textarea
      bind:value={input}
      onkeydown={handleKeydown}
      placeholder="Message Claude… (Enter to send, Shift+Enter for newline)"
      rows={1}
      class="input-base resize-none flex-1 leading-5"
      style="field-sizing: content; max-height: 120px;"
      disabled={isStreaming}
    ></textarea>

    <button onclick={sendMessage} disabled={!input.trim() || isStreaming} class="btn-primary shrink-0 h-9 w-9 flex items-center justify-center p-0">
      <Send size={15} />
    </button>

    {#if messages.length > 0}
      <button onclick={reset} class="btn-ghost shrink-0 h-9 w-9 flex items-center justify-center p-0" title="Clear chat">
        <RotateCcw size={14} />
      </button>
    {/if}
  </div>
</div>
