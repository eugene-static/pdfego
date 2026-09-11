package tests

import (
	"slices"
	"time"
)

type View struct {
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

	Certificate Certificate

	Details        []Detail
	DetailsOrdered []string
}

type Detail struct {
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

type Certificate struct {
	SignerName         string
	SignDate           string
	KEP                string
	ValidityPeriodFrom string
	ValidityPeriodTo   string
}

func newView(detailsNum int) View {
	details := []Detail{
		{
			Number:           "1",
			Code:             "2001915",
			Title:            "Бумага офисная Комус Документ Standard+ А4 80 г/кв.м марка С 146 CIE (500 листов)",
			KindID:           "--",
			OkeiID:           "796",
			OkeiCode:         "шт.",
			Quantity:         "100",
			Price:            "300,00",
			AmountWithoutVat: "30 000,00",
			Excise:           "без акциза",
			Vat:              "22%",
			AmountVat:        "6 600,00",
			AmountWithVat:    "36 600,00",
			CountryID:        "643",
			CountryName:      "Россия",
			Gtd:              "10013160/161122/3554104/10",
		},
	}

	details = slices.Repeat(details, detailsNum)

	cert := Certificate{
		SignerName:         "Michael Gary Scott",
		SignDate:           time.Now().Format("02.01.2006, 15:04"),
		KEP:                "90379e6a254d4df79c93",
		ValidityPeriodFrom: "01.01.2025 09:00",
		ValidityPeriodTo:   "01.01.2028 09:00",
	}

	upd := View{
		Status:                       "1",
		SfNum:                        "1234567890",
		SfDate:                       time.Now().Format("02.01.2006"),
		LkID:                         "00000000001",
		OrgPrintName:                 "Dunder Mifflin Paper Company",
		OrgPrintAddress:              "1725 Slough Avenue, Scranton, PA 18505, USA",
		OrgInnKpp:                    "1020304050/1121314151",
		ShipperPrintNameAddress:      "он же",
		ConsigneePrintNameAddress:    "Valve Corporation, 10400 NE 4th Street, Suite 1400 Bellevue, WA 98004, USA",
		PaymentAndSettlementDocument: "",
		ShippingDocuments:            "УПД № 777 от 06.06.2026г.",
		SupplierPrintName:            "Valve Corporation",
		SupplierPrintAddress:         "10400 NE 4th Street, Suite 1400 Bellevue, WA 98004, USA",
		SupplierInnKpp:               "1222324252/1323334353",
		CurrencyNameCode:             "Американский доллар, 840",
		AmountWithoutVatTotal:        "30 000,00",
		AmountVatTotal:               "6 600,00",
		AmountWithVatTotal:           "36 600,00",
		OrgChiefName:                 "Michael Gary Scott",
		OrgAccountantName:            "Kevin Jamel Malone",
		Contract:                     "Договор оферты №777 от 04.04.2024 г.",
		StoreKeeperPosition:          "Начальник склада",
		StoreKeeperName:              "Darryl Philbin",
		DocSendDate:                  time.Now().Format("02.01.2006"),
		SenderChiefPosition:          "Региональный менеджер",
		SenderChiefName:              "Michael Gary Scott",
		RecipientPosition:            "Менеджер по закупкам и логистике",
		RecipientName:                "Jeliy Fisher",
		RecipientChiefPosition:       "Генеральный директор",
		RecipientChiefName:           "Gabe Newell",
		DocReceiveDate:               "",
		Certificate:                  cert,
		Details:                      details,
		DetailsOrdered:               make([]string, 16),
	}

	return upd
}
