# Package Plan: pkg/pdf

## Overview

A PDF generation package for creating documents, reports, and invoices. Supports HTML-to-PDF conversion, template-based generation, and programmatic PDF creation with tables, images, and styling.

## Goals

1. **HTML-to-PDF** - Convert HTML/CSS to PDF
2. **Templates** - Template-based PDF generation
3. **Programmatic API** - Build PDFs with code
4. **Tables & Charts** - Support complex layouts
5. **Images** - Embed images and logos
6. **Headers/Footers** - Page headers and footers
7. **Multiple Engines** - wkhtmltopdf, Chrome/Puppeteer, native

## Architecture

```
pkg/pdf/
├── pdf.go                # Core Generator interface
├── config.go             # Configuration
├── options.go            # Functional options
├── document.go           # Document builder
├── template.go           # Template support
├── errors.go             # Custom error types
├── engine/
│   ├── engine.go         # Engine interface
│   ├── wkhtml.go         # wkhtmltopdf
│   ├── chrome.go         # Chrome/Puppeteer
│   └── native.go         # Pure Go (gofpdf)
├── components/
│   ├── table.go          # Table component
│   ├── image.go          # Image component
│   ├── text.go           # Text/paragraph
│   └── chart.go          # Chart component
├── examples/
│   ├── basic/
│   ├── invoice/
│   ├── report/
│   └── from-html/
└── README.md
```

## Core Interfaces

```go
package pdf

import (
    "context"
    "io"
)

// Generator generates PDF documents
type Generator interface {
    // Generate creates PDF from document
    Generate(ctx context.Context, doc *Document) ([]byte, error)

    // GenerateFromHTML creates PDF from HTML
    GenerateFromHTML(ctx context.Context, html string, opts ...Option) ([]byte, error)

    // GenerateFromTemplate creates PDF from template
    GenerateFromTemplate(ctx context.Context, tmpl string, data interface{}, opts ...Option) ([]byte, error)

    // GenerateToWriter writes PDF to writer
    GenerateToWriter(ctx context.Context, doc *Document, w io.Writer) error
}

// Document represents a PDF document
type Document struct {
    // Metadata
    Title    string
    Author   string
    Subject  string
    Keywords []string

    // Page settings
    PageSize    PageSize
    Orientation Orientation
    Margins     Margins

    // Header/Footer
    Header *HeaderFooter
    Footer *HeaderFooter

    // Content
    Elements []Element
}

// Element is a document element
type Element interface {
    Render(ctx *RenderContext) error
}

// PageSize constants
type PageSize string
const (
    A4     PageSize = "A4"
    Letter PageSize = "Letter"
    Legal  PageSize = "Legal"
    A3     PageSize = "A3"
)

// Orientation constants
type Orientation string
const (
    Portrait  Orientation = "portrait"
    Landscape Orientation = "landscape"
)

// Margins in millimeters
type Margins struct {
    Top    float64
    Right  float64
    Bottom float64
    Left   float64
}

// HeaderFooter configuration
type HeaderFooter struct {
    HTML   string
    Height float64
}
```

## Document Builder

```go
// NewDocument creates a document builder
func NewDocument() *DocumentBuilder

type DocumentBuilder struct {
    doc *Document
}

func (b *DocumentBuilder) Title(title string) *DocumentBuilder
func (b *DocumentBuilder) Author(author string) *DocumentBuilder
func (b *DocumentBuilder) PageSize(size PageSize) *DocumentBuilder
func (b *DocumentBuilder) Landscape() *DocumentBuilder
func (b *DocumentBuilder) Margins(top, right, bottom, left float64) *DocumentBuilder
func (b *DocumentBuilder) Header(html string, height float64) *DocumentBuilder
func (b *DocumentBuilder) Footer(html string, height float64) *DocumentBuilder
func (b *DocumentBuilder) Add(elements ...Element) *DocumentBuilder
func (b *DocumentBuilder) Build() *Document

// Element constructors
func Text(content string, opts ...TextOption) Element
func Paragraph(content string, opts ...TextOption) Element
func Heading(level int, content string) Element
func Image(src string, opts ...ImageOption) Element
func Table(headers []string, rows [][]string, opts ...TableOption) Element
func PageBreak() Element
func Spacer(height float64) Element
func HTML(content string) Element
```

## Configuration

