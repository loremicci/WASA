<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import axios from '@/services/axios'

const props = defineProps({
	currentUser: Object
})

const emit = defineEmits(['logout', 'updateUser'])

// State
const conversations = ref([])
const activeConversation = ref(null)
const messages = ref([])

// UI State
const searchQuery = ref('')
const searchResults = ref([])
const isSearching = ref(false)

const messageText = ref('')
const messagePhoto = ref(null)
const replyTo = ref(null)

const loadingConvs = ref(false)
const loadingMessages = ref(false)
const errorMsg = ref('')

// Modals
const showProfileModal = ref(false)
const showGroupModal = ref(false)
const showNewChatModal = ref(false)
const showGroupInfoModal = ref(false)
const showForwardModal = ref(false)

const convSearchQuery = ref('')
const filteredConversations = computed(() => {
	if (!convSearchQuery.value) return conversations.value
	const q = convSearchQuery.value.toLowerCase()
	return conversations.value.filter(c => c.name.toLowerCase().includes(q))
})

// Forms
const profileForm = ref({ name: props.currentUser.username, photo: null })
const groupForm = ref({ name: '', search: '', results: [], selectedUsers: [] })
const groupInfoForm = ref({ name: '', photo: null, search: '', results: [] })
const forwardForm = ref({ search: '', results: [] })
const forwardMsgId = ref(null)

// Polling interval
let pollInterval = null

const isDarkMode = ref(localStorage.getItem('theme') === 'dark')

function toggleTheme() {
	isDarkMode.value = !isDarkMode.value
	if (isDarkMode.value) {
		document.body.classList.add('dark-theme')
		localStorage.setItem('theme', 'dark')
	} else {
		document.body.classList.remove('dark-theme')
		localStorage.setItem('theme', 'light')
	}
}

onMounted(() => {
	if (isDarkMode.value) {
		document.body.classList.add('dark-theme')
	}
	fetchConversations()
	// Poll every 3 seconds for updates
	pollInterval = setInterval(() => {
		fetchConversations()
		if (activeConversation.value) {
			fetchMessages(activeConversation.value.identifier, true)
		}
	}, 3000)
})

onUnmounted(() => {
	if (pollInterval) clearInterval(pollInterval)
})

const myUserId = computed(() => props.currentUser.identifier)

// API Calls
async function fetchConversations() {
	try {
		const res = await axios.get('/conversations')
		conversations.value = res.data || []
		if (activeConversation.value) {
			const updated = conversations.value.find(c => c.identifier === activeConversation.value.identifier)
			if (updated) {
				activeConversation.value.name = updated.name
				activeConversation.value.photoUrl = updated.photoUrl
				activeConversation.value.isOnline = updated.isOnline
			}
		}
	} catch (e) {
		console.error("Failed to fetch conversations", e)
	}
}

async function fetchUserSearch(query) {
	try {
		const url = query && query.length > 0 ? `/users?username=${encodeURIComponent(query)}` : `/users`
		const res = await axios.get(url)
		return (res.data || []).filter(u => u.identifier !== myUserId.value)
	} catch (e) {
		return []
	}
}

async function doSearch() {
	searchResults.value = await fetchUserSearch(searchQuery.value)
}

async function doGroupSearch() {
	groupForm.value.results = await fetchUserSearch(groupForm.value.search)
}

async function doGroupInfoSearch() {
	groupInfoForm.value.results = await fetchUserSearch(groupInfoForm.value.search)
}

async function doForwardSearch() {
	forwardForm.value.results = await fetchUserSearch(forwardForm.value.search)
}

async function doForwardToUser(user) {
	try {
		const res = await axios.post('/conversations', { userId: user.identifier })
		const convId = res.data.identifier
		await doForward(convId)
		forwardForm.value.search = ''
		forwardForm.value.results = []
	} catch (e) {
		alert("Failed to forward to user")
	}
}

async function startDirectChat(user) {
	try {
		const res = await axios.post('/conversations', { userId: user.identifier })
		searchQuery.value = ''
		searchResults.value = []
		await fetchConversations()
		openConversation(res.data)
	} catch (e) {
		alert(e.response?.data?.message || "Failed to start chat")
	}
}

function openConversation(conv) {
	activeConversation.value = conv
	replyTo.value = null
	messageText.value = ''
	messagePhoto.value = null
	fetchMessages(conv.identifier)
}

async function fetchMessages(convId, silent = false) {
	if (!silent) loadingMessages.value = true
	try {
		const res = await axios.get(`/conversations/${convId}`)
		messages.value = res.data || []
	} catch (e) {
		if (!silent) errorMsg.value = "Failed to load messages"
	} finally {
		loadingMessages.value = false
	}
}

async function sendMessage() {
	if (!messageText.value.trim() && !messagePhoto.value) return
	if (!activeConversation.value) return

	const formData = new FormData()
	if (messageText.value.trim()) formData.append("text", messageText.value.trim())
	if (messagePhoto.value) formData.append("photo", messagePhoto.value)
	if (replyTo.value) formData.append("replyTo", replyTo.value.identifier)

	try {
		await axios.post(`/conversations/${activeConversation.value.identifier}/messages`, formData, {
			headers: { "Content-Type": "multipart/form-data" }
		})
		messageText.value = ''
		messagePhoto.value = null
		replyTo.value = null
		// Refresh
		fetchConversations()
		fetchMessages(activeConversation.value.identifier)
	} catch (e) {
		alert("Failed to send message")
	}
}

function handlePhotoSelect(e) {
	if (e.target.files.length > 0) {
		messagePhoto.value = e.target.files[0]
	} else {
		messagePhoto.value = null
	}
}

function clearPhoto() {
	messagePhoto.value = null
	const input = document.getElementById("photo-upload")
	if (input) input.value = ''
}

