package main

type HTML struct {
	Status string
	SfNum  string
	SfDate string
	LkID   string

	OrgPrintName                 string
	OrgPrintAddress              string
	OrgInnKpp                    string
	ShipperPrintNameAddress      string
	ConsigneePrintNameAddress    string
	PaymentAndSettlementDocument string
	ShippingDocuments            string

	SupplierPrintName    string
	SupplierPrintAddress string
	SupplierInnKpp       string
	CurrencyNameCode     string

	AmountWithoutVatTotal string
	AmountVatTotal        string
	AmountWithVatTotal    string

	OrgChiefName      string
	OrgAccountantName string

	Contract string

	StoreKeeperPosition string
	StoreKeeperName     string
	DocSendDate         string
	SenderChiefPosition string
	SenderChiefName     string

	RecipientPosition      string
	RecipientName          string
	RecipientChiefPosition string
	RecipientChiefName     string
	DocReceiveDate         string

	OrgSignImage string

	Details []HTMLDetail
}

type HTMLDetail struct {
	Number           string
	Code             string
	Title            string
	KindID           string
	OkeiID           string
	OkeiCode         string
	Quantity         string
	Price            string
	AmountWithoutVat string
	Excise           string
	Vat              string
	AmountVat        string
	AmountWithVat    string
	CountryID        string
	CountryName      string
	Gtd              string
}
