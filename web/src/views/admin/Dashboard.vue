<template>
  <div class="dashboard-page">
    <section class="dashboard-hero">
      <div>
        <span>{{ greeting }}</span>
        <h1>{{ auth.admin?.name || auth.admin?.account || '管理员' }}，欢迎回来</h1>
        <p>这里汇总了内容数据、待处理事项和常用入口，方便快速判断站点状态。</p>
      </div>
      <div class="dashboard-hero-actions">
        <RouterLink class="button-link" to="/admin/articles/create">发布文章</RouterLink>
        <a class="button-link muted" href="/" target="_blank" rel="noreferrer">查看前台</a>
      </div>
    </section>

    <section class="dashboard-grid">
      <article v-for="metric in metrics" :key="metric.key" class="metric-card rich">
        <div class="metric-icon">{{ metric.icon }}</div>
        <div>
          <span>{{ metric.label }}</span>
          <strong>{{ loading ? '...' : metric.value }}</strong>
          <small>{{ metric.note }}</small>
        </div>
      </article>
    </section>

    <section class="dashboard-columns">
      <article class="dashboard-panel">
        <header>
          <div>
            <span>Quick Actions</span>
            <h2>常用操作</h2>
          </div>
        </header>
        <div class="quick-action-grid">
          <RouterLink v-for="action in quickActions" :key="action.to" :to="action.to">
            <strong>{{ action.name }}</strong>
            <span>{{ action.desc }}</span>
          </RouterLink>
        </div>
      </article>

      <article class="dashboard-panel">
        <header>
          <div>
            <span>Review</span>
            <h2>待处理</h2>
          </div>
        </header>
        <div class="todo-list">
          <RouterLink to="/admin/comments">
            <strong>{{ loading ? '...' : pending.comments }}</strong>
            <span>待审核评论</span>
          </RouterLink>
          <RouterLink to="/admin/friend-links">
            <strong>{{ loading ? '...' : pending.friendLinks }}</strong>
            <span>待审核友链</span>
          </RouterLink>
          <RouterLink to="/admin/articles">
            <strong>{{ loading ? '...' : pending.articles }}</strong>
            <span>草稿/下线文章</span>
          </RouterLink>
        </div>
      </article>
    </section>

    <section class="dashboard-columns">
      <article class="dashboard-panel">
        <header>
          <div>
            <span>Recent</span>
            <h2>最新文章</h2>
          </div>
          <RouterLink class="small-link" to="/admin/articles">全部</RouterLink>
        </header>
        <div class="recent-list">
          <RouterLink v-for="article in recentArticles" :key="article.id" :to="`/admin/articles/${article.id}/edit`">
            <strong>{{ article.title }}</strong>
            <span>{{ article.author || '未设置作者' }} · {{ article.created_at || article.create_time || '-' }}</span>
          </RouterLink>
          <p v-if="!loading && !recentArticles.length" class="empty-note">暂无文章数据</p>
        </div>
      </article>

      <article class="dashboard-panel">
        <header>
          <div>
            <span>System</span>
            <h2>运行状态</h2>
          </div>
        </header>
        <dl class="system-list">
          <div><dt>接口地址</dt><dd>{{ apiBase }}</dd></div>
          <div><dt>登录账号</dt><dd>{{ auth.admin?.account || auth.admin?.name || '-' }}</dd></div>
          <div><dt>权限数量</dt><dd>{{ auth.permissions?.length || 0 }}</dd></div>
          <div><dt>前端版本</dt><dd>Vue 3 + Vite</dd></div>
        </dl>
        <p v-if="error" class="form-error">{{ error }}</p>
      </article>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { listResource } from '../../api/adminResources'
import { useAuthStore } from '../../stores/auth'

const auth = useAuthStore()
const loading = ref(true)
const error = ref('')
const counts = reactive({
  articles: 0,
  comments: 0,
  users: 0,
  categories: 0,
  tags: 0,
  friendLinks: 0
})
const pending = reactive({
  comments: 0,
  friendLinks: 0,
  articles: 0
})
const recentArticles = ref([])
const apiBase = import.meta.env.VITE_API_BASE_URL || '/api'

const greeting = computed(() => {
  const hour = new Date().getHours()
  if (hour < 6) return '夜深了'
  if (hour < 12) return '早上好'
  if (hour < 18) return '下午好'
  return '晚上好'
})

const metrics = computed(() => [
  { key: 'articles', label: '文章总数', value: counts.articles, note: '含正常和回收站文章', icon: '文' },
  { key: 'comments', label: '评论总数', value: counts.comments, note: `${pending.comments} 条待审核`, icon: '评' },
  { key: 'users', label: '用户总数', value: counts.users, note: '前台注册用户', icon: '用' },
  { key: 'taxonomy', label: '分类/标签', value: `${counts.categories}/${counts.tags}`, note: '内容组织结构', icon: '类' }
])

const quickActions = [
  { name: '发布文章', desc: '创建 Markdown 内容', to: '/admin/articles/create' },
  { name: '评论审核', desc: '处理留言和回复', to: '/admin/comments' },
  { name: '网站设置', desc: 'LOGO、SEO、微信配置', to: '/admin/system-configs' },
  { name: '权限管理', desc: '菜单和接口权限', to: '/admin/permissions' }
]

onMounted(loadDashboard)

async function loadDashboard() {
  loading.value = true
  error.value = ''
  try {
    const [articles, comments, users, categories, tags, friendLinks, pendingComments, pendingFriends, hiddenArticles] =
      await Promise.allSettled([
        listResource('articles', { page: 1, limit: 5 }),
        listResource('comments', { page: 1, limit: 1 }),
        listResource('users', { page: 1, limit: 1 }),
        listResource('categories', { page: 1, limit: 1 }),
        listResource('tags', { page: 1, limit: 1 }),
        listResource('friend-links', { page: 1, limit: 1 }),
        listResource('comments', { page: 1, limit: 1, status: 0 }),
        listResource('friend-links', { page: 1, limit: 1, status: 0 }),
        listResource('articles', { page: 1, limit: 1, status: 0 })
      ])

    counts.articles = countOf(articles)
    counts.comments = countOf(comments)
    counts.users = countOf(users)
    counts.categories = countOf(categories)
    counts.tags = countOf(tags)
    counts.friendLinks = countOf(friendLinks)
    pending.comments = countOf(pendingComments)
    pending.friendLinks = countOf(pendingFriends)
    pending.articles = countOf(hiddenArticles)
    recentArticles.value = dataOf(articles).slice(0, 5)
  } catch (err) {
    error.value = err.message || '控制台数据加载失败'
  } finally {
    loading.value = false
  }
}

function countOf(result) {
  if (result.status !== 'fulfilled') return 0
  return Number(result.value?.count || 0)
}

function dataOf(result) {
  if (result.status !== 'fulfilled') return []
  return Array.isArray(result.value?.data) ? result.value.data : []
}
</script>
