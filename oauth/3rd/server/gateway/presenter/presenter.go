package presenter

import "text/template"

var HTML = html{}

type html map[string]*template.Template
