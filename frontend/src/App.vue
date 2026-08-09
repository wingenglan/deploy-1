<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

type Status = 'todo' | 'in_progress' | 'done'

interface Task {
  id: number
  title: string
  owner: string
  status: Status
  dueDate: string
  createdAt: string
}

interface TaskDraft {
  title: string
  owner: string
  status: Status
  dueDate: string
}

const tasks = ref<Task[]>([])
const isLoading = ref(true)
const isSaving = ref(false)
const message = ref('')
const filter = ref<'all' | Status>('all')
const formOpen = ref(false)
const editingId = ref<number | null>(null)
const draft = ref<TaskDraft>(emptyDraft())

const visibleTasks = computed(() => filter.value === 'all' ? tasks.value : tasks.value.filter((task) => task.status === filter.value))
const activeCount = computed(() => tasks.value.filter((task) => task.status !== 'done').length)

// emptyDraft returns the explicit defaults used to open a new task form.
function emptyDraft(): TaskDraft {
  return { title: '', owner: '', status: 'todo', dueDate: new Date().toISOString().slice(0, 10) }
}

// loadTasks fetches the persisted board and surfaces a clear retry message on failure.
async function loadTasks() {
  isLoading.value = true
  try {
    const response = await fetch('/api/tasks')
    if (!response.ok) throw new Error('读取任务失败')
    tasks.value = await response.json()
  } catch {
    message.value = '无法连接服务。请确认后端正在运行后重试。'
  } finally {
    isLoading.value = false
  }
}

// openCreate opens a fresh form without carrying values from a prior edit.
function openCreate() {
  editingId.value = null
  draft.value = emptyDraft()
  formOpen.value = true
}

// openEdit copies a selected task into the form for deliberate editing.
function openEdit(task: Task) {
  editingId.value = task.id
  draft.value = { title: task.title, owner: task.owner, status: task.status, dueDate: task.dueDate }
  formOpen.value = true
}

// saveTask creates or updates a task, then reloads the board from the database.
async function saveTask() {
  isSaving.value = true
  message.value = ''
  const isEditing = editingId.value !== null
  try {
    const response = await fetch(isEditing ? `/api/tasks/${editingId.value}` : '/api/tasks', {
      method: isEditing ? 'PUT' : 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(draft.value),
    })
    if (!response.ok) {
      const body = await response.json()
      throw new Error(body.error)
    }
    formOpen.value = false
    message.value = isEditing ? '任务已更新。' : '任务已创建。'
    await loadTasks()
  } catch (error) {
    message.value = error instanceof Error ? error.message : '保存任务失败'
  } finally {
    isSaving.value = false
  }
}

// deleteTask removes a task after an explicit user confirmation.
async function deleteTask(task: Task) {
  if (!window.confirm(`确定删除“${task.title}”吗？`)) return
  const response = await fetch(`/api/tasks/${task.id}`, { method: 'DELETE' })
  if (response.ok) {
    message.value = '任务已删除。'
    await loadTasks()
  } else {
    message.value = '删除任务失败。'
  }
}

// statusLabel translates persisted status values into user-facing labels.
function statusLabel(status: Status) {
  return { todo: '待处理', in_progress: '进行中', done: '已完成' }[status]
}

onMounted(loadTasks)
</script>

<template>
  <main class="shell">
    <header class="masthead">
      <div>
        <p class="eyebrow">OPERATIONS / FLOWBOARD</p>
        <h1>让每一个任务（jenkins测试）<br /><em>落到实处。</em></h1>
      </div>
      <div class="pulse"><span></span>{{ activeCount }} 项正在推进</div>
    </header>

    <section class="toolbar" aria-label="任务操作">
      <div class="filters">
        <button v-for="item in ['all', 'todo', 'in_progress', 'done'] as const" :key="item" :class="{ selected: filter === item }" @click="filter = item">
          {{ item === 'all' ? '全部任务' : statusLabel(item) }}
        </button>
      </div>
      <button class="create" @click="openCreate">+ 新建任务</button>
    </section>

    <p v-if="message" class="message" role="status">{{ message }}</p>

    <section class="board" aria-live="polite">
      <div v-if="isLoading" class="empty">正在读取任务板…</div>
      <div v-else-if="visibleTasks.length === 0" class="empty">当前筛选条件下没有任务。新建一项开始推进。</div>
      <article v-for="task in visibleTasks" :key="task.id" class="task-card" :class="task.status">
        <div class="status-mark" :title="statusLabel(task.status)"></div>
        <div class="task-copy">
          <p class="task-meta">{{ statusLabel(task.status) }} · {{ task.owner }}</p>
          <h2>{{ task.title }}</h2>
          <p class="due">截止 {{ task.dueDate }}</p>
        </div>
        <div class="actions">
          <button @click="openEdit(task)">编辑</button>
          <button class="danger" @click="deleteTask(task)">删除</button>
        </div>
      </article>
    </section>

    <div v-if="formOpen" class="modal-backdrop" @click.self="formOpen = false">
      <form class="task-form" @submit.prevent="saveTask">
        <div class="form-top"><p class="eyebrow">TASK ENTRY</p><button type="button" class="close" @click="formOpen = false">×</button></div>
        <h2>{{ editingId === null ? '新建一项工作' : '更新任务' }}</h2>
        <label>任务名称<input v-model="draft.title" required maxlength="120" placeholder="例如：完成版本发布检查" /></label>
        <label>负责人<input v-model="draft.owner" required maxlength="40" placeholder="姓名" /></label>
        <div class="form-grid">
          <label>状态<select v-model="draft.status"><option value="todo">待处理</option><option value="in_progress">进行中</option><option value="done">已完成</option></select></label>
          <label>截止日期<input v-model="draft.dueDate" required type="date" /></label>
        </div>
        <button class="create submit" :disabled="isSaving">{{ isSaving ? '正在保存…' : '保存任务' }}</button>
      </form>
    </div>
  </main>
</template>
