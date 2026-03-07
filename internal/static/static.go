package static

import "embed"

//go:embed style.css main.js
var FS embed.FS
