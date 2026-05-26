package api

import (
	"net/http"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {
	// Login
	rt.router.POST("/session", rt.wrap(rt.doLogin))

	// Users
	rt.router.GET("/users", rt.AuthMiddleware(rt.wrap(rt.searchUsers)))
	rt.router.PUT("/users/me/name", rt.AuthMiddleware(rt.wrap(rt.setMyUserName)))
	rt.router.PUT("/users/me/photo", rt.AuthMiddleware(rt.wrap(rt.setMyPhoto)))

	// Groups
	rt.router.POST("/groups", rt.AuthMiddleware(rt.wrap(rt.createGroup)))
	rt.router.PUT("/groups/:groupId/name", rt.AuthMiddleware(rt.wrap(rt.setGroupName)))
	rt.router.PUT("/groups/:groupId/photo", rt.AuthMiddleware(rt.wrap(rt.setGroupPhoto)))
	rt.router.POST("/groups/:groupId/members", rt.AuthMiddleware(rt.wrap(rt.addToGroup)))
	rt.router.DELETE("/groups/:groupId/members/me", rt.AuthMiddleware(rt.wrap(rt.leaveGroup)))

	// Conversations
	rt.router.GET("/conversations", rt.AuthMiddleware(rt.wrap(rt.getMyConversations)))
	rt.router.POST("/conversations", rt.AuthMiddleware(rt.wrap(rt.createConversation)))
	rt.router.GET("/conversations/:conversationId", rt.AuthMiddleware(rt.wrap(rt.getConversation)))

	// Messages
	rt.router.POST("/conversations/:conversationId/messages", rt.AuthMiddleware(rt.wrap(rt.sendMessage)))
	rt.router.DELETE("/messages/:messageId", rt.AuthMiddleware(rt.wrap(rt.deleteMessage)))
	rt.router.POST("/messages/:messageId/forward", rt.AuthMiddleware(rt.wrap(rt.forwardMessage)))
	
	// Reactions
	rt.router.PUT("/messages/:messageId/comments/:emoticon", rt.AuthMiddleware(rt.wrap(rt.commentMessage)))
	rt.router.DELETE("/messages/:messageId/comments/:emoticon", rt.AuthMiddleware(rt.wrap(rt.uncommentMessage)))

	// Uploads
	rt.router.GET("/uploads/:filename", rt.wrap(rt.getUpload))

	// Special routes
	rt.router.GET("/liveness", rt.liveness)

	return rt.router
}
