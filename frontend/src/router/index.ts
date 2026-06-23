import { createRouter, createWebHistory } from 'vue-router'
import UnreadItems from '@/views/UnreadItems.vue'
import FeedItems from '@/views/FeedItems.vue'
import StarredItems from '@/views/StarredItems.vue'
import Settings from '@/views/Settings.vue'
import Login from '@/views/Login.vue'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/unread' },
    { path: '/login', name: 'login', component: Login, meta: { public: true } },
    { path: '/unread', name: 'unread', component: UnreadItems },
    { path: '/feed/:id', name: 'feed', component: FeedItems },
    { path: '/starred', name: 'starred', component: StarredItems },
    { path: '/settings', name: 'settings', component: Settings },
  ],
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (!to.meta.public && !auth.isAuthenticated) {
    return { name: 'login' }
  }
  if (to.name === 'login' && auth.isAuthenticated) {
    return { name: 'unread' }
  }
})

export default router