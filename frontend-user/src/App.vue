<template>
  <main class="card">
    <h1>用户信息提交</h1>
    <p class="subtitle">请填写姓名和手机号，提交后可在管理后台查看。</p>

    <form @submit.prevent="onSubmit">
      <div class="field">
        <label for="name">姓名</label>
        <input
          id="name"
          v-model.trim="name"
          type="text"
          autocomplete="name"
          placeholder="请输入姓名"
          maxlength="50"
        />
      </div>

      <div class="field">
        <label for="phone">手机号</label>
        <input
          id="phone"
          v-model.trim="phone"
          type="tel"
          inputmode="numeric"
          autocomplete="tel"
          placeholder="请输入 11 位手机号"
          maxlength="11"
        />
      </div>

      <button type="submit" :disabled="loading">
        {{ loading ? '提交中…' : '提交' }}
      </button>
    </form>

    <p v-if="message" class="message" :class="status">{{ message }}</p>
  </main>
</template>

<script setup>
import { ref } from 'vue'

const API_BASE = import.meta.env.VITE_API_BASE || ''

const name = ref('')
const phone = ref('')
const loading = ref(false)
const message = ref('')
const status = ref('')

const cnMobile = /^1[3-9]\d{9}$/

async function onSubmit() {
  message.value = ''
  status.value = ''

  if (!name.value) {
    status.value = 'error'
    message.value = '请填写姓名'
    return
  }
  if (!cnMobile.test(phone.value)) {
    status.value = 'error'
    message.value = '请输入有效的中国大陆手机号'
    return
  }

  loading.value = true
  try {
    const res = await fetch(`${API_BASE}/api/submissions`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: name.value, phone: phone.value }),
    })
    const data = await res.json().catch(() => ({}))
    if (!res.ok) {
      throw new Error(data.error || '提交失败')
    }
    status.value = 'ok'
    message.value = `提交成功：${data.name}（${data.phone}）`
    name.value = ''
    phone.value = ''
  } catch (err) {
    status.value = 'error'
    message.value = err.message || '网络异常，请确认后端已启动'
  } finally {
    loading.value = false
  }
}
</script>
