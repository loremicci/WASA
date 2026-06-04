<script setup>
import { ref, onMounted } from 'vue'
import LoginView from './views/LoginView.vue'
import HomeView from './views/HomeView.vue'
import axios from '@/services/axios'

const isAuthenticated = ref(false)
const currentUser = ref(null)

onMounted(() => {
	const token = localStorage.getItem("token")
	const username = localStorage.getItem("username")
	if (token && username) {
		currentUser.value = { identifier: token, username }
		isAuthenticated.value = true
	}
})

function onLoginSuccess(user) {
	currentUser.value = { identifier: user.token, username: user.username }
	isAuthenticated.value = true
}

function onLogout() {
	localStorage.removeItem("token")
	localStorage.removeItem("username")
	currentUser.value = null
	isAuthenticated.value = false
}

function onUpdateUser(updatedUser) {
	if (updatedUser.name) {
		currentUser.value.username = updatedUser.name
		localStorage.setItem("username", updatedUser.name)
	}
}
</script>

<template>
	<LoginView v-if="!isAuthenticated" @loginSuccess="onLoginSuccess" />
	<HomeView v-else :currentUser="currentUser" @logout="onLogout" @updateUser="onUpdateUser" />
</template>

<style>
/* Global glassmorphism classes */
.glass-card {
	background: rgba(255, 255, 255, 0.8) !important;
	backdrop-filter: blur(16px);
	-webkit-backdrop-filter: blur(16px);
	border: 1px solid rgba(0, 0, 0, 0.08) !important;
	border-radius: 16px;
	box-shadow: 0 4px 30px rgba(0, 0, 0, 0.05) !important;
}
</style>
