package testutil

import (
	"html/template"

	"github.com/gin-gonic/gin"
)

// LoadTestTemplates registers minimal stub templates so c.HTML() doesn't panic.
// templateNames should be the names that handlers reference, e.g.
// "admin/pages/login.html", "public/pages/login.html".
func LoadTestTemplates(r *gin.Engine, templateNames ...string) {
	t := template.New("")
	for _, name := range templateNames {
		template.Must(t.New(name).Parse(`{{define "` + name + `"}}OK{{end}}`))
	}
	r.SetHTMLTemplate(t)
}
