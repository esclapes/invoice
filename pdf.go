package main

import (
	"fmt"
	"image"
	"os"
	"strconv"
	"strings"

	"github.com/signintech/gopdf"
)

const (
	quantityColumnOffset = 360
	rateColumnOffset     = 405
	amountColumnOffset   = 480
)

const (
	subtotalLabel = "Subtotal"
	discountLabel = "Discount"
	taxLabel      = "Tax"
	totalLabel    = "Total"
)

func writeHeader(pdf *gopdf.GoPdf, logo string, from string, title string, id string, date string, to string) {
	const leftMargin = 40.0
	const rightMargin = 555.0
	startY := pdf.GetY()

	// === RIGHT COLUMN: Company info (right-aligned) ===
	rightY := startY

	if logo != "" {
		width, height := getImageDimension(logo)
		scaledWidth := 80.0
		scaledHeight := float64(height) * scaledWidth / float64(width)
		_ = pdf.Image(logo, rightMargin-scaledWidth, rightY, &gopdf.Rect{W: scaledWidth, H: scaledHeight})
		rightY += scaledHeight + 12
	}

	formattedFrom := strings.ReplaceAll(from, `\n`, "\n")
	fromLines := strings.Split(formattedFrom, "\n")

	pdf.SetTextColor(55, 55, 55)
	for i, line := range fromLines {
		if i == 0 {
			_ = pdf.SetFont("Inter", "", 11)
		} else {
			_ = pdf.SetFont("Inter", "", 9)
		}
		w, _ := pdf.MeasureTextWidth(line)
		pdf.SetX(rightMargin - w)
		pdf.SetY(rightY)
		_ = pdf.Cell(nil, line)
		if i == 0 {
			rightY += 16
		} else {
			rightY += 13
		}
	}

	// === LEFT COLUMN: Title + ID + Bill To ===
	leftY := startY

	// Title
	pdf.SetX(leftMargin)
	pdf.SetY(leftY)
	_ = pdf.SetFont("Inter-Bold", "", 18)
	pdf.SetTextColor(0, 0, 0)
	_ = pdf.Cell(nil, title)
	leftY += 24

	// ID and date
	pdf.SetX(leftMargin)
	pdf.SetY(leftY)
	_ = pdf.SetFont("Inter", "", 11)
	pdf.SetTextColor(100, 100, 100)
	_ = pdf.Cell(nil, "#"+id)
	pdf.SetTextColor(150, 150, 150)
	_ = pdf.Cell(nil, "  ·  ")
	pdf.SetTextColor(100, 100, 100)
	_ = pdf.Cell(nil, date)
	leftY += 24

	// Bill To
	pdf.SetX(leftMargin)
	pdf.SetY(leftY)
	pdf.SetTextColor(75, 75, 75)
	_ = pdf.SetFont("Inter", "", 9)
	_ = pdf.Cell(nil, "CUSTOMER")
	leftY += 16

	formattedTo := strings.ReplaceAll(to, `\n`, "\n")
	toLines := strings.Split(formattedTo, "\n")

	for i, line := range toLines {
		pdf.SetX(leftMargin)
		pdf.SetY(leftY)
		if i == 0 {
			_ = pdf.SetFont("Inter", "", 13)
			_ = pdf.Cell(nil, line)
			leftY += 18
		} else {
			_ = pdf.SetFont("Inter", "", 9)
			_ = pdf.Cell(nil, line)
			leftY += 13
		}
	}

	// Set Y to max of both columns
	maxY := leftY
	if rightY > maxY {
		maxY = rightY
	}
	pdf.SetY(maxY)
	pdf.SetX(leftMargin)

	// Separator line
	pdf.Br(8)
	pdf.SetStrokeColor(225, 225, 225)
	pdf.Line(leftMargin, pdf.GetY(), rightMargin, pdf.GetY())
	pdf.Br(16)
}

func writeDueDate(pdf *gopdf.GoPdf, due string) {
	_ = pdf.SetFont("Inter", "", 9)
	pdf.SetTextColor(75, 75, 75)
	_ = pdf.Cell(nil, "Due Date")
	_ = pdf.SetFont("Inter-Bold", "", 9)
	pdf.SetTextColor(0, 0, 0)
	_ = pdf.Cell(nil, "  "+due)
	pdf.Br(12)
}

func writeHeaderRow(pdf *gopdf.GoPdf) {
	_ = pdf.SetFont("Inter", "", 9)
	pdf.SetTextColor(55, 55, 55)
	_ = pdf.Cell(nil, "ITEM")
	pdf.SetX(quantityColumnOffset)
	_ = pdf.Cell(nil, "QTY")
	pdf.SetX(rateColumnOffset)
	_ = pdf.Cell(nil, "RATE")
	pdf.SetX(amountColumnOffset)
	_ = pdf.Cell(nil, "AMOUNT")
	pdf.Br(18)
}

