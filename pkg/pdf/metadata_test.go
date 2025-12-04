package pdf

import (
	"testing"

	"github.com/invopop/gobl"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/currency"
	"github.com/invopop/gobl/note"
	"github.com/invopop/gobl/org"
	"github.com/invopop/gobl/tax"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetadataFromEnvelope_Invoice(t *testing.T) {
	inv := &bill.Invoice{
		Series:   "INV",
		Code:     "001",
		Type:     bill.InvoiceTypeStandard,
		Currency: currency.EUR,
		Supplier: &org.Party{
			Name: "Test Supplier",
		},
	}
	env, err := gobl.Envelop(inv)
	require.NoError(t, err)

	md := MetadataFromEnvelope(env)

	assert.NotNil(t, md)
	assert.Equal(t, "Invoice INV-001", md.Title)
	assert.Equal(t, "Invoice", md.Subject)
	assert.Equal(t, "Test Supplier", md.Author)
	assert.Equal(t, "GOBL HTML", md.Creator)
	assert.Contains(t, md.Keywords, "EUR")
}

func TestMetadataFromEnvelope_CreditNote(t *testing.T) {
	inv := &bill.Invoice{
		Series:   "CN",
		Code:     "100",
		Type:     bill.InvoiceTypeCreditNote,
		Currency: currency.USD,
		Supplier: &org.Party{
			Name: "Acme Corp",
		},
	}
	env, err := gobl.Envelop(inv)
	require.NoError(t, err)

	md := MetadataFromEnvelope(env)

	assert.NotNil(t, md)
	assert.Equal(t, "Invoice CN-100", md.Title)
	assert.Equal(t, "Credit Note", md.Subject)
	assert.Equal(t, "Acme Corp", md.Author)
}

func TestMetadataFromEnvelope_InvoiceWithRegime(t *testing.T) {
	inv := &bill.Invoice{
		Regime:   tax.WithRegime("ES"),
		Series:   "F",
		Code:     "2024-001",
		Type:     bill.InvoiceTypeStandard,
		Currency: currency.EUR,
		Supplier: &org.Party{
			Name: "Spanish Company",
			TaxID: &tax.Identity{
				Country: "ES",
				Code:    "B12345678",
			},
		},
	}
	env, err := gobl.Envelop(inv)
	require.NoError(t, err)

	md := MetadataFromEnvelope(env)

	assert.NotNil(t, md)
	assert.Equal(t, "Invoice F-2024-001", md.Title)
	assert.Contains(t, md.Keywords, "ES")
	assert.Contains(t, md.Keywords, "B12345678")
}

func TestMetadataFromEnvelope_Payment(t *testing.T) {
	payment := &bill.Payment{
		Series:   "PAY",
		Code:     "001",
		Currency: currency.GBP,
		Supplier: &org.Party{
			Name: "Payment Provider",
		},
	}
	env, err := gobl.Envelop(payment)
	require.NoError(t, err)

	md := MetadataFromEnvelope(env)

	assert.NotNil(t, md)
	assert.Equal(t, "Payment PAY-001", md.Title)
	assert.Equal(t, "Payment", md.Subject)
	assert.Equal(t, "Payment Provider", md.Author)
	assert.Contains(t, md.Keywords, "GBP")
}

func TestMetadataFromEnvelope_Delivery(t *testing.T) {
	delivery := &bill.Delivery{
		Series:   "DEL",
		Code:     "2024-100",
		Currency: currency.EUR,
		Supplier: &org.Party{
			Name: "Logistics Company",
		},
	}
	env, err := gobl.Envelop(delivery)
	require.NoError(t, err)

	md := MetadataFromEnvelope(env)

	assert.NotNil(t, md)
	assert.Equal(t, "Delivery Note DEL-2024-100", md.Title)
	assert.Equal(t, "Delivery Note", md.Subject)
	assert.Equal(t, "Logistics Company", md.Author)
}

func TestMetadataFromEnvelope_Order(t *testing.T) {
	order := &bill.Order{
		Series:   "PO",
		Code:     "001",
		Type:     bill.OrderTypePurchase,
		Currency: currency.EUR,
		Seller: &org.Party{
			Name: "Seller Company",
		},
	}
	env, err := gobl.Envelop(order)
	require.NoError(t, err)

	md := MetadataFromEnvelope(env)

	assert.NotNil(t, md)
	assert.Equal(t, "Order PO-001", md.Title)
	assert.Equal(t, "Purchase Order", md.Subject)
	assert.Equal(t, "Seller Company", md.Author)
}

func TestMetadataFromEnvelope_Message(t *testing.T) {
	msg := &note.Message{
		Title:   "Important Notice",
		Content: "This is a test message",
	}
	env, err := gobl.Envelop(msg)
	require.NoError(t, err)

	md := MetadataFromEnvelope(env)

	assert.NotNil(t, md)
	assert.Equal(t, "Important Notice", md.Title)
	assert.Equal(t, "Message", md.Subject)
}

func TestMetadataFromEnvelope_MessageWithoutTitle(t *testing.T) {
	msg := &note.Message{
		Content: "This is a test message",
	}
	env, err := gobl.Envelop(msg)
	require.NoError(t, err)

	md := MetadataFromEnvelope(env)

	assert.NotNil(t, md)
	assert.Equal(t, "Message", md.Title)
	assert.Equal(t, "Message", md.Subject)
}

func TestMetadataFromEnvelope_Party(t *testing.T) {
	party := &org.Party{
		Name: "Test Organization",
	}
	env, err := gobl.Envelop(party)
	require.NoError(t, err)

	md := MetadataFromEnvelope(env)

	assert.NotNil(t, md)
	assert.Equal(t, "Test Organization", md.Title)
	assert.Equal(t, "Party", md.Subject)
}

func TestMetadataFromEnvelope_Nil(t *testing.T) {
	md := MetadataFromEnvelope(nil)
	assert.Nil(t, md)
}

func TestMetadataFromEnvelope_InvoiceWithoutCode(t *testing.T) {
	inv := &bill.Invoice{
		Type:     bill.InvoiceTypeStandard,
		Currency: currency.EUR,
		Supplier: &org.Party{
			Name: "No Code Supplier",
		},
	}
	env, err := gobl.Envelop(inv)
	require.NoError(t, err)

	md := MetadataFromEnvelope(env)

	assert.NotNil(t, md)
	assert.Equal(t, "Invoice", md.Title)
}

func TestMetadataFromEnvelope_InvoiceCodeOnly(t *testing.T) {
	inv := &bill.Invoice{
		Code:     cbc.Code("001"),
		Type:     bill.InvoiceTypeStandard,
		Currency: currency.EUR,
	}
	env, err := gobl.Envelop(inv)
	require.NoError(t, err)

	md := MetadataFromEnvelope(env)

	assert.NotNil(t, md)
	assert.Equal(t, "Invoice 001", md.Title)
}
