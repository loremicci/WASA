package api

import (
	"github.com/loremicci/WASA/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"path/filepath"
)

func (rt *_router) getUpload(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	filename := ps.ByName("filename")
	http.ServeFile(w, r, filepath.Join("uploads", filename))
}
