<script setup>
import { ref } from 'vue'
import axios from '@/services/axios'

const emit = defineEmits(['loginSuccess'])

const username = ref('')
const errorMsg = ref('')
const loading = ref(false)

async function login() {
	if (username.value.trim().length < 3) {
		errorMsg.value = "Username must be at least 3 characters long."
		return
	}
	loading.value = true
	errorMsg.value = ''
	try {
		const res = await axios.post('/session', { name: username.value.trim() })
		const token = res.data.identifier
		// Save to local storage
		localStorage.setItem("token", token)
		localStorage.setItem("username", username.value.trim())
		// Emit success
		emit('loginSuccess', { token, username: username.value.trim() })
	} catch (e) {
		errorMsg.value = e.response?.data?.message || e.toString()
	} finally {
		loading.value = false
	}
}
</script>

<template>
	<div class="login-container">
		<div class="glass-card login-card">
			<h1 class="brand-title">WASAText</h1>
			<p class="subtitle">Connect effortlessly</p>
			
			<div class="form-group mt-4">
				<label class="form-label text-muted">Choose a username</label>
				<input v-model="username" type="text" class="form-control glass-input" placeholder="e.g. Maria" @keyup.enter="login" :disabled="loading" />
			</div>
			
			<ErrorMsg v-if="errorMsg" :msg="errorMsg" class="mt-3" />
			
			<button class="btn btn-primary w-100 mt-4 glass-btn" @click="login" :disabled="loading">
				<LoadingSpinner :loading="loading">
					Start Chatting <svg class="feather ms-2"><use href="/feather-sprite-v4.29.0.svg#arrow-right"/></svg>
				</LoadingSpinner>
			</button>
		</div>
	</div>
</template>

<style scoped>
.login-container {
	height: 100vh;
	width: 100vw;
	display: flex;
	align-items: center;
	justify-content: center;
	background: #0f172a;
	position: relative;
	overflow: hidden;
}

.login-container::before {
	content: '';
	position: absolute;
	width: 600px;
	height: 600px;
	background: #4f46e5;
	filter: blur(150px);
	border-radius: 50%;
	top: -200px;
	left: -100px;
	opacity: 0.35;
	animation: orbFloat1 18s ease-in-out infinite;
}

.login-container::after {
	content: '';
	position: absolute;
	width: 500px;
	height: 500px;
	background: #ec4899;
	filter: blur(150px);
	border-radius: 50%;
	bottom: -150px;
	right: -100px;
	opacity: 0.35;
	animation: orbFloat2 15s ease-in-out infinite;
}

@keyframes orbFloat1 {
	0% { transform: translate(0, 0) scale(1); }
	33% { transform: translate(25vw, 15vh) scale(1.2); }
	66% { transform: translate(-10vw, 30vh) scale(0.8); }
	100% { transform: translate(0, 0) scale(1); }
}

@keyframes orbFloat2 {
	0% { transform: translate(0, 0) scale(1); }
	33% { transform: translate(-20vw, -25vh) scale(1.3); }
	66% { transform: translate(15vw, -10vh) scale(0.9); }
	100% { transform: translate(0, 0) scale(1); }
}

.login-card {
	width: 100%;
	max-width: 400px;
	padding: 3rem 2.5rem;
	text-align: center;
	z-index: 10;
	animation: slideUpFade 0.8s cubic-bezier(0.16, 1, 0.3, 1);
}

.brand-title {
	font-weight: 800;
	font-size: 2.5rem;
	background: linear-gradient(135deg, #4f46e5, #ec4899);
	-webkit-background-clip: text;
	-webkit-text-fill-color: transparent;
	margin-bottom: 0.25rem;
}

.subtitle {
	color: #475569;
	font-size: 0.95rem;
	margin-bottom: 2rem;
}

.glass-input {
	background: rgba(0, 0, 0, 0.03) !important;
	border: 1px solid rgba(0, 0, 0, 0.12) !important;
	color: #0f172a !important;
	padding: 0.75rem 1rem;
	border-radius: 12px;
	transition: all 0.3s ease;
}

.glass-input:focus {
	background: rgba(0, 0, 0, 0.05) !important;
	border-color: #4f46e5 !important;
	box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.15) !important;
}

.glass-input::placeholder {
	color: rgba(0, 0, 0, 0.4);
}

.glass-btn {
	background: linear-gradient(135deg, #4f46e5, #ec4899);
	border: none;
	border-radius: 12px;
	padding: 0.75rem 1rem;
	font-weight: 600;
	letter-spacing: 0.5px;
	transition: transform 0.2s, box-shadow 0.2s;
	display: flex;
	align-items: center;
	justify-content: center;
	color: white !important;
}

.glass-btn:hover {
	transform: translateY(-2px);
	box-shadow: 0 10px 25px rgba(236, 72, 153, 0.3);
}

.glass-btn:active {
	transform: translateY(0);
}

@keyframes slideUp {
	from {
		opacity: 0;
		transform: translateY(30px);
	}
	to {
		opacity: 1;
		transform: translateY(0);
	}
}
</style>
