package export

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// HTMLExporter exports markdown to HTML.
type HTMLExporter struct {
	md goldmark.Markdown
}

// NewHTMLExporter creates a new HTML exporter.
func NewHTMLExporter() *HTMLExporter {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Typographer,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	return &HTMLExporter{md: md}
}

// htmlTemplate is the HTML template for export.
const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{TITLE}}</title>
    <style>
        :root {
            --bg: #ffffff;
            --fg: #1a1a1a;
            --accent: #0066cc;
            --code-bg: #f5f5f5;
            --border: #e0e0e0;
        }

        @media (prefers-color-scheme: dark) {
            :root {
                --bg: #1a1a1a;
                --fg: #e0e0e0;
                --accent: #66b3ff;
                --code-bg: #2d2d2d;
                --border: #404040;
            }
        }

        * {
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            line-height: 1.6;
            max-width: 800px;
            margin: 0 auto;
            padding: 2rem;
            background: var(--bg);
            color: var(--fg);
        }

        h1, h2, h3, h4, h5, h6 {
            margin-top: 2rem;
            margin-bottom: 1rem;
            font-weight: 600;
            line-height: 1.3;
        }

        h1 { font-size: 2rem; border-bottom: 2px solid var(--border); padding-bottom: 0.5rem; }
        h2 { font-size: 1.5rem; }
        h3 { font-size: 1.25rem; }

        p { margin: 1rem 0; }

        a {
            color: var(--accent);
            text-decoration: none;
        }

        a:hover {
            text-decoration: underline;
        }

        code {
            font-family: 'SF Mono', Consolas, 'Liberation Mono', Menlo, monospace;
            font-size: 0.9em;
            background: var(--code-bg);
            padding: 0.2em 0.4em;
            border-radius: 3px;
        }

        pre {
            background: var(--code-bg);
            padding: 1rem;
            border-radius: 6px;
            overflow-x: auto;
        }

        pre code {
            background: none;
            padding: 0;
        }

        blockquote {
            border-left: 4px solid var(--accent);
            margin: 1rem 0;
            padding: 0.5rem 1rem;
            background: var(--code-bg);
        }

        ul, ol {
            margin: 1rem 0;
            padding-left: 2rem;
        }

        li { margin: 0.5rem 0; }

        table {
            border-collapse: collapse;
            width: 100%;
            margin: 1rem 0;
        }

        th, td {
            border: 1px solid var(--border);
            padding: 0.75rem;
            text-align: left;
        }

        th {
            background: var(--code-bg);
            font-weight: 600;
        }

        hr {
            border: none;
            border-top: 1px solid var(--border);
            margin: 2rem 0;
        }

        img {
            max-width: 100%;
            height: auto;
        }
    </style>
</head>
<body>
{{CONTENT}}
</body>
</html>`

// Export exports markdown content to an HTML file.
func (e *HTMLExporter) Export(content, outputPath string) error {
	var buf bytes.Buffer
	if err := e.md.Convert([]byte(content), &buf); err != nil {
		return err
	}

	// Get title from filename
	title := filepath.Base(outputPath)
	title = strings.TrimSuffix(title, filepath.Ext(title))

	// Build HTML
	html := strings.Replace(htmlTemplate, "{{TITLE}}", title, 1)
	html = strings.Replace(html, "{{CONTENT}}", buf.String(), 1)

	return os.WriteFile(outputPath, []byte(html), 0644)
}

// ExportToString returns HTML as a string.
func (e *HTMLExporter) ExportToString(content, title string) (string, error) {
	var buf bytes.Buffer
	if err := e.md.Convert([]byte(content), &buf); err != nil {
		return "", err
	}

	html := strings.Replace(htmlTemplate, "{{TITLE}}", title, 1)
	html = strings.Replace(html, "{{CONTENT}}", buf.String(), 1)

	return html, nil
}
