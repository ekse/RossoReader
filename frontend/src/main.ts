import { createApp } from "vue";
import { createPinia } from "pinia";
import App from "./App.vue";
import router from "./router";
import "./style.css";
import { useAuthStore } from "./stores/auth";

const saved = localStorage.getItem("theme");
if (saved === "dark" || (!saved && window.matchMedia("(prefers-color-scheme: dark)").matches)) {
  document.documentElement.classList.add("dark");
}

const app = createApp(App);
const pinia = createPinia();
app.use(pinia);

// Restore the session before installing the router so the initial navigation
// guard sees the authenticated user and doesn't bounce to /login.
const auth = useAuthStore();
auth.fetchMe().finally(() => {
  app.use(router);
  app.mount("#app");
});
