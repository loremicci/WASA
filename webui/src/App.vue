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

async function onLogout() {
	try {
		await axios.delete('/session')
	} catch (e) {
		console.warn("Server non ha risposto al logout", e)
	}
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
	if (updatedUser.photoUrl !== undefined) {
		currentUser.value.photoUrl = updatedUser.photoUrl
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
	background: rgba(255, 255, 255, 0.85) !important;
	backdrop-filter: blur(20px);
	-webkit-backdrop-filter: blur(20px);
	border: 1px solid rgba(255, 255, 255, 0.5) !important;
	border-radius: 20px;
	box-shadow: 0 8px 32px rgba(0, 0, 0, 0.05), inset 0 0 0 1px rgba(255, 255, 255, 0.4) !important;
	transition: all 0.3s ease;
}
</style>
