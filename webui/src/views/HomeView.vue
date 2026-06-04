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
const showGroupInfoModal = ref(false)
const showForwardModal = ref(false)

// Forms
const profileForm = ref({ name: props.currentUser.username, photo: null })
const groupForm = ref({ name: '', search: '', results: [], selectedUsers: [] })
const groupInfoForm = ref({ name: '', photo: null, search: '', results: [] })
const forwardMsgId = ref(null)

// Polling interval
let pollInterval = null

onMounted(() => {
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
	} catch (e) {
		console.error("Failed to fetch conversations", e)
	}
}

async function searchUsers(query, listRef) {
	if (!query || query.length < 1) {
		listRef.value = []
		return
	}
	try {
		const res = await axios.get(`/users?username=${encodeURIComponent(query)}`)
		// Exclude self
		listRef.value = (res.data || []).filter(u => u.identifier !== myUserId.value)
	} catch (e) {
		console.error("Search failed", e)
	}
}

async function doSearch() {
	await searchUsers(searchQuery.value, searchResults)
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

function openForward(msgId) {
	forwardMsgId.value = msgId
	showForwardModal.value = true
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
		fetchConversations()
		openConversation(res.data)
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
		if (profileForm.value.photo) {
			const fd = new FormData()
			fd.append("photo", profileForm.value.photo)
			await axios.put('/users/me/photo', fd)
		}
		showProfileModal.value = false
		emit('updateUser', { name: profileForm.value.name.trim(), photo: profileForm.value.photo })
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
	if (url.startsWith('data:')) return url
	return __API_URL__ + url
}

function getRepliedMessage(replyId) {
	return messages.value.find(m => m.identifier === replyId)
}
</script>

<template>
	<div class="app-layout">
		<!-- Sidebar -->
		<div class="sidebar">
			<!-- Header -->
			<div class="sidebar-header">
				<div class="d-flex align-items-center w-100">
					<div class="avatar sm me-3 shadow-sm">{{ getInitials(currentUser.username) }}</div>
					<div class="fw-bold flex-grow-1 text-truncate fs-6" style="color: var(--accent-primary);">{{ currentUser.username }}</div>
					<div class="d-flex gap-2">
						<button class="glass-btn-secondary p-2 rounded-circle d-flex align-items-center justify-content-center" style="width:32px; height:32px;" @click="showProfileModal = true" title="Profile Settings">
							<svg class="feather" style="color: var(--accent-primary); width: 16px; height: 16px;"><use href="/feather-sprite-v4.29.0.svg#settings"/></svg>
						</button>
						<button class="glass-btn-secondary p-2 rounded-circle d-flex align-items-center justify-content-center" style="width:32px; height:32px;" @click="emit('logout')" title="Log Out">
							<svg class="feather" style="color: var(--danger); width: 16px; height: 16px;"><use href="/feather-sprite-v4.29.0.svg#log-out"/></svg>
						</button>
					</div>
				</div>
			</div>
			
			<!-- Search & Create Group -->
			<div class="p-3 border-bottom" style="border-color: var(--border-light) !important;">
				<div class="position-relative">
					<input v-model="searchQuery" @input="doSearch" type="text" class="form-control glass-input w-100" placeholder="Search users..." />
					<div v-if="searchQuery && searchResults.length > 0" class="search-dropdown">
						<div v-for="u in searchResults" :key="u.identifier" class="search-item" @click="startDirectChat(u)">
							<img v-if="u.photoUrl" :src="resolveAvatar(u.photoUrl)" class="avatar sm me-2" />
							<div v-else class="avatar sm me-2">{{ getInitials(u.username) }}</div>
							<span>{{ u.username }}</span>
						</div>
					</div>
					<div v-else-if="searchQuery" class="search-dropdown text-muted p-2 text-center fs-7">
						No users found.
					</div>
				</div>
				<button class="glass-btn-secondary w-100 mt-2 py-1 fs-7 rounded-3" @click="showGroupModal = true">
					<svg class="feather me-1"><use href="/feather-sprite-v4.29.0.svg#users"/></svg> New Group
				</button>
			</div>
			
			<!-- Conversations List -->
			<div class="conv-list">
				<div v-if="conversations.length === 0" class="text-center text-muted p-4 mt-4 fs-7">
					No conversations yet.<br/>Search for a user to start chatting!
				</div>
				<div v-for="c in conversations" :key="c.identifier" 
					class="conv-item" :class="{ 'active': activeConversation?.identifier === c.identifier }"
					@click="openConversation(c)">
					<img v-if="c.photoUrl" :src="resolveAvatar(c.photoUrl)" class="avatar me-3" />
					<div v-else class="avatar me-3">{{ getInitials(c.name) }}</div>
					
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
		</div>

		<!-- Main Chat Pane -->
		<div class="chat-pane">
			<div v-if="!activeConversation" class="empty-state">
				<div class="empty-icon"><svg class="feather" style="width:64px;height:64px;"><use href="/feather-sprite-v4.29.0.svg#message-circle"/></svg></div>
				<h2 class="mt-4 fw-bold">WASAText</h2>
				<p class="text-muted">Select a conversation to start chatting.</p>
			</div>
			
			<div v-else class="chat-container">
				<!-- Chat Header -->
				<div class="chat-header">
					<div class="d-flex align-items-center">
						<img v-if="activeConversation.photoUrl" :src="resolveAvatar(activeConversation.photoUrl)" class="avatar me-3" />
						<div v-else class="avatar me-3">{{ getInitials(activeConversation.name) }}</div>
						<div>
							<div class="fw-bold fs-5">{{ activeConversation.name }}</div>
							<div class="text-muted fs-7">{{ activeConversation.isGroup ? 'Group Chat' : 'Direct Chat' }}</div>
						</div>
					</div>
					<div v-if="activeConversation.isGroup">
						<button class="glass-btn-secondary p-2 rounded-circle" @click="showGroupInfoModal = true" title="Group Info">
							<svg class="feather"><use href="/feather-sprite-v4.29.0.svg#info"/></svg>
						</button>
					</div>
				</div>

				<!-- Messages Area -->
				<div class="messages-area">
					<LoadingSpinner :loading="loadingMessages" class="mt-3" />
					
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
								<span v-for="r in m.reactions" :key="r.user.identifier + r.emoticon" class="reaction-pill" :title="r.user.username">
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
						<label class="btn glass-btn-secondary rounded-circle p-2 me-2 mb-0 cursor-pointer" title="Attach Photo">
							<input type="file" id="photo-upload" class="d-none" accept="image/*" @change="handlePhotoSelect" />
							<svg class="feather"><use href="/feather-sprite-v4.29.0.svg#paperclip"/></svg>
						</label>
						<input v-model="messageText" @keyup.enter="sendMessage" type="text" class="form-control glass-input flex-grow-1" placeholder="Type a message..." />
						<button class="glass-btn ms-2 rounded-circle p-2" style="width:42px; height:42px;" @click="sendMessage" :disabled="!messageText.trim() && !messagePhoto">
							<svg class="feather" style="margin:0;"><use href="/feather-sprite-v4.29.0.svg#send"/></svg>
						</button>
					</div>
				</div>
			</div>
		</div>

		<!-- Modals -->

		<!-- Profile Modal -->
		<div v-if="showProfileModal" class="modal-overlay" @click.self="showProfileModal = false">
			<div class="modal-content-glass">
				<h4 class="mb-4">Edit Profile</h4>
				<div class="mb-3">
					<label class="form-label text-muted fs-7">Username</label>
					<input v-model="profileForm.name" type="text" class="form-control glass-input" />
				</div>
				<div class="mb-4">
					<label class="form-label text-muted fs-7">Profile Photo</label>
					<input type="file" class="form-control glass-input" accept="image/*" @change="e => profileForm.photo = e.target.files[0]" />
				</div>
				<div class="d-flex justify-content-end gap-2">
					<button class="glass-btn-secondary" @click="showProfileModal = false">Cancel</button>
					<button class="glass-btn" @click="saveProfile">Save</button>
				</div>
			</div>
		</div>

		<!-- Create Group Modal -->
		<div v-if="showGroupModal" class="modal-overlay" @click.self="showGroupModal = false">
			<div class="modal-content-glass" style="max-width: 500px;">
				<h4 class="mb-3">Create Group</h4>
				<input v-model="groupForm.name" type="text" class="form-control glass-input mb-3" placeholder="Group Subject" />
				
				<label class="form-label text-muted fs-7">Add Members</label>
				<input v-model="groupForm.search" @input="() => searchUsers(groupForm.search, groupForm.results)" type="text" class="form-control glass-input mb-2" placeholder="Search users..." />
				
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
				<input v-model="groupInfoForm.search" @input="() => searchUsers(groupInfoForm.search, groupInfoForm.results)" type="text" class="form-control glass-input mb-2" placeholder="Search users to add..." />
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
				<p class="text-muted fs-7">Select a conversation to forward to:</p>
				<div class="conv-list" style="max-height: 300px; border: 1px solid rgba(255,255,255,0.1); border-radius: 8px;">
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
	background-color: #f8fafc;
	background-image: radial-gradient(rgba(0, 0, 0, 0.06) 1px, transparent 1px);
	background-size: 20px 20px;
}