async function deleteMsg(id) {
	if (!confirm("Delete message?")) return
	try {
		await axios.delete(`/messages/${id}`)
		fetchMessages(activeConversation.value.identifier)
		fetchConversations()
	} catch (e) {
		alert("Failed to delete")
	}
}

async function openForward(msgId) {
	forwardMsgId.value = msgId
	showForwardModal.value = true
	forwardForm.value.search = ''
	forwardForm.value.results = await fetchUserSearch('')
}

async function doForward(convId) {
	try {
		await axios.post(`/messages/${forwardMsgId.value}/forward`, { conversationId: convId })
		showForwardModal.value = false
		forwardMsgId.value = null
		fetchConversations()
		if (activeConversation.value?.identifier === convId) {
			fetchMessages(convId)
		} else {
			alert("Message forwarded!")
		}
	} catch (e) {
		alert("Failed to forward")
	}
}

async function toggleReaction(msgId, emoticon, hasReacted) {
	try {
		if (hasReacted) {
			await axios.delete(`/messages/${msgId}/comments/${encodeURIComponent(emoticon)}`)
		} else {
			await axios.put(`/messages/${msgId}/comments/${encodeURIComponent(emoticon)}`)
		}
		fetchMessages(activeConversation.value.identifier, true)
	} catch (e) {
		console.error("Reaction failed")
	}
}

const EMOJIS = ['👍', '❤️', '😂', '😮', '😢', '🙏']

function hasUserReacted(msg, emoticon) {
	if (!msg.reactions) return false
	return msg.reactions.some(r => r.emoticon === emoticon && r.user.identifier === myUserId.value)
}

function showReactionAuthors(emoticon, reactions) {
	const authors = reactions.filter(r => r.emoticon === emoticon).map(r => r.user.username).join(', ')
	alert(`${emoticon} reacted by: ${authors}`)
}

// Group Mgmt
async function createGroup() {
	if (!groupForm.value.name.trim()) return
	const memberIds = groupForm.value.selectedUsers.map(u => u.identifier)
	try {
		const res = await axios.post('/groups', {
			name: groupForm.value.name.trim(),
			members: memberIds
		})
		showGroupModal.value = false
		groupForm.value = { name: '', search: '', results: [], selectedUsers: [] }
		await fetchConversations()
		const newConv = conversations.value.find(c => c.identifier === res.data.identifier)
		openConversation(newConv || { ...res.data, isGroup: true })
	} catch (e) {
		alert("Failed to create group")
	}
}

function toggleGroupUserSelect(u) {
	const idx = groupForm.value.selectedUsers.findIndex(x => x.identifier === u.identifier)
	if (idx >= 0) {
		groupForm.value.selectedUsers.splice(idx, 1)
	} else {
		groupForm.value.selectedUsers.push(u)
	}
}

// Group Info Modals
async function saveGroupInfo() {
	if (!activeConversation.value || !activeConversation.value.isGroup) return
	const gid = activeConversation.value.identifier
	try {
		if (groupInfoForm.value.name.trim() && groupInfoForm.value.name.trim() !== activeConversation.value.name) {
			await axios.put(`/groups/${gid}/name`, { name: groupInfoForm.value.name.trim() })
		}
		if (groupInfoForm.value.photo) {
			const fd = new FormData()
			fd.append("photo", groupInfoForm.value.photo)
			await axios.put(`/groups/${gid}/photo`, fd)
		}
		showGroupInfoModal.value = false
		activeConversation.value.name = groupInfoForm.value.name.trim()
		fetchConversations()
	} catch (e) {
		alert("Failed to update group")
	}
}

async function addGroupMember(user) {
	const gid = activeConversation.value.identifier
	try {
		await axios.post(`/groups/${gid}/members`, { userId: user.identifier })
		alert(`${user.username} added to group!`)
		groupInfoForm.value.search = ''
		groupInfoForm.value.results = []
	} catch (e) {
		alert("Failed to add member")
	}
}

async function leaveGroup() {
	if (!confirm("Are you sure you want to leave this group?")) return
	const gid = activeConversation.value.identifier
	try {
		await axios.delete(`/groups/${gid}/members/me`)
		showGroupInfoModal.value = false
		activeConversation.value = null
		fetchConversations()
	} catch (e) {
		alert("Failed to leave group")
	}
}

// Profile
async function saveProfile() {
	try {
		if (profileForm.value.name.trim() && profileForm.value.name.trim() !== props.currentUser.username) {
			await axios.put('/users/me/name', { name: profileForm.value.name.trim() })
		}
		let newPhotoUrl = props.currentUser.photoUrl
		if (profileForm.value.photo) {
			const fd = new FormData()
			fd.append("photo", profileForm.value.photo)
			await axios.put('/users/me/photo', fd)
			newPhotoUrl = URL.createObjectURL(profileForm.value.photo)
		}
		showProfileModal.value = false
		emit('updateUser', { name: profileForm.value.name.trim(), photoUrl: newPhotoUrl })
		alert("Profile updated!")
	} catch (e) {
		if (e.response && e.response.status === 409) {
			alert("Failed to update profile: Username already taken")
		} else {
			alert("Failed to update profile")
		}
	}
}

// Utils
function formatTime(ts) {
	if (!ts) return ''
	const d = new Date(ts)
	return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}
