<template>
  <div v-if="auth.isAuthenticated" class="layout">
    <aside class="sidebar">
      <router-link to="/dashboard">Dashboard</router-link>
      <router-link to="/gateway">Gateway</router-link>
      <router-link to="/agents">Agents</router-link>
      <router-link to="/bindings">Bindings</router-link>
      <router-link to="/skills">Skills</router-link>
      <router-link to="/config">Config</router-link>
      <router-link to="/backups">Backups</router-link>
      <router-link to="/tasks">Tasks</router-link>
      <router-link to="/admin/users">Users</router-link>
    </aside>
    <main class="content"><router-view /></main>
  </div>
  <router-view v-else />
</template>
<script setup lang="ts">
import { useAuthStore } from './stores/auth'

const auth = useAuthStore()
</script>
<style scoped>
.layout{display:flex;min-height:100vh}.sidebar{width:220px;display:flex;flex-direction:column;gap:8px;padding:12px;background:#f5f7fa}.content{flex:1;padding:16px}
</style>
