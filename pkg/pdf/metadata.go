package pdf

import (
	"fmt"
	"strings"

	"github.com/invopop/gobl"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/note"
	"github.com/invopop/gobl/org"
)

const (
	defaultCreator = "GOBL HTML"
)

// MetadataFromEnvelope extracts PDF metadata from a GOBL envelope.
// It uses the document type to determine appropriate title and subject,
// extracts the author from the supplier/issuer, and builds relevant keywords.
func MetadataFromEnvelope(env *gobl.Envelope) *Metadata {
	if env == nil {
		return nil
	}

	md := &Metadata{
		Creator: defaultCreator,
	}

	doc := env.Extract()
	switch d := doc.(type) {
	case *bill.Invoice:
		md.Title = invoiceTitle(d)
		md.Subject = invoiceSubject(d)
		md.Author = partyName(d.Supplier)
		md.Keywords = invoiceKeywords(d)
	case *bill.Payment:
		md.Title = paymentTitle(d)
		md.Subject = "Payment"
		md.Author = partyName(d.Supplier)
		md.Keywords = paymentKeywords(d)
	case *bill.Delivery:
		md.Title = deliveryTitle(d)
		md.Subject = "Delivery Note"
		md.Author = partyName(d.Supplier)
		md.Keywords = deliveryKeywords(d)
	case *bill.Order:
		md.Title = orderTitle(d)
		md.Subject = orderSubject(d)
		md.Author = partyName(d.Seller)
		md.Keywords = orderKeywords(d)
	case *note.Message:
		md.Title = messageTitle(d)
		md.Subject = "Message"
	case *org.Party:
		md.Title = d.Name
		md.Subject = "Party"
	default:
		md.Title = "GOBL Document"
		md.Subject = "Document"
	}

	return md
}

func invoiceTitle(inv *bill.Invoice) string {
	code := inv.Series.Join(inv.Code)
	if code == "" {
		return "Invoice"
	}
	return fmt.Sprintf("Invoice %s", code)
}

func invoiceSubject(inv *bill.Invoice) string {
	switch inv.Type {
	case bill.InvoiceTypeStandard:
		return "Invoice"
	case bill.InvoiceTypeCreditNote:
		return "Credit Note"
	case bill.InvoiceTypeDebitNote:
		return "Debit Note"
	case bill.InvoiceTypeCorrective:
		return "Corrective Invoice"
	case bill.InvoiceTypeProforma:
		return "Proforma Invoice"
	default:
		return "Invoice"
	}
}

func invoiceKeywords(inv *bill.Invoice) string {
	keywords := make([]string, 0)

	// Add document type
	keywords = append(keywords, string(inv.Type))

	// Add currency
	if inv.Currency != "" {
		keywords = append(keywords, string(inv.Currency))
	}

	// Add tax regime
	if !inv.Regime.IsEmpty() {
		keywords = append(keywords, string(inv.Regime.GetRegime()))
	}

	// Add supplier tax ID if available
	if inv.Supplier != nil && len(inv.Supplier.Identities) > 0 {
		for _, id := range inv.Supplier.Identities {
			if id.Code != "" {
				keywords = append(keywords, id.Code.String())
				break
			}
		}
	}
	if inv.Supplier != nil && inv.Supplier.TaxID != nil && inv.Supplier.TaxID.Code != "" {
		keywords = append(keywords, inv.Supplier.TaxID.Code.String())
	}

	return strings.Join(keywords, ", ")
}

func paymentTitle(p *bill.Payment) string {
	code := p.Series.Join(p.Code)
	if code == "" {
		return "Payment"
	}
	return fmt.Sprintf("Payment %s", code)
}

func paymentKeywords(p *bill.Payment) string {
	keywords := make([]string, 0)

	if p.Currency != "" {
		keywords = append(keywords, string(p.Currency))
	}

	if !p.Regime.IsEmpty() {
		keywords = append(keywords, string(p.Regime.GetRegime()))
	}

	return strings.Join(keywords, ", ")
}

func deliveryTitle(d *bill.Delivery) string {
	code := d.Series.Join(d.Code)
	if code == "" {
		return "Delivery Note"
	}
	return fmt.Sprintf("Delivery Note %s", code)
}

func deliveryKeywords(d *bill.Delivery) string {
	keywords := make([]string, 0)

	if !d.Regime.IsEmpty() {
		keywords = append(keywords, string(d.Regime.GetRegime()))
	}

	return strings.Join(keywords, ", ")
}

func orderTitle(o *bill.Order) string {
	code := o.Series.Join(o.Code)
	if code == "" {
		return "Order"
	}
	return fmt.Sprintf("Order %s", code)
}

func orderSubject(o *bill.Order) string {
	switch o.Type {
	case bill.OrderTypePurchase:
		return "Purchase Order"
	case bill.OrderTypeSale:
		return "Sales Order"
	case bill.OrderTypeQuote:
		return "Quote"
	default:
		return "Order"
	}
}

func orderKeywords(o *bill.Order) string {
	keywords := make([]string, 0)

	if o.Type != "" {
		keywords = append(keywords, string(o.Type))
	}

	if o.Currency != "" {
		keywords = append(keywords, string(o.Currency))
	}

	if !o.Regime.IsEmpty() {
		keywords = append(keywords, string(o.Regime.GetRegime()))
	}

	return strings.Join(keywords, ", ")
}

func messageTitle(m *note.Message) string {
	if m.Title != "" {
		return m.Title
	}
	return "Message"
}

func partyName(p *org.Party) string {
	if p == nil {
		return ""
	}
	return p.Name
}
