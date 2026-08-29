<script lang="ts">
  import {
    Plus, Trash2, ChevronDown, ChevronUp,
    Circle, Clock, CheckCircle2, X, Calendar,
  } from 'lucide-svelte'

  type Status = 'todo' | 'in_progress' | 'done'
  type Priority = 'low' | 'medium' | 'high'

  interface Task {
    id: string
    title: string
    description: string
    status: Status
    priority: Priority
    due_date: string
    created_at: string
    updated_at: string
  }

  let tasks = $state<Task[]>([])
  let loading = $state(true)
  let error = $state<string | null>(null)
  let expandedId = $state<string | null>(null)
  let showNewForm = $state(false)

  let newTitle = $state('')
  let newDesc = $state('')
  let newPriority = $state<Priority>('medium')
  let newDueDate = $state('')
  let creating = $state(false)

  const todoCount = $derived(tasks.filter(t => t.status === 'todo').length)
  const inProgressCount = $derived(tasks.filter(t => t.status === 'in_progress').length)
  const doneCount = $derived(tasks.filter(t => t.status === 'done').length)

  async function fetchTasks() {
    loading = true
    error = null
    try {
      const resp = await fetch('/api/widgets/tasks/items')
      if (!resp.ok) throw new Error(`HTTP ${resp.status}`)
      const data = await resp.json()
      tasks = data.tasks ?? []
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load tasks'
    } finally {
      loading = false
    }
  }

  async function createTask() {
    if (!newTitle.trim() || creating) return
    creating = true
    try {
      const resp = await fetch('/api/widgets/tasks/items', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          title: newTitle.trim(),
          description: newDesc.trim(),
          priority: newPriority,
          due_date: newDueDate,
        }),
      })
      if (!resp.ok) throw new Error(`HTTP ${resp.status}`)
      const task: Task = await resp.json()
      tasks = [task, ...tasks]
      newTitle = ''
      newDesc = ''
      newPriority = 'medium'
      newDueDate = ''
      showNewForm = false
    } finally {
      creating = false
    }
  }

  async function updateTask(id: string, updates: Record<string, string>) {
    try {
      const resp = await fetch(`/api/widgets/tasks/items/${id}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(updates),
      })
      if (!resp.ok) return
      const updated: Task = await resp.json()
      tasks = tasks.map(t => t.id === id ? updated : t)
      // Re-sort: active first, then done
      tasks = [
        ...tasks.filter(t => t.status !== 'done'),
        ...tasks.filter(t => t.status === 'done'),
      ]
    } catch { /* silent */ }
  }

  async function deleteTask(id: string) {
    try {
      await fetch(`/api/widgets/tasks/items/${id}`, { method: 'DELETE' })
      tasks = tasks.filter(t => t.id !== id)
      if (expandedId === id) expandedId = null
    } catch { /* silent */ }
  }

  function cycleStatus(task: Task) {
    const next: Record<Status, Status> = {
      todo: 'in_progress',
      in_progress: 'done',
      done: 'todo',
    }
    updateTask(task.id, { status: next[task.status] })
  }

  function isOverdue(dueDate: string): boolean {
    if (!dueDate) return false
    return new Date(dueDate) < new Date(new Date().toDateString())
  }

  function formatDue(dueDate: string): string {
    if (!dueDate) return ''
    const d = new Date(dueDate + 'T00:00:00')
    return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
  }

  $effect(() => { fetchTasks() })
</script>

{#if loading}
  <div class="space-y-3 animate-pulse">
    <div class="flex gap-2 mb-4">
      <div class="h-8 w-20 bg-white/5 rounded-lg"></div>
      <div class="h-8 w-24 bg-white/5 rounded-lg"></div>
      <div class="h-8 w-16 bg-white/5 rounded-lg"></div>
    </div>
    {#each [1,2,3] as _}
      <div class="h-10 bg-white/5 rounded-lg"></div>
    {/each}
  </div>

{:else if error}
  <div class="flex flex-col items-center gap-3 py-6 text-center">
    <p class="text-sm text-red-400">{error}</p>
    <button onclick={fetchTasks} class="btn-ghost text-xs">Retry</button>
  </div>

{:else}
  <div class="flex flex-col gap-3 h-full">

    <!-- Summary + Add button -->
    <div class="flex items-center justify-between gap-2">
      <div class="flex items-center gap-2 flex-wrap">
        <span class="tag bg-white/8 text-gray-400">
          <Circle size={10} class="inline mr-1" />{todoCount} To Do
        </span>
        <span class="tag bg-blue-500/15 text-blue-300">
          <Clock size={10} class="inline mr-1" />{inProgressCount} In Progress
        </span>
        <span class="tag bg-green-500/15 text-green-400">
          <CheckCircle2 size={10} class="inline mr-1" />{doneCount} Done
        </span>
      </div>
      <button
        onclick={() => { showNewForm = !showNewForm }}
        class="btn-ghost p-1.5 shrink-0"
        title="Add task"
      >
        {#if showNewForm}
          <X size={14} />
        {:else}
          <Plus size={14} />
        {/if}
      </button>
    </div>

    <!-- New task form -->
    {#if showNewForm}
      <div class="rounded-lg bg-white/5 border border-white/10 p-3 space-y-2">
        <input
          type="text"
          placeholder="Task title…"
          bind:value={newTitle}
          onkeydown={(e) => e.key === 'Enter' && createTask()}
          class="w-full bg-transparent text-sm text-gray-100 placeholder-gray-600 outline-none border-b border-white/10 pb-1 focus:border-accent/50 transition-colors"
        />
        <textarea
          placeholder="Description (optional)"
          bind:value={newDesc}
          rows="2"
          class="w-full bg-transparent text-xs text-gray-400 placeholder-gray-600 outline-none resize-none"
        ></textarea>
        <div class="flex items-center gap-2 pt-1">
          <select
            bind:value={newPriority}
            class="bg-white/5 text-xs text-gray-400 rounded px-2 py-1 outline-none border border-white/10"
          >
            <option value="low">Low</option>
            <option value="medium">Medium</option>
            <option value="high">High</option>
          </select>
          <input
            type="date"
            bind:value={newDueDate}
            class="bg-white/5 text-xs text-gray-400 rounded px-2 py-1 outline-none border border-white/10"
          />
          <button
            onclick={createTask}
            disabled={!newTitle.trim() || creating}
            class="ml-auto btn-ghost text-xs text-accent disabled:opacity-40 disabled:cursor-not-allowed"
          >
            {creating ? 'Adding…' : 'Add Task'}
          </button>
        </div>
      </div>
    {/if}

    <!-- Task list -->
    <div class="flex-1 overflow-y-auto space-y-1 min-h-0">
      {#if tasks.length === 0}
        <div class="flex flex-col items-center justify-center py-10 text-center gap-2">
          <p class="text-sm text-gray-600">No tasks yet</p>
          <button onclick={() => showNewForm = true} class="btn-ghost text-xs text-accent">
            <Plus size={12} class="inline mr-1" />Add your first task
          </button>
        </div>
      {:else}
        {#each tasks as task (task.id)}
          {@const overdue = isOverdue(task.due_date) && task.status !== 'done'}
          {@const expanded = expandedId === task.id}

          <div class="rounded-lg border transition-colors
            {task.status === 'done'
              ? 'border-white/5 bg-white/3 opacity-60'
              : 'border-white/8 bg-white/5 hover:bg-white/8'}">

            <!-- Task row -->
            <div class="flex items-center gap-2 px-3 py-2.5">
              <!-- Status toggle -->
              <button
                onclick={() => cycleStatus(task)}
                class="shrink-0 transition-colors"
                title="Cycle status"
              >
                {#if task.status === 'todo'}
                  <Circle size={16} class="text-gray-500 hover:text-gray-300" />
                {:else if task.status === 'in_progress'}
                  <Clock size={16} class="text-blue-400" />
                {:else}
                  <CheckCircle2 size={16} class="text-green-500" />
                {/if}
              </button>

              <!-- Priority dot -->
              <span class="shrink-0 w-1.5 h-1.5 rounded-full
                {task.priority === 'high' ? 'bg-red-400' :
                 task.priority === 'medium' ? 'bg-amber-400' : 'bg-gray-600'}">
              </span>

              <!-- Title -->
              <span class="flex-1 text-sm truncate
                {task.status === 'done' ? 'line-through text-gray-500' : 'text-gray-200'}">
                {task.title}
              </span>

              <!-- Due date -->
              {#if task.due_date}
                <span class="shrink-0 flex items-center gap-1 text-xs
                  {overdue ? 'text-red-400 font-medium' : 'text-gray-600'}">
                  <Calendar size={10} />
                  {formatDue(task.due_date)}
                </span>
              {/if}

              <!-- Expand toggle -->
              <button
                onclick={() => expandedId = expanded ? null : task.id}
                class="shrink-0 btn-ghost p-0.5"
              >
                {#if expanded}
                  <ChevronUp size={13} class="text-gray-500" />
                {:else}
                  <ChevronDown size={13} class="text-gray-600" />
                {/if}
              </button>
            </div>

            <!-- Expanded detail -->
            {#if expanded}
              <div class="px-3 pb-3 pt-0 border-t border-white/5 space-y-3">
                {#if task.description}
                  <p class="text-xs text-gray-400 leading-relaxed pt-2">{task.description}</p>
                {:else}
                  <p class="text-xs text-gray-600 italic pt-2">No description</p>
                {/if}

                <div class="flex items-center gap-2 flex-wrap">
                  <!-- Status selector -->
                  <select
                    value={task.status}
                    onchange={(e) => updateTask(task.id, { status: (e.target as HTMLSelectElement).value })}
                    class="bg-white/5 text-xs rounded px-2 py-1 outline-none border border-white/10
                      {task.status === 'todo' ? 'text-gray-400' :
                       task.status === 'in_progress' ? 'text-blue-300' : 'text-green-400'}"
                  >
                    <option value="todo">To Do</option>
                    <option value="in_progress">In Progress</option>
                    <option value="done">Done</option>
                  </select>

                  <!-- Priority selector -->
                  <select
                    value={task.priority}
                    onchange={(e) => updateTask(task.id, { priority: (e.target as HTMLSelectElement).value })}
                    class="bg-white/5 text-xs rounded px-2 py-1 outline-none border border-white/10
                      {task.priority === 'high' ? 'text-red-400' :
                       task.priority === 'medium' ? 'text-amber-400' : 'text-gray-500'}"
                  >
                    <option value="low">Low</option>
                    <option value="medium">Medium</option>
                    <option value="high">High</option>
                  </select>

                  <!-- Due date -->
                  <input
                    type="date"
                    value={task.due_date}
                    onchange={(e) => updateTask(task.id, { due_date: (e.target as HTMLInputElement).value })}
                    class="bg-white/5 text-xs text-gray-400 rounded px-2 py-1 outline-none border border-white/10"
                  />

                  <!-- Delete -->
                  <button
                    onclick={() => deleteTask(task.id)}
                    class="ml-auto btn-ghost p-1 text-gray-600 hover:text-red-400 transition-colors"
                    title="Delete task"
                  >
                    <Trash2 size={13} />
                  </button>
                </div>
              </div>
            {/if}
          </div>
        {/each}
      {/if}
    </div>
  </div>
{/if}
