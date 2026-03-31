# Invoice Tool - Custom Fork

Customized fork of [maaslalani/invoice](https://github.com/maaslalani/invoice) for generating PDF invoices from the command line.

## Project Structure

```
invoice/
├── config/                      # Personal configuration (gitignored)
│   ├── company.json            # Business details, tax rate, notes
│   ├── clients.json            # Client database with shorthand names
│   ├── company.example.json    # Template for company config
│   └── clients.example.json    # Template for clients config
├── scripts/
│   ├── invoice-generate        # Main generation script
│   └── invoice-regenerate      # Regenerate from manifest.yaml
├── invoices/                    # Generated invoices (gitignored)
│   └── YYYY/QN/
│       ├── ID.pdf              # Generated PDFs
│       └── manifest.yaml       # Invoice records for regeneration
├── main.go                      # CLI entrypoint, flag definitions
├── pdf.go                       # PDF layout and rendering
├── import.go                    # Config file import
└── currency.go                  # Currency symbols
```

## Key Architecture Decisions

### PDF Layout (pdf.go)
- **Two-column header**: `writeHeader()` renders invoice title + client on the left, company info right-aligned on the right. Replaced the original stacked `writeLogo`/`writeTitle`/`writeBillTo`.
- **Compact spacing**: Font sizes and line heights are tightened to fit dense invoices (5+ multiline items) on a single page.
- **Totals pinned to bottom**: `writeTotals()` calculates a target Y position so the totals/due/notes/footer section anchors near the bottom of the page. Falls back to normal flow when content is long.
- **Render order**: items → totals → due date → notes → footer (notes come after totals).
- **Multi-page support**: If content exceeds ~700pt Y, a new page is added automatically.
- Uses gopdf (github.com/signintech/gopdf) — manual X/Y positioning, no CSS-like layout. See "Future Improvements" for alternatives.

### CLI (main.go)
- Items use `StringArrayVarP` (not `StringSliceVarP`) to prevent commas in item descriptions from being split into separate items.
- `--date` and `--due` flags accept format like "Jan 14, 2026".
- Tax rate is passed through to `writeTotals` for display as percentage.

### Generation Script (scripts/invoice-generate)
- Uses **bash arrays** to build the command, not string concatenation + eval. This preserves quoting for arguments with spaces and special characters.
- Loads company/client config from JSON via jq.
- Auto-determines year and quarter for output path.
- Output filename: `{ID}.pdf` (e.g., `2026-001.pdf`, `MA-2026-001.pdf`).

### Manifest System (invoices/YYYY/QN/manifest.yaml)
- YAML file storing all parameters for each invoice in a quarter.
- Used by `scripts/invoice-regenerate` to regenerate one or all invoices.
- Gitignored alongside PDFs (contains financial data).
- Requires `yq` for YAML parsing (`brew install yq`).

## Workflows

### Generate a single invoice
```bash
./scripts/invoice-generate <series> <number> <client> [flags]
# series: normal | ma
# number: invoice number (auto zero-padded to 3 digits)
# client: key from clients.json, or omit and pass --to directly
```

### Regenerate from manifest
```bash
./scripts/invoice-regenerate invoices/2026/Q1/manifest.yaml              # all
./scripts/invoice-regenerate invoices/2026/Q1/manifest.yaml 2026-005     # one
./scripts/invoice-regenerate invoices/2026/Q1/manifest.yaml MA-2026-001  # one
```

### Build after code changes
```bash
go build -o invoice . && go install .
```
The script calls `invoice` from PATH, so `go install` is needed after code changes.

## Gotchas

- **Commas in items are safe** because of `StringArrayVarP`. Each `--item` flag is a separate item.
- **Multiline descriptions** use literal `\n` in the string (e.g., `--item "Service\nDate1\nDate2"`). The Go code converts `\n` to newlines.
- **company.json `tax_note`** supports `\\n` for multiline notes (escaped for JSON).
- **The script requires jq** for JSON parsing and **invoice-regenerate requires yq** for YAML parsing.
- **Upstream conflicts** are likely in `pdf.go` and `main.go` if syncing from upstream.

## Future Improvements

- **Maroto library** (github.com/johnfercher/maroto): Bootstrap-like 12-column grid for Go PDFs. Would simplify layout code significantly if a rewrite is warranted.
- **HTML/CSS → PDF** via chromedp: Full CSS flexbox/grid support, but adds Chrome dependency.
- Per-client notes or payment terms
- Email integration for sending invoices

## License

MIT License — same as upstream.