.sidebar {
	position: static !important;
	width: 320px !important;
	min-width: 320px !important;
	max-width: 320px !important;
	height: 100vh !important;
	border-right: 1px solid var(--border-light) !important;
	background: var(--bg-panel) !important;
	display: flex !important;
	flex-direction: column !important;
	backdrop-filter: blur(20px);
	z-index: 10;
	padding: 0 !important;
	box-shadow: none !important;
}

.sidebar-header {
	padding: 1rem;
	background: rgba(255,255,255,0.02);
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
	transition: background 0.2s;
	border-bottom: 1px solid rgba(255,255,255,0.02);
}
.conv-item:hover { background: rgba(255,255,255,0.03); }
.conv-item.active { background: rgba(79, 70, 229, 0.15); border-left: 3px solid var(--accent-primary); }

.conv-info { flex-grow: 1; overflow: hidden; }

.chat-pane {
	flex-grow: 1;
	display: flex;
	flex-direction: column;
	position: relative;
	background: rgba(0,0,0,0.2);
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
	background: rgba(255,255,255,0.02);
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
	padding: 0.6rem 1rem;
	border-radius: 18px;
	background: var(--bg-panel);
	border: 1px solid var(--border-light);
	color: var(--text-primary);
	border-bottom-left-radius: 4px;
	position: relative;
	box-shadow: 0 2px 10px rgba(0,0,0,0.1);
}
.message-bubble.mine {
	background: linear-gradient(135deg, var(--accent-primary), #6366f1);
	border: none;
	border-bottom-right-radius: 4px;
	border-bottom-left-radius: 18px;
	color: white !important;
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
	color: rgba(255, 255, 255, 0.75);
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
	background: var(--btn-sec-bg);
	border: 1px solid var(--border-light);
	border-radius: 10px;
	padding: 2px 6px;
	font-size: 0.75rem;
	cursor: pointer;
	color: var(--text-primary);
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
	color: var(--danger);
}
.message-actions button.text-danger:hover {
	background: var(--danger);
	color: white;
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
	background: rgba(255,255,255,0.02);
	border-top: 1px solid var(--border-light);
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