```go
// Config holds PDF generator configuration
type Config struct {
    // Engine: "wkhtml", "chrome", "native"
    Engine string `env:"PDF_ENGINE" default:"wkhtml"`

    // Default page size
    DefaultPageSize PageSize `env:"PDF_PAGE_SIZE" default:"A4"`

    // Default margins (mm)
    DefaultMargins Margins
}

// WkhtmlConfig for wkhtmltopdf
type WkhtmlConfig struct {
    // Path to wkhtmltopdf binary
    BinaryPath string `env:"WKHTML_PATH" default:"wkhtmltopdf"`

    // Enable JavaScript
    EnableJS bool `env:"WKHTML_ENABLE_JS" default:"true"`

    // JavaScript delay (ms)
    JSDelay int `env:"WKHTML_JS_DELAY" default:"200"`
}

// ChromeConfig for Chrome/Puppeteer
type ChromeConfig struct {
    // Chrome executable path
    ExecPath string `env:"CHROME_PATH"`

    // Remote debugging URL
    RemoteURL string `env:"CHROME_REMOTE_URL"`

    // Headless mode
    Headless bool `env:"CHROME_HEADLESS" default:"true"`
}
```

## Usage Examples

### HTML to PDF

```go
package main

import (
    "context"
    "os"
    "github.com/user/core-backend/pkg/pdf"
)

func main() {
    gen, _ := pdf.New(pdf.Config{
        Engine: "wkhtml",
    })

    ctx := context.Background()

    html := `
    <!DOCTYPE html>
    <html>
    <head>
        <style>
            body { font-family: Arial, sans-serif; }
            h1 { color: #333; }
            table { width: 100%; border-collapse: collapse; }
            th, td { border: 1px solid #ddd; padding: 8px; }
        </style>
    </head>
    <body>
        <h1>Sales Report</h1>
        <table>
            <tr><th>Product</th><th>Sales</th></tr>
            <tr><td>Widget A</td><td>$1,234</td></tr>
            <tr><td>Widget B</td><td>$5,678</td></tr>
        </table>
    </body>
    </html>
    `

    data, _ := gen.GenerateFromHTML(ctx, html,
        pdf.WithPageSize(pdf.A4),
        pdf.WithMargins(20, 15, 20, 15),
    )

    os.WriteFile("report.pdf", data, 0644)
}
```

### Programmatic Document

```go
func main() {
    gen, _ := pdf.New(cfg)

    doc := pdf.NewDocument().
        Title("Invoice #12345").
        PageSize(pdf.A4).
        Margins(20, 15, 20, 15).
        Header(`<div style="text-align:right">ACME Corp</div>`, 15).
        Footer(`<div style="text-align:center">Page {{page}} of {{pages}}</div>`, 10).
        Add(
            pdf.Image("logo.png", pdf.ImageWidth(100)),
            pdf.Heading(1, "Invoice #12345"),
            pdf.Spacer(10),
            pdf.Text("Date: January 15, 2024"),
            pdf.Text("Due: January 30, 2024"),
            pdf.Spacer(20),
            pdf.Table(
                []string{"Item", "Qty", "Price", "Total"},
                [][]string{
                    {"Widget A", "10", "$10.00", "$100.00"},
                    {"Widget B", "5", "$20.00", "$100.00"},
                    {"Service Fee", "1", "$50.00", "$50.00"},
                },
                pdf.TableBordered(),
                pdf.TableStriped(),
            ),
            pdf.Spacer(10),
            pdf.Text("Total: $250.00", pdf.Bold(), pdf.AlignRight()),
        ).
        Build()

    data, _ := gen.Generate(ctx, doc)
    os.WriteFile("invoice.pdf", data, 0644)
}
```

### Template-Based

```go
//go:embed templates/invoice.html
var invoiceTemplate string

type InvoiceData struct {
    Number    string
    Date      string
    DueDate   string
    Customer  Customer
    Items     []LineItem
    Total     float64
}

func main() {
    gen, _ := pdf.New(cfg)

    data := InvoiceData{
        Number:  "INV-12345",
        Date:    "2024-01-15",
        DueDate: "2024-01-30",
        Items: []LineItem{
            {Name: "Widget A", Qty: 10, Price: 10.00},
        },
        Total: 100.00,
    }

    pdfData, _ := gen.GenerateFromTemplate(ctx, invoiceTemplate, data)
    os.WriteFile("invoice.pdf", pdfData, 0644)
}
```

## Dependencies

- **Required:** None (native engine uses pure Go)
- **Optional:**
  - `wkhtmltopdf` binary for wkhtml engine
  - `github.com/chromedp/chromedp` for Chrome engine
  - `github.com/go-pdf/fpdf` for native engine

## Implementation Phases

### Phase 1: Core Interface & wkhtmltopdf
1. Define Generator interface
2. HTML-to-PDF with wkhtmltopdf
3. Basic options (page size, margins)

### Phase 2: Document Builder
1. Programmatic document API
2. Text, heading, image elements
3. Table component

### Phase 3: Additional Engines
1. Chrome/Puppeteer engine
2. Native Go engine (fpdf)

### Phase 4: Templates
1. Template rendering
2. Header/footer support
3. Page numbers

### Phase 5: Documentation
1. README
2. Invoice example
3. Report example
