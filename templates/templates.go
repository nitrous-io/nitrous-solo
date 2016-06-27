package templates

import "html/template"

var T *template.Template

func init() {
	T = template.Must(template.New("templates").Parse(rootIndex))
}
