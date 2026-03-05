package export

import (
	"context"
	"os"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// PDFExporter exports HTML to PDF using headless Chrome.
type PDFExporter struct {
	htmlExporter *HTMLExporter
}

// NewPDFExporter creates a new PDF exporter.
func NewPDFExporter() *PDFExporter {
	return &PDFExporter{
		htmlExporter: NewHTMLExporter(),
	}
}

// Export exports markdown content to a PDF file.
func (e *PDFExporter) Export(content, outputPath string) error {
	// First convert markdown to HTML
	title := "Document"
	html, err := e.htmlExporter.ExportToString(content, title)
	if err != nil {
		return err
	}

	// Create Chrome context with timeout
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Generate PDF
	var pdfBuf []byte
	err = chromedp.Run(ctx,
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			frameTree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return err
			}

			return page.SetDocumentContent(frameTree.Frame.ID, html).Do(ctx)
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfBuf, _, err = page.PrintToPDF().
				WithPrintBackground(true).
				WithMarginTop(0.5).
				WithMarginBottom(0.5).
				WithMarginLeft(0.5).
				WithMarginRight(0.5).
				Do(ctx)
			return err
		}),
	)
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, pdfBuf, 0644)
}

// ExportWithOptions exports with custom PDF options.
func (e *PDFExporter) ExportWithOptions(content, outputPath string, opts PDFOptions) error {
	title := opts.Title
	if title == "" {
		title = "Document"
	}

	html, err := e.htmlExporter.ExportToString(content, title)
	if err != nil {
		return err
	}

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var pdfBuf []byte
	err = chromedp.Run(ctx,
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			frameTree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return err
			}
			return page.SetDocumentContent(frameTree.Frame.ID, html).Do(ctx)
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error

			printOpts := page.PrintToPDF().
				WithPrintBackground(true).
				WithMarginTop(opts.MarginTop).
				WithMarginBottom(opts.MarginBottom).
				WithMarginLeft(opts.MarginLeft).
				WithMarginRight(opts.MarginRight)

			if opts.Landscape {
				printOpts = printOpts.WithLandscape(true)
			}

			if opts.PageWidth > 0 && opts.PageHeight > 0 {
				printOpts = printOpts.WithPaperWidth(opts.PageWidth).WithPaperHeight(opts.PageHeight)
			}

			pdfBuf, _, err = printOpts.Do(ctx)
			return err
		}),
	)
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, pdfBuf, 0644)
}

// PDFOptions configures PDF export.
type PDFOptions struct {
	Title        string
	Landscape    bool
	PageWidth    float64 // inches
	PageHeight   float64 // inches
	MarginTop    float64 // inches
	MarginBottom float64 // inches
	MarginLeft   float64 // inches
	MarginRight  float64 // inches
}

// DefaultPDFOptions returns sensible default options.
func DefaultPDFOptions() PDFOptions {
	return PDFOptions{
		Title:        "Document",
		Landscape:    false,
		PageWidth:    8.5,  // Letter size
		PageHeight:   11.0, // Letter size
		MarginTop:    0.5,
		MarginBottom: 0.5,
		MarginLeft:   0.5,
		MarginRight:  0.5,
	}
}
