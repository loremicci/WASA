package api

import (
	"encoding/base64"
	"encoding/json"
	"github.com/loremicci/WASA/service/api/reqcontext"
	"github.com/loremicci/WASA/service/database"
	"github.com/julienschmidt/httprouter"
	"io"
	"net/http"
)

func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userId := GetUserIDFromContext(r)
	conversationId := ps.ByName("conversationId")

	sender, err := rt.db.GetUserByID(userId)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var msg database.Message
	msg.ConversationID = conversationId
	msg.Sender = sender

	replyTo := r.FormValue("replyTo")
	text := r.FormValue("text")

	file, header, errFile := r.FormFile("photo")
	if errFile == nil {
		defer file.Close()
		bytes, err := io.ReadAll(file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "image/jpeg"
		}
		base64Str := base64.StdEncoding.EncodeToString(bytes)
		msg.Type = "photo"
		msg.Content = "data:" + contentType + ";base64," + base64Str
	} else if text != "" {
		msg.Type = "text"
		msg.Content = text
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sentMsg, err := rt.db.SendMessage(msg, replyTo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(sentMsg)
}

func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userId := GetUserIDFromContext(r)
	messageId := ps.ByName("messageId")

	if err := rt.db.DeleteMessage(messageId, userId); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userId := GetUserIDFromContext(r)
	messageId := ps.ByName("messageId")

	var req struct {
		ConversationId string `json:"conversationId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	msg, err := rt.db.ForwardMessage(messageId, req.ConversationId, userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(msg)
}

func (rt *_router) commentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userId := GetUserIDFromContext(r)
	messageId := ps.ByName("messageId")
	emoticon := ps.ByName("emoticon")

	if err := rt.db.CommentMessage(messageId, userId, emoticon); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (rt *_router) uncommentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userId := GetUserIDFromContext(r)
	messageId := ps.ByName("messageId")
	emoticon := ps.ByName("emoticon")

	if err := rt.db.UncommentMessage(messageId, userId, emoticon); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
