package api

import (
	"net/http"
	"path/filepath"
	"github.com/julienschmidt/httprouter"
	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
)

func (rt *_router) getUpload(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	filename := ps.ByName("filename")
	http.ServeFile(w, r, filepath.Join("uploads", filename))
}
