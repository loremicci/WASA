package api

import (
	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"path/filepath"
)

func (rt *_router) getUpload(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	filename := ps.ByName("filename")
	http.ServeFile(w, r, filepath.Join("uploads", filename))
}
