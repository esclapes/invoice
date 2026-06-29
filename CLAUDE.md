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
│   ├── invoice-add             # DEFAULT: record a new invoice in the manifest + generate it
│   ├── invoice-generate        # One-off PDF (preview), not recorded in any manifest
│   └── invoice-regenerate      # Re-render existing manifest entries
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

### Manifest is the source of truth
- The manifest (`invoices/YYYY/QN/manifest.yaml`) is **authoritative** for accountancy and regeneration. The PDF is a render of it, never the other way around.
- **`invoice-add` is the default entry point**: it appends a new entry to the manifest (auto-assigning the next ID), then renders the PDF via `invoice-regenerate`. The assigned `id` is written into the manifest and **frozen** — regeneration never recomputes it, so adding invoices out of order later can't shift earlier IDs.
- `invoice-generate` is the low-level renderer + a one-off "preview" escape hatch. It does **not** touch the manifest, so anything created with it directly is unrecorded.

### Numbering (scripts/invoice-add)
- Numbers are **per-series, per-year** (not per-quarter): `MA-2026-001` (Q1) is followed by `MA-2026-002` (Q2). `invoice-add` finds the next number by scanning all of that year's quarter manifests for the series and taking `max(number)+1`.
- Builds the manifest entry as JSON (jq, for safe escaping) and appends it with `yq`, then forces block style on the new entry so the file stays human-readable.

### Filing by invoice date, not wall-clock (scripts/invoice-generate)
- Year/quarter for the output path are derived from `--date` (parsed via BSD `date -j -f "%b %d, %Y"`), falling back to today only if `--date` is absent/unparseable. This fixes a former bug where generating/regenerating filed PDFs under the *current* quarter instead of the invoice's quarter.
- Uses **bash arrays** to build the command (not string concat + eval), preserving quoting for spaces/special characters.
- Loads company/client config from JSON via jq. Output filename: `{ID}.pdf`.

### Manifest System (invoices/YYYY/QN/manifest.yaml)
- YAML file storing all parameters for each invoice in a quarter.
- Written by `invoice-add`; consumed by `invoice-regenerate`.
- Gitignored alongside PDFs (contains financial data).
- Requires `yq` for YAML parsing (`brew install yq`).

## Workflows

### Create a new invoice (DEFAULT — use this)
Record it in the manifest and render it in one step. The ID is auto-assigned and frozen.
```bash
./scripts/invoice-add <series> <client> [flags]
# series : normal | ma
# client : key from clients.json
# flags  : --item/-i, --quantity/-q, --rate/-r (repeatable, paired by order)
#          --date "May 12, 2026"  (default: today; also decides the YYYY/QN folder)
#          --due  "May 12, 2026"  (default: same as --date)

# Example — Mundo (midwife / ma series), 3h at €45:
./scripts/invoice-add ma mundo -i "Assistant work" -q 3 -r 45 --date "May 12, 2026"
```
Do NOT hand-pick the number — `invoice-add` assigns the next one for the series/year.

### Edit or re-render an existing invoice
The manifest is the source of truth. To change an invoice (amounts, dates, items), edit its entry in the quarter's `manifest.yaml`, then re-render:
```bash
./scripts/invoice-regenerate invoices/2026/Q1/manifest.yaml              # all in the file
./scripts/invoice-regenerate invoices/2026/Q1/manifest.yaml 2026-005     # one
./scripts/invoice-regenerate invoices/2026/Q1/manifest.yaml MA-2026-001  # one
```

### One-off preview (NOT recorded)
For a throwaway PDF you don't want in the books — e.g. checking layout. This does not write the manifest, so don't use it for real invoices:
```bash
./scripts/invoice-generate <series> <number> <client> [flags]
```

### Build after code changes
```bash
go build -o invoice . && go install .
```
The script calls `invoice` from PATH, so `go install` is needed after code changes.

## Gotchas

- **Use `invoice-add`, not `invoice-generate`, for real invoices.** `invoice-generate` produces an unrecorded PDF; only `invoice-add` writes the manifest. An invoice that isn't in a manifest effectively doesn't exist for accounting/regeneration.
- **Don't choose invoice numbers by hand.** `invoice-add` auto-increments per series/year. Manually filing into the wrong quarter or skipping a number breaks the `max(number)+1` assumption.
- **Numbers are per-year, not per-quarter.** The next number is computed across all of the year's manifests for that series.
- **Commas in items are safe** because of `StringArrayVarP`. Each `--item` flag is a separate item.
- **Multiline descriptions** use literal `\n` in the string (e.g., `--item "Service\nDate1\nDate2"`). The Go code converts `\n` to newlines (`pdf.go`). This survives the jq→yq→regenerate round-trip.
- **company.json `tax_note`** supports `\\n` for multiline notes (escaped for JSON).
- **Dependencies:** `jq` (JSON) and `yq` (YAML) are both required by `invoice-add`; `invoice-regenerate` needs `yq`; `invoice-generate` needs `jq`. `brew install jq yq`.
- **Upstream conflicts** are likely in `pdf.go` and `main.go` if syncing from upstream.

## Future Improvements

- **Maroto library** (github.com/johnfercher/maroto): Bootstrap-like 12-column grid for Go PDFs. Would simplify layout code significantly if a rewrite is warranted.
- **HTML/CSS → PDF** via chromedp: Full CSS flexbox/grid support, but adds Chrome dependency.
- **Interactive wizard** (possible follow-up): a prompt-driven front-end over `invoice-add` (pick client, enter line items, confirm) for fewer flags to remember. Deliberately deferred — `invoice-add` covers the core need.
- **Storage stays flat-file (decided, not a TODO):** YAML manifests beat a local SQLite DB at this scale — human-/git-readable, greppable, auditable, zero schema/driver overhead. Revisit a DB only if volume grows into the hundreds and reporting queries ("total billed to client X in Q2") become routine; migration from the manifests would be trivial then.
- Per-client notes or payment terms (e.g. `default_items` / due-date offset in `clients.json`)
- Email integration for sending invoices

## License

MIT License — same as upstream.
