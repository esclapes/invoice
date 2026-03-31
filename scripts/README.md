# Invoice Scripts

Scripts for generating and managing invoices.

## Setup

1. Copy example configs:
   ```bash
   cp config/company.example.json config/company.json
   cp config/clients.example.json config/clients.json
   ```

2. Fill in your details in `config/company.json` and `config/clients.json`

3. Build and install the invoice binary:
   ```bash
   go build -o invoice . && go install .
   ```

4. Install dependencies:
   ```bash
   brew install jq yq
   ```

## invoice-generate

Generate a single invoice.

```bash
invoice-generate <series> <number> [client] [options]
```

### Series Types
- `normal` - Standard series (YYYY-NNN format, e.g., 2026-001)
- `ma` - MA series (MA-YYYY-NNN format, e.g., MA-2026-001)

### Examples

```bash
# Simple invoice with configured client
invoice-generate normal 1 buffer --item "Software Development" --quantity 1 --rate 5000

# Multiline items with dates
invoice-generate ma 1 mundo \
  --item "On-call service\n13-02-2026\n19-02-2026" \
  --quantity 4 --rate 700

# Custom date and due date
invoice-generate normal 5 buffer \
  --item "Consulting" --quantity 1 --rate 3000 \
  --date "Feb 12, 2026" --due "Feb 12, 2026"

# Custom client (not in config)
invoice-generate normal 3 \
  --to "Custom Client\nAddress\nCity, Country\nTax ID" \
  --item "Development" --quantity 10 --rate 150
```

## invoice-regenerate

Regenerate invoices from a manifest.yaml file.

```bash
# Regenerate all invoices in a quarter
invoice-regenerate invoices/2026/Q1/manifest.yaml

# Regenerate a single invoice
invoice-regenerate invoices/2026/Q1/manifest.yaml 2026-005
```

Requires `yq` for YAML parsing.

## Output

Invoices are organized by year and quarter:
```
invoices/
└── 2026/
    └── Q1/
        ├── 2026-001.pdf
        ├── MA-2026-001.pdf
        └── manifest.yaml
```

## Configuration

- `config/company.json` - Business details, tax rate, notes (supports `\n` for multiline)
- `config/clients.json` - Client database with shorthand names

Both are gitignored. See `config/*.example.json` for structure.
