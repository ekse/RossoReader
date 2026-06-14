import { createRouter, createWebHistory } from 'vue-router'
import FeedItems from '@/views/FeedItems.vue'
import StarredItems from '@/views/StarredItems.vue'
import Settings from '@/views/Settings.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/starred' },
    { path: '/feed/:id', name: 'feed', component: FeedItems },
    { path: '/starred', name: 'starred', component: StarredItems },
    { path: '/settings', name: 'settings', component: Settings },
  ],
})

export default router