function formatDate(ts) {
	if (!ts) return ''
	const d = new Date(ts)
	return d.toLocaleDateString() + ' ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

function getInitials(name) {
	if (!name) return '?'
	return name.substring(0, 2).toUpperCase()
}

function resolveAvatar(url) {
	if (!url) return null
	if (url.startsWith('data:') || url.startsWith('blob:') || url.startsWith('http')) return url
	return __API_URL__ + url
}

function getRepliedMessage(replyId) {
	return messages.value.find(m => m.identifier === replyId)
}
</script>

<template>
	<div class="app-layout">
		<!-- Icon Sidebar (Far Left) -->
		<div class="icon-sidebar d-flex flex-column align-items-center py-3">
			<div class="top-icons d-flex flex-column gap-3 w-100 align-items-center">
				<button class="icon-btn active" title="Chats" style="margin-top: 10px;">
					<svg class="feather"><use href="/feather-sprite-v4.29.0.svg#message-square"/></svg>
				</button>
			</div>
		</div>

		<!-- Conversations Sidebar -->
		<div class="sidebar">
			<!-- Header -->
			<div class="sidebar-header border-0 pb-0 pt-4 px-4">
				<h2 class="fw-bold fs-3 mb-4" style="background: linear-gradient(135deg, #4f46e5, #ec4899); -webkit-background-clip: text; -webkit-text-fill-color: transparent;">WASAText</h2>
				<button class="btn btn-primary w-100 rounded-4 py-3 fw-bold d-flex justify-content-center align-items-center text-white shadow-sm" style="background: linear-gradient(135deg, #4f46e5, #ec4899); border: none; font-size: 1rem;" @click="showNewChatModal = true; searchQuery = ''; searchResults = [];">
					<svg class="feather me-2"><use href="/feather-sprite-v4.29.0.svg#edit-3"/></svg> New Chat
				</button>
			</div>
			
			<!-- Search -->
			<div class="p-4 pb-2 border-0">
				<div class="position-relative">
					<svg class="feather position-absolute" style="left: 14px; top: 12px; color: var(--text-secondary); width: 16px; height: 16px;"><use href="/feather-sprite-v4.29.0.svg#search"/></svg>
					<input v-model="convSearchQuery" type="text" class="form-control border-0 rounded-4 w-100" style="background: var(--bg-panel-hover); padding-left: 40px; padding-top: 10px; padding-bottom: 10px; color: var(--text-primary);" placeholder="Search conversations..." />
				</div>
			</div>
			
			<!-- Conversations List -->
			<div class="conv-list">
				<div v-if="filteredConversations.length === 0" class="text-center text-muted p-4 mt-4 fs-7">
					No conversations found.
				</div>
				<div v-for="c in filteredConversations" :key="c.identifier" 
					class="conv-item" :class="{ 'active': activeConversation?.identifier === c.identifier }"
					@click="openConversation(c)">
					
					<div class="avatar-container position-relative me-3">
						<img v-if="c.photoUrl" :src="resolveAvatar(c.photoUrl)" class="avatar" />
						<div v-else class="avatar">{{ getInitials(c.name) }}</div>
						<span v-if="!c.isGroup && c.isOnline" class="status-dot online"></span>
						<span v-if="!c.isGroup && !c.isOnline" class="status-dot offline"></span>
					</div>
					
					<div class="conv-info">
						<div class="d-flex justify-content-between align-items-baseline mb-1">
							<div class="fw-bold text-truncate" style="max-width: 140px;">{{ c.name }}</div>
							<div class="text-muted fs-8">{{ formatTime(c.latestMessageTimestamp) }}</div>
						</div>
						<div class="text-muted fs-7 text-truncate">
							<span v-if="c.latestMessagePreview === '[Photo]'">📷 Photo</span>
							<span v-else>{{ c.latestMessagePreview || 'No messages' }}</span>
						</div>
					</div>
				</div>
			</div>
			
			<!-- User Profile Footer -->
			<div class="user-profile-footer mt-auto p-3 d-flex align-items-center justify-content-between" style="background: var(--bg-panel); min-height: 85px;">
				<div class="d-flex align-items-center">
					<img v-if="currentUser.photoUrl" :src="resolveAvatar(currentUser.photoUrl)" class="avatar sm shadow-sm me-2" style="width:40px;height:40px;object-fit:cover;" />
					<div v-else class="avatar sm shadow-sm me-2" style="width:40px;height:40px; border-radius:50%; display:flex; align-items:center; justify-content:center; background: linear-gradient(135deg, #4f46e5, #ec4899); color: white;">{{ getInitials(currentUser.username) }}</div>
					<div>
						<div class="fw-bold">{{ currentUser.username }}</div>
						<div class="text-success fs-8 d-flex align-items-center">
							<span style="display:inline-block; width:6px; height:6px; background:#10b981; border-radius:50%; margin-right:4px;"></span> Online
						</div>
					</div>
				</div>
				<div class="d-flex align-items-center">
					<button class="btn btn-sm text-muted p-2" @click="toggleTheme" title="Toggle Theme">
						<svg class="feather">
							<use v-if="isDarkMode" href="/feather-sprite-v4.29.0.svg#sun"/>
							<use v-else href="/feather-sprite-v4.29.0.svg#moon"/>
						</svg>
					</button>
					<button class="btn btn-sm text-muted p-2" @click="showProfileModal = true" title="Settings"><svg class="feather"><use href="/feather-sprite-v4.29.0.svg#settings"/></svg></button>
				</div>
			</div>
		</div>

		<!-- Main Chat Pane -->
		<div class="chat-pane">
			<div v-if="!activeConversation" class="empty-state">
				<div class="empty-icon-graphic mb-2">
					<svg width="180" height="150" viewBox="0 0 180 150" fill="none" xmlns="http://www.w3.org/2000/svg">
						<path d="M40 30 L43 38 L51 41 L43 44 L40 52 L37 44 L29 41 L37 38 Z" fill="#c4b5fd" opacity="0.6"/>
						<path d="M140 70 L142 75 L147 77 L142 79 L140 84 L138 79 L133 77 L138 75 Z" fill="#c4b5fd" opacity="0.6"/>
						<path d="M85 10 L87 15 L92 17 L87 19 L85 24 L83 19 L78 17 L83 15 Z" fill="#c4b5fd" opacity="0.8"/>
						<path d="M125 35 L126.5 38.5 L130 40 L126.5 41.5 L125 45 L123.5 41.5 L120 40 L123.5 38.5 Z" fill="#c4b5fd" opacity="0.6"/>
						<g filter="url(#shadow-bubble)">
							<path d="M70 50 H120 C128.3 50 135 56.7 135 65 V95 C135 103.3 128.3 110 120 110 V120 L105 110 H70 C61.7 110 55 103.3 55 95 V65 C55 56.7 61.7 50 70 50 Z" fill="white"/>
						</g>
						<path d="M45 40 H95 C103.3 40 110 46.7 110 55 V85 C110 93.3 103.3 100 95 100 H70 L55 110 V100 C46.7 100 40 93.3 40 85 V55 C40 46.7 46.7 40 45 40 Z" fill="url(#grad-bubble)"/>
						<circle cx="65" cy="70" r="4" fill="white"/>
						<circle cx="77.5" cy="70" r="4" fill="white"/>
						<circle cx="90" cy="70" r="4" fill="white"/>
						<defs>
							<linearGradient id="grad-bubble" x1="40" y1="40" x2="110" y2="110" gradientUnits="userSpaceOnUse">
								<stop stop-color="#4f46e5" />
								<stop offset="1" stop-color="#ec4899" />
							</linearGradient>
							<filter id="shadow-bubble" x="40" y="40" width="120" height="100" filterUnits="userSpaceOnUse" color-interpolation-filters="sRGB">
								<feDropShadow dx="0" dy="5" stdDeviation="10" flood-opacity="0.1" />
							</filter>
						</defs>
					</svg>
				</div>
				<h2 class="mt-2 fw-bold" style="color: var(--text-primary);">Start chatting!</h2>
				<p class="text-muted text-center" style="max-width: 300px;">Select a conversation from the sidebar to begin.</p>
			</div>
			
			<div v-else class="chat-container">
				<!-- Chat Header -->
				<div class="chat-header">
					<div class="d-flex align-items-center">
						<img v-if="activeConversation.photoUrl" :src="resolveAvatar(activeConversation.photoUrl)" class="avatar me-3" />
						<div v-else class="avatar me-3">{{ getInitials(activeConversation.name) }}</div>
						<div>
							<div class="fw-bold fs-5">{{ activeConversation.name }}</div>
							<div v-if="!activeConversation.isGroup" :class="activeConversation.isOnline ? 'text-success' : 'text-danger'" class="fs-7 d-flex align-items-center">
								<span :style="{ display: 'inline-block', width: '8px', height: '8px', borderRadius: '50%', marginRight: '4px', background: activeConversation.isOnline ? '#10b981' : '#ef4444' }"></span>
								{{ activeConversation.isOnline ? 'Online' : 'Offline' }}
							</div>
						</div>
					</div>
				</div>

				<!-- Messages Area -->
				<div class="messages-area">
					<LoadingSpinner :loading="loadingMessages" class="mt-3" />
					
					<div v-if="!loadingMessages && messages.length === 0" class="d-flex flex-column align-items-center justify-content-center h-100" style="opacity: 0.9;">
						<div class="empty-icon-graphic mb-1" style="transform: scale(0.85);">
							<svg width="180" height="150" viewBox="0 0 180 150" fill="none" xmlns="http://www.w3.org/2000/svg">
								<path d="M40 30 L43 38 L51 41 L43 44 L40 52 L37 44 L29 41 L37 38 Z" fill="#c4b5fd" opacity="0.6"/>
								<path d="M140 70 L142 75 L147 77 L142 79 L140 84 L138 79 L133 77 L138 75 Z" fill="#c4b5fd" opacity="0.6"/>
								<path d="M85 10 L87 15 L92 17 L87 19 L85 24 L83 19 L78 17 L83 15 Z" fill="#c4b5fd" opacity="0.8"/>
								<path d="M125 35 L126.5 38.5 L130 40 L126.5 41.5 L125 45 L123.5 41.5 L120 40 L123.5 38.5 Z" fill="#c4b5fd" opacity="0.6"/>
								<g filter="url(#shadow-bubble-inner)">
									<path d="M70 50 H120 C128.3 50 135 56.7 135 65 V95 C135 103.3 128.3 110 120 110 V120 L105 110 H70 C61.7 110 55 103.3 55 95 V65 C55 56.7 61.7 50 70 50 Z" fill="white"/>
								</g>
								<path d="M45 40 H95 C103.3 40 110 46.7 110 55 V85 C110 93.3 103.3 100 95 100 H70 L55 110 V100 C46.7 100 40 93.3 40 85 V55 C40 46.7 46.7 40 45 40 Z" fill="url(#grad-bubble-inner)"/>
								<circle cx="65" cy="70" r="4" fill="white"/>
								<circle cx="77.5" cy="70" r="4" fill="white"/>
								<circle cx="90" cy="70" r="4" fill="white"/>
								<defs>
									<linearGradient id="grad-bubble-inner" x1="40" y1="40" x2="110" y2="110" gradientUnits="userSpaceOnUse">
										<stop stop-color="#4f46e5" />
										<stop offset="1" stop-color="#ec4899" />
									</linearGradient>
									<filter id="shadow-bubble-inner" x="40" y="40" width="120" height="100" filterUnits="userSpaceOnUse" color-interpolation-filters="sRGB">
										<feDropShadow dx="0" dy="5" stdDeviation="10" flood-opacity="0.1" />
									</filter>
								</defs>
							</svg>
						</div>
						<h3 class="fw-bold fs-4" style="color: var(--text-primary);">Start the conversation!</h3>
						<p class="text-muted text-center mx-auto" style="max-width: 300px; font-size: 0.95rem;">
							Send a message to {{ activeConversation.name }} to start chatting.
						</p>
					</div>
					
					<div v-for="m in messages.slice().reverse()" :key="m.identifier" class="message-wrapper" :class="{ 'mine': m.sender.identifier === myUserId }">
						
						<!-- Sender Info for others in Group -->
						<div v-if="m.sender.identifier !== myUserId && activeConversation.isGroup" class="message-sender-info">
							<img v-if="m.sender.photoUrl" :src="resolveAvatar(m.sender.photoUrl)" class="avatar sm me-1" style="width:20px;height:20px;"/>
							<span class="fs-8 text-muted fw-bold">{{ m.sender.username }}</span>
						</div>

						<div class="message-bubble" :class="{ 'mine': m.sender.identifier === myUserId }">
							<!-- Forwarded Marker -->
							<div v-if="m.forwarded" class="text-muted fs-8 mb-1 fst-italic">
								<svg class="feather" style="width:12px;height:12px;margin-right:2px;"><use href="/feather-sprite-v4.29.0.svg#corner-up-right"/></svg> Forwarded
							</div>

							<!-- Reply context -->
							<div v-if="m.replyTo" class="reply-context">
								<div class="reply-sender">{{ getRepliedMessage(m.replyTo)?.sender.username || 'Someone' }}</div>
								<div class="reply-text text-truncate">
									<span v-if="getRepliedMessage(m.replyTo)?.photo">📷 Photo </span>
									{{ getRepliedMessage(m.replyTo)?.content || '' }}
								</div>
							</div>
							
							<!-- Content -->
							<div v-if="m.photo" class="message-photo">
								<img :src="resolveAvatar(m.photo)" alt="Photo" />
							</div>
							<div v-if="m.content" class="message-text" :class="{'mt-2': m.photo}">{{ m.content }}</div>
							
							<!-- Meta: Time + Status -->
							<div class="message-meta">
								<span>{{ formatTime(m.timestamp) }}</span>
								<span v-if="m.sender.identifier === myUserId" class="ms-1 status-check">
									<svg v-if="m.status !== 'read'" class="feather" style="color: #3b82f6;"><use href="/feather-sprite-v4.29.0.svg#check"/></svg>
									<span v-if="m.status === 'read'" class="double-check" style="color: #3b82f6; display: inline-flex;">
										<svg class="feather"><use href="/feather-sprite-v4.29.0.svg#check"/></svg>
										<svg class="feather" style="margin-left: -8px;"><use href="/feather-sprite-v4.29.0.svg#check"/></svg>
									</span>
								</span>
							</div>

							<!-- Reactions -->
							<div v-if="m.reactions && m.reactions.length > 0" class="reactions-list">
								<span v-for="r in m.reactions" :key="r.user.identifier + r.emoticon" class="reaction-pill cursor-pointer" :title="r.user.username" @click="showReactionAuthors(r.emoticon, m.reactions)">
									{{ r.emoticon }}
								</span>
							</div>
						</div>

						<!-- Actions Menu (Hover) -->
						<div class="message-actions">
							<button @click="replyTo = m" title="Reply"><svg class="feather"><use href="/feather-sprite-v4.29.0.svg#corner-up-left"/></svg></button>
							<button @click="openForward(m.identifier)" title="Forward"><svg class="feather"><use href="/feather-sprite-v4.29.0.svg#corner-up-right"/></svg></button>
							<button v-if="m.sender.identifier === myUserId" @click="deleteMsg(m.identifier)" title="Delete" class="text-danger"><svg class="feather"><use href="/feather-sprite-v4.29.0.svg#trash-2"/></svg></button>
							<div class="emoji-picker-container">
								<button title="React"><svg class="feather"><use href="/feather-sprite-v4.29.0.svg#smile"/></svg></button>
								<div class="emoji-picker">
									<span v-for="emoji in EMOJIS" :key="emoji" 
										class="emoji-btn" :class="{'active': hasUserReacted(m, emoji)}"
										@click="toggleReaction(m.identifier, emoji, hasUserReacted(m, emoji))">
										{{ emoji }}
									</span>
								</div>
							</div>
						</div>
					</div>
				</div>

				<!-- Input Area -->
				<div class="chat-input-area">
					<div v-if="replyTo" class="reply-preview">
						<div>
							<strong>Replying to {{ replyTo.sender.username }}</strong>
							<div class="text-truncate text-muted fs-8">{{ replyTo.content }}</div>
						</div>
						<button class="btn btn-sm text-muted p-0" @click="replyTo = null"><svg class="feather"><use href="/feather-sprite-v4.29.0.svg#x"/></svg></button>
					</div>
					
					<div v-if="messagePhoto" class="photo-preview mb-2">
						<span class="text-info fs-7"><svg class="feather me-1"><use href="/feather-sprite-v4.29.0.svg#image"/></svg> Photo attached</span>
						<button class="btn btn-sm text-danger p-0 ms-2" @click="clearPhoto"><svg class="feather"><use href="/feather-sprite-v4.29.0.svg#x"/></svg></button>
					</div>

					<div class="d-flex align-items-center">
						<label class="btn glass-btn-secondary rounded-circle p-2 me-2 mb-0 cursor-pointer d-flex align-items-center justify-content-center" title="Attach Photo" style="width:42px; height:42px; flex-shrink: 0;">
							<input type="file" id="photo-upload" class="d-none" accept="image/*" @change="handlePhotoSelect" />
							<svg class="feather" style="margin:0; flex-shrink: 0;"><use href="/feather-sprite-v4.29.0.svg#paperclip"/></svg>
						</label>
						<input v-model="messageText" @keyup.enter="sendMessage" type="text" class="form-control border-0 rounded-pill flex-grow-1" style="background: #f1f5f9; padding: 12px 20px;" placeholder="Type a message..." />
						<button class="btn btn-primary ms-3 rounded-circle d-flex align-items-center justify-content-center text-white" style="width:48px; height:48px; background: linear-gradient(135deg, var(--accent-primary), var(--accent-secondary)); border: none; flex-shrink:0;" @click="sendMessage" :disabled="!messageText.trim() && !messagePhoto">
							<svg class="feather" style="margin:0; width: 20px; height: 20px;"><use href="/feather-sprite-v4.29.0.svg#send"/></svg>
						</button>
					</div>
				</div>
			</div>
		</div>

		<!-- Modals -->

		<!-- Profile Modal -->
		<div v-if="showProfileModal" class="modal-overlay" @click.self="showProfileModal = false">
			<div class="modal-content-glass">
				<h4 class="mb-4">Profile & Settings</h4>
				<div class="mb-3">
					<label class="form-label text-muted fs-7">Username</label>
					<input v-model="profileForm.name" type="text" class="form-control glass-input" />
				</div>
				<div class="mb-4">
					<label class="form-label text-muted fs-7">Profile Photo</label>
					<input type="file" class="form-control glass-input" accept="image/*" @change="e => profileForm.photo = e.target.files[0]" />
				</div>
				<div class="d-flex justify-content-between align-items-center mt-4">
					<button class="btn btn-outline-danger" @click="emit('logout')">Log Out</button>
					<div class="d-flex gap-2">
						<button class="glass-btn-secondary" @click="showProfileModal = false">Cancel</button>
						<button class="glass-btn" @click="saveProfile">Save</button>
					</div>
				</div>
			</div>
		</div>

		<!-- New Chat Modal -->
		<div v-if="showNewChatModal" class="modal-overlay" @click.self="showNewChatModal = false">
			<div class="modal-content-glass" style="max-width: 500px;">
				<div class="d-flex justify-content-between align-items-center mb-4">
					<h4 class="mb-0">Start New Chat</h4>
					<button class="btn btn-sm text-muted" @click="showNewChatModal = false"><svg class="feather"><use href="/feather-sprite-v4.29.0.svg#x"/></svg></button>
				</div>
				<div class="position-relative mb-3">
					<svg class="feather position-absolute" style="left: 14px; top: 12px; color: var(--text-secondary); width: 16px; height: 16px;"><use href="/feather-sprite-v4.29.0.svg#search"/></svg>
					<input v-model="searchQuery" @input="doSearch" type="text" class="form-control glass-input w-100" style="padding-left: 40px;" placeholder="Search users by name..." />
				</div>
				<div class="search-results-box" style="max-height:250px; overflow-y:auto;" v-if="searchQuery">
					<div v-if="searchResults.length === 0" class="text-center text-muted p-3">No users found.</div>
					<div v-for="u in searchResults" :key="u.identifier" class="d-flex align-items-center justify-content-between p-2 border-bottom cursor-pointer" @click="startDirectChat(u); showNewChatModal = false;">
						<div class="d-flex align-items-center">
							<img v-if="u.photoUrl" :src="resolveAvatar(u.photoUrl)" class="avatar sm me-3" />
							<div v-else class="avatar sm me-3">{{ getInitials(u.username) }}</div>
							<span class="fw-bold">{{ u.username }}</span>
						</div>
					</div>
				</div>
			</div>
		</div>

		<!-- Create Group Modal -->
		<div v-if="showGroupModal" class="modal-overlay" @click.self="showGroupModal = false">
			<div class="modal-content-glass" style="max-width: 500px;">
				<h4 class="mb-3">Create Group</h4>
				<input v-model="groupForm.name" type="text" class="form-control glass-input mb-3" placeholder="Group Subject" />
				
				<label class="form-label text-muted fs-7">Add Members</label>
				<input v-model="groupForm.search" @input="doGroupSearch" type="text" class="form-control glass-input mb-2" placeholder="Search users..." />
				
				<div class="search-results-box mb-3 border border-secondary rounded p-2" style="max-height:150px; overflow-y:auto;" v-if="groupForm.search">
					<div v-for="u in groupForm.results" :key="u.identifier" class="d-flex align-items-center justify-content-between p-1 border-bottom border-dark">
						<span>{{ u.username }}</span>
						<button class="btn btn-sm btn-outline-primary py-0 px-2" @click="toggleGroupUserSelect(u)">
							{{ groupForm.selectedUsers.some(x => x.identifier === u.identifier) ? 'Remove' : 'Add' }}
						</button>
					</div>
				</div>

				<div class="selected-members mb-4 d-flex flex-wrap gap-1">
					<span v-for="u in groupForm.selectedUsers" :key="u.identifier" class="badge bg-primary rounded-pill">
						{{ u.username }} <span style="cursor:pointer" @click="toggleGroupUserSelect(u)">&times;</span>
					</span>
				</div>

				<div class="d-flex justify-content-end gap-2">
					<button class="glass-btn-secondary" @click="showGroupModal = false">Cancel</button>
					<button class="glass-btn" @click="createGroup" :disabled="!groupForm.name.trim()">Create</button>
				</div>
			</div>
		</div>

		<!-- Group Info Modal -->
		<div v-if="showGroupInfoModal && activeConversation" class="modal-overlay" @click.self="showGroupInfoModal = false">
			<div class="modal-content-glass" style="max-width: 500px;">
				<h4 class="mb-3">Group Info</h4>
				<div class="mb-3">
					<label class="text-muted fs-7">Name</label>
					<input v-model="groupInfoForm.name" type="text" class="form-control glass-input" :placeholder="activeConversation.name" />
				</div>
				<div class="mb-3">
					<label class="text-muted fs-7">Group Photo</label>
					<input type="file" class="form-control glass-input" accept="image/*" @change="e => groupInfoForm.photo = e.target.files[0]" />
				</div>
				<button class="glass-btn w-100 mb-4" @click="saveGroupInfo">Update Info</button>

				<hr class="border-secondary"/>
				<h6 class="mt-3">Add Member</h6>
				<input v-model="groupInfoForm.search" @input="doGroupInfoSearch" type="text" class="form-control glass-input mb-2" placeholder="Search users to add..." />
				<div class="search-results-box mb-3 border border-secondary rounded p-2" style="max-height:150px; overflow-y:auto;" v-if="groupInfoForm.search">
					<div v-for="u in groupInfoForm.results" :key="u.identifier" class="d-flex align-items-center justify-content-between p-1 border-bottom border-dark">
						<span>{{ u.username }}</span>
						<button class="btn btn-sm btn-primary py-0 px-2" @click="addGroupMember(u)">Add</button>
					</div>
				</div>

				<div class="mt-4 pt-3 border-top border-secondary text-center">
					<button class="glass-btn-danger w-100" @click="leaveGroup">Leave Group</button>
				</div>
			</div>
		</div>

		<!-- Forward Modal -->
		<div v-if="showForwardModal" class="modal-overlay" @click.self="showForwardModal = false">
			<div class="modal-content-glass">
				<h4 class="mb-3">Forward Message</h4>
				
				<!-- Search All Users -->
				<div class="mb-3">
					<label class="text-muted fs-7">Forward to any user:</label>
					<input v-model="forwardForm.search" @input="doForwardSearch" type="text" class="form-control glass-input mb-2" placeholder="Search users to forward to..." />
					<div class="search-results-box mb-2 border border-secondary rounded p-2" style="max-height:120px; overflow-y:auto;">
						<div v-for="u in forwardForm.results" :key="u.identifier" class="d-flex align-items-center justify-content-between p-1 border-bottom border-dark">
							<span>{{ u.username }}</span>
							<button class="btn btn-sm btn-primary py-0 px-2" @click="doForwardToUser(u)">Forward</button>
						</div>
					</div>
				</div>

				<p class="text-muted fs-7 mb-1">Or select an existing conversation:</p>
				<div class="conv-list" style="max-height: 200px; border: 1px solid rgba(255,255,255,0.1); border-radius: 8px;">
					<div v-for="c in conversations" :key="c.identifier" class="conv-item p-2" @click="doForward(c.identifier)">
						<div class="fw-bold fs-7">{{ c.name }}</div>
					</div>
				</div>
				<button class="glass-btn-secondary w-100 mt-3" @click="showForwardModal = false">Cancel</button>
			</div>
		</div>

	</div>
</template>

<style scoped>
.app-layout {
	display: flex;
	height: 100vh;
	width: 100vw;
	background: var(--bg-dark);
}

.icon-sidebar {
	width: 72px;
	min-width: 72px;
	height: 100%;
	background: var(--bg-panel);
	z-index: 15;
}

.icon-btn {
	background: transparent;
	border: none;
	width: 48px;
	height: 48px;
	border-radius: 16px;
	display: flex;
	align-items: center;
	justify-content: center;
	color: var(--text-secondary);
	transition: all 0.2s;
}
.icon-btn:hover { background: #f8fafc; color: var(--accent-primary); }
.icon-btn.active {
	background: #f3e8ff;
	color: var(--accent-primary);
}

.sidebar {
	position: static !important;
	width: 350px !important;
	min-width: 350px !important;
	max-width: 350px !important;
	height: 100% !important;
	border-right: 1px solid var(--border-light) !important;
	background: var(--bg-panel) !important;
	display: flex !important;
	flex-direction: column !important;
	z-index: 10;
	padding: 0 !important;
	box-shadow: none !important;
}

.sidebar-header {
	padding: 1rem;
	background: transparent;
	border-bottom: 1px solid var(--border-light);
}

.fs-7 { font-size: 0.85rem; }
.fs-8 { font-size: 0.75rem; }
.cursor-pointer { cursor: pointer; }

.search-dropdown {
	position: absolute;
	top: 100%;
	left: 0;
	right: 0;
	background: var(--bg-dark);
	border: 1px solid var(--border-light);
	border-radius: 8px;
	margin-top: 4px;
	max-height: 250px;
	overflow-y: auto;
	z-index: 100;
	box-shadow: 0 10px 25px rgba(0,0,0,0.5);
}

.search-item {
	display: flex;
	align-items: center;
	padding: 0.5rem 1rem;
	cursor: pointer;
	border-bottom: 1px solid rgba(255,255,255,0.05);
}
.search-item:hover {
	background: rgba(255,255,255,0.05);
}

.conv-list {
	flex-grow: 1;
	overflow-y: auto;
}

.conv-item {
	display: flex;
	align-items: center;
	padding: 0.85rem 1rem;
	cursor: pointer;
	transition: all 0.25s cubic-bezier(0.16, 1, 0.3, 1);
	border-bottom: 1px solid rgba(0,0,0,0.02);
	margin: 0 8px;
	border-radius: 12px;
}
.conv-item:hover { 
	background: rgba(0,0,0,0.02); 
	transform: translateY(-2px) scale(1.01);
	box-shadow: 0 4px 12px rgba(0,0,0,0.04);
}
.conv-item.active { 
	background: rgba(79, 70, 229, 0.1); 
	border-left: 3px solid var(--accent-primary); 
	border-bottom: none;
}

.conv-info { flex-grow: 1; overflow: hidden; }

.chat-pane {
	flex-grow: 1;
	display: flex;
	flex-direction: column;
	position: relative;
	background: var(--bg-dark);
}

.empty-state {
	flex-grow: 1;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
}
.empty-icon {
	width: 120px; height: 120px;
	border-radius: 50%;
	background: linear-gradient(135deg, rgba(79,70,229,0.2), rgba(236,72,153,0.2));
	display: flex; align-items: center; justify-content: center;
	color: #a5b4fc;
}

.chat-container {
	display: flex;
	flex-direction: column;
	height: 100%;
}

.chat-header {
	padding: 1rem 1.5rem;
	border-bottom: 1px solid var(--border-light);
	background: var(--bg-panel);
	display: flex;
	justify-content: space-between;
	align-items: center;
}

.messages-area {
	flex-grow: 1;
	overflow-y: auto;
	padding: 1.5rem;
	display: flex;
	flex-direction: column-reverse; /* Shows newest at bottom if normal, but since DOM order is reversed it shows newest at bottom visually. Wait, if we use column-reverse, the first DOM element is at the bottom! */
	gap: 1rem;
}

.message-wrapper {
	display: flex;
	flex-direction: column;
	align-items: flex-start;
	position: relative;
	max-width: 75%;
	animation: slideUpFade 0.3s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}
.message-wrapper.mine {
	align-self: flex-end;
	align-items: flex-end;
}

.message-sender-info {
	margin-bottom: 2px;
	margin-left: 4px;
}

.message-bubble {
	padding: 0.8rem 1.2rem;
	border-radius: 20px;
	background: var(--bg-panel-hover);
	border: none;
	color: var(--text-primary);
	border-bottom-left-radius: 4px;
	position: relative;
	box-shadow: none;
}
.message-bubble.mine {
	background: var(--accent-light);
	border: none;
	border-bottom-right-radius: 4px;
	border-bottom-left-radius: 20px;
	color: var(--text-primary) !important;
}

.reply-context {
	background: rgba(0,0,0,0.05);
	border-left: 3px solid var(--accent-primary);
	padding: 4px 8px;
	border-radius: 4px;
	font-size: 0.8rem;
	margin-bottom: 6px;
}
.reply-sender { font-weight: bold; color: var(--accent-primary); }

.message-bubble.mine .reply-context {
	background: rgba(255, 255, 255, 0.15);
	border-left: 3px solid white;
}
.message-bubble.mine .reply-sender {
	color: white;
}

.message-photo img {
	max-width: 100%;
	max-height: 250px;
	border-radius: 12px;
	margin-top: 5px;
}

.message-text {
	word-break: break-word;
	line-height: 1.4;
}

.message-meta {
	font-size: 0.65rem;
	color: var(--text-secondary);
	display: flex;
	justify-content: flex-end;
	align-items: center;
	margin-top: 4px;
}
.message-bubble.mine .message-meta {
	color: var(--text-secondary);
}
.message-bubble.mine .status-check {
	color: var(--accent-primary);
}
.status-check .feather {
	width: 12px; height: 12px;
}

.reactions-list {
	display: flex;
	flex-wrap: wrap;
	gap: 4px;
	margin-top: 4px;
	position: absolute;
	bottom: -12px;
	right: 10px;
}
.message-bubble.mine .reactions-list {
	right: auto;
	left: 10px;
}
.reaction-pill {
	background: var(--bg-panel);
	border: 1px solid var(--border-light);
	border-radius: 12px;
	padding: 2px 8px;
	font-size: 0.8rem;
	cursor: pointer;
	color: var(--text-primary);
	box-shadow: 0 2px 5px rgba(0,0,0,0.05);
	animation: popIn 0.35s cubic-bezier(0.175, 0.885, 0.32, 1.275);
	transition: transform 0.2s;
}
.reaction-pill:hover {
	transform: scale(1.1);
}

/* Actions Menu */
.message-actions {
	display: flex;
	gap: 6px;
	opacity: 0;
	transition: all 0.2s ease;
	position: absolute;
	top: 50%;
	transform: translateY(-50%);
	left: 100%;
	margin-left: 8px;
	background: var(--bg-panel);
	padding: 6px;
	border-radius: 12px;
	border: 1px solid var(--border-light);
	box-shadow: 0 4px 12px rgba(0,0,0,0.08);
	z-index: 20;
}
.message-wrapper.mine .message-actions {
	left: auto;
	right: 100%;
	margin-left: 0;
	margin-right: 8px;
}
.message-wrapper:hover .message-actions {
	opacity: 1;
}
.message-actions button {
	background: var(--btn-sec-bg);
	border: 1px solid var(--border-light);
	color: var(--accent-primary);
	padding: 6px;
	cursor: pointer;
	border-radius: 8px;
	display: flex;
	align-items: center;
	justify-content: center;
	transition: all 0.2s;
}
.message-actions button:hover {
	background: var(--accent-primary);
	color: white;
	border-color: var(--accent-primary);
	transform: translateY(-2px);
	box-shadow: 0 4px 8px rgba(79, 70, 229, 0.2);
}
.message-actions button.text-danger {
	color: var(--danger) !important;
}
.message-actions button.text-danger:hover {
	background: var(--danger);
	color: white !important;
	border-color: var(--danger);
	box-shadow: 0 4px 8px rgba(239, 68, 68, 0.2);
}
.message-actions button .feather { width: 16px; height: 16px; }

/* Emoji Picker */
.emoji-picker-container { position: relative; }
.emoji-picker {
	position: absolute;
	bottom: 100%;
	left: 50%;
	transform: translateX(-50%);
	background: var(--bg-dark);
	border: 1px solid var(--border-light);
	border-radius: 20px;
	padding: 6px;
	display: flex;
	gap: 4px;
	opacity: 0;
	visibility: hidden;
	transition: all 0.2s;
}
.emoji-picker-container:hover .emoji-picker {
	opacity: 1;
	visibility: visible;
}
.emoji-btn {
	cursor: pointer;
	font-size: 1.1rem;
	padding: 2px 4px;
	border-radius: 8px;
	transition: transform 0.1s;
}
.emoji-btn:hover { transform: scale(1.2); }
.emoji-btn.active { background: rgba(79,70,229,0.3); }

/* Input Area */
.chat-input-area {
	padding: 1rem 1.5rem;
	background: var(--bg-panel);
	border-top: 1px solid var(--border-light);
	min-height: 85px;
	display: flex;
	flex-direction: column;
	justify-content: center;
}

.reply-preview {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 8px 12px;
	background: rgba(79,70,229,0.1);
	border-left: 3px solid var(--accent-primary);
	border-radius: 8px;
	margin-bottom: 8px;
	font-size: 0.85rem;
}

</style>