func writeNotes(pdf *gopdf.GoPdf, notes string) {
	pdf.Br(16)

	_ = pdf.SetFont("Inter", "", 8)
	pdf.SetTextColor(55, 55, 75)
	_ = pdf.Cell(nil, "NOTES")
	pdf.Br(14)
	_ = pdf.SetFont("Inter", "", 8)
	pdf.SetTextColor(0, 0, 0)

	formattedNotes := strings.ReplaceAll(notes, `\n`, "\n")
	notesLines := strings.Split(formattedNotes, "\n")

	for i := 0; i < len(notesLines); i++ {
		_ = pdf.Cell(nil, notesLines[i])
		pdf.Br(12)
	}

	pdf.Br(16)
}
func writeFooter(pdf *gopdf.GoPdf, id string) {
	// Position footer at bottom of current page
	currentY := pdf.GetY()
	// If we're past 750, we're near the bottom already
	if currentY < 750 {
		pdf.SetY(800)
	} else {
		pdf.Br(24) // Just add some space
	}

	_ = pdf.SetFont("Inter", "", 10)
	pdf.SetTextColor(55, 55, 55)
	_ = pdf.Cell(nil, id)
	pdf.SetStrokeColor(225, 225, 225)
	pdf.Line(pdf.GetX()+10, pdf.GetY()+6, 550, pdf.GetY()+6)
	pdf.Br(48)
}

func writeRow(pdf *gopdf.GoPdf, item string, quantity int, rate float64) {
	_ = pdf.SetFont("Inter", "", 10)
	pdf.SetTextColor(0, 0, 0)

	total := float64(quantity) * rate
	amount := strconv.FormatFloat(total, 'f', 2, 64)

	// Support multiline items like from/to/notes
	formattedItem := strings.ReplaceAll(item, `\n`, "\n")
	itemLines := strings.Split(formattedItem, "\n")

	// Write first line with all columns
	_ = pdf.Cell(nil, itemLines[0])
	pdf.SetX(quantityColumnOffset)
	_ = pdf.Cell(nil, strconv.Itoa(quantity))
	pdf.SetX(rateColumnOffset)
	_ = pdf.Cell(nil, currencySymbols[file.Currency]+strconv.FormatFloat(rate, 'f', 2, 64))
	pdf.SetX(amountColumnOffset)
	_ = pdf.Cell(nil, currencySymbols[file.Currency]+amount)
	pdf.Br(15)

	// Write additional lines (item description details) with smaller font
	if len(itemLines) > 1 {
		_ = pdf.SetFont("Inter", "", 8)
		pdf.SetTextColor(100, 100, 100)
		for i := 1; i < len(itemLines); i++ {
			_ = pdf.Cell(nil, itemLines[i])
			pdf.Br(12)
		}
		pdf.SetTextColor(0, 0, 0)
		pdf.Br(4) // Extra spacing after multiline item
	}
}

func writeTotals(pdf *gopdf.GoPdf, subtotal float64, tax float64, discount float64, taxRate float64) {
	currentY := pdf.GetY()

	// Estimate bottom section height: totals + due + notes + footer
	// Totals: 3×18=54, Due: 12, Notes: ~70, Footer gap: ~30 = ~166
	bottomSectionHeight := 166.0
	if discount > 0 {
		bottomSectionHeight += 18
	}
	targetY := 800 - bottomSectionHeight

	if currentY > 700 {
		pdf.AddPage()
		pdf.SetY(100)
	} else if currentY < targetY {
		// Pin totals to the bottom of the page
		pdf.SetY(targetY)
	} else {
		pdf.Br(16)
	}

	writeTotal(pdf, subtotalLabel, subtotal)
	// Always show tax with percentage, even if 0%
	writeTotalWithRate(pdf, taxLabel, tax, taxRate)
	if discount > 0 {
		writeTotal(pdf, discountLabel, discount)
	}
	writeTotal(pdf, totalLabel, subtotal+tax-discount)
}

func writeTotal(pdf *gopdf.GoPdf, label string, total float64) {
	_ = pdf.SetFont("Inter", "", 9)
	pdf.SetTextColor(75, 75, 75)
	pdf.SetX(rateColumnOffset)
	_ = pdf.Cell(nil, label)
	pdf.SetTextColor(0, 0, 0)
	_ = pdf.SetFontSize(12)
	pdf.SetX(amountColumnOffset - 15)
	if label == totalLabel {
		_ = pdf.SetFont("Inter-Bold", "", 11.5)
	}
	_ = pdf.Cell(nil, currencySymbols[file.Currency]+strconv.FormatFloat(total, 'f', 2, 64))
	pdf.Br(18)
}

func writeTotalWithRate(pdf *gopdf.GoPdf, label string, total float64, rate float64) {
	_ = pdf.SetFont("Inter", "", 9)
	pdf.SetTextColor(75, 75, 75)
	pdf.SetX(rateColumnOffset)
	// Show label with percentage
	labelWithRate := fmt.Sprintf("%s (%.0f%%)", label, rate*100)
	_ = pdf.Cell(nil, labelWithRate)
	pdf.SetTextColor(0, 0, 0)
	_ = pdf.SetFontSize(12)
	pdf.SetX(amountColumnOffset - 15)
	_ = pdf.Cell(nil, currencySymbols[file.Currency]+strconv.FormatFloat(total, 'f', 2, 64))
	pdf.Br(18)
}

func getImageDimension(imagePath string) (int, int) {
	file, err := os.Open(imagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
	defer file.Close()

	image, _, err := image.DecodeConfig(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", imagePath, err)
	}
	return image.Width, image.Height
}
