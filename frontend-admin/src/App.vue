<template>
  <div class="page">
    <header class="header">
      <div>
        <h1>
          提交记录
          <span class="badge">{{ items.length }}</span>
        </h1>
        <p class="subtitle">查看所有已提交的姓名、手机号与时间（演示环境无登录保护）。</p>
      </div>
      <button type="button" :disabled="loading" @click="load">
        {{ loading ? '刷新中…' : '刷新' }}
      </button>
    </header>

    <section class="card">
      <p v-if="loading && !items.length" class="loading">正在加载…</p>
      <p v-else-if="error" class="error">{{ error }}</p>
      <p v-else-if="!items.length" class="empty">暂无提交记录</p>
      <table v-else>
        <thead>
          <tr>
            <th>姓名</th>
            <th>手机号</th>
            <th>提交时间</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in items" :key="item.id">
            <td>{{ item.name }}</td>
            <td class="phone">{{ item.phone }}</td>
            <td>{{ formatTime(item.submittedAt) }}</td>
          </tr>
        </tbody>
      </table>
    </section>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'

const API_BASE = String(import.meta.env.VITE_API_BASE || '').replace(/\/+$/, '')

const items = ref([])
const loading = ref(false)
const error = ref('')

function formatTime(value) {
  if (!value) return '-'
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return value
  return d.toLocaleString('zh-CN', { hour12: false })
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await fetch(`${API_BASE}/api/submissions`)
    const data = await res.json().catch(() => ({}))
    if (!res.ok) {
      throw new Error(data.error || '加载失败')
    }
    items.value = Array.isArray(data) ? data : []
  } catch (err) {
    error.value = err.message || '网络异常，请确认后端已启动'
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>
