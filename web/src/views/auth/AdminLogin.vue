<template>
  <main class="login-page">
    <section class="login-shell">
      <div class="login-visual">
        <div class="login-brand">
          <span class="brand-mark">W</span>
          <div>
            <strong>wjfcm-go</strong>
            <p>内容管理控制台</p>
          </div>
        </div>
        <div class="login-copy">
          <span>Gin + Vue CMS</span>
          <h1>把内容、用户和站点配置收进一个清爽后台。</h1>
          <p>文章发布、评论审核、菜单权限、微信配置和友链申请统一管理，适合日常维护和迁移验收。</p>
        </div>
        <div class="login-feature-grid">
          <div><strong>SEO</strong><span>服务端渲染前台</span></div>
          <div><strong>RBAC</strong><span>角色权限控制</span></div>
          <div><strong>Trace</strong><span>请求链路日志</span></div>
        </div>
      </div>
      <form class="login-panel" @submit.prevent="submit">
        <div class="login-panel-head">
          <span>Admin</span>
          <h2>后台登录</h2>
          <p>请输入管理员账号继续。</p>
        </div>
        <label>
          账号
          <input v-model.trim="form.account" type="text" autocomplete="username" placeholder="请输入账号" />
        </label>
        <label>
          密码
          <input v-model="form.password" type="password" autocomplete="current-password" placeholder="请输入密码" />
        </label>
        <p v-if="error" class="form-error">{{ error }}</p>
        <button class="login-submit" type="submit" :disabled="loading">
          {{ loading ? '登录中...' : '登录控制台' }}
        </button>
        <p class="muted-text">建议生产环境开启 HTTPS，并定期更换管理员密码。</p>
      </form>
    </section>
  </main>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const loading = ref(false)
const error = ref('')

const form = reactive({
  account: '',
  password: ''
})

async function submit() {
  error.value = ''
  if (!form.account || !form.password) {
    error.value = '请输入账号和密码'
    return
  }

  loading.value = true
  try {
    await auth.login(form.account, form.password)
    router.push(route.query.redirect || '/admin')
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}
</script>
