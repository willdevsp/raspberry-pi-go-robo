package web

import "embed"

//go:embed index.html static/app.js
var Assets embed.FS
