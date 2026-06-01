package upd

import (
	"slices"
	"time"
)

type UPD struct {
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

func NewUPD(detailsNum int) UPD {
	details := []Detail{
		{
			Number:           "1",
			Code:             "1732",
			Title:            "РЗШ.Штраф: поставка № 26532296 объемом 5000 штук, запланированная на 2025-02-10, была привезена в объеме 180 штук. РЗШ.Штраф: поставка № 26532296 объемом 5000 штук, запланированная на 2025-02-10, была привезена в объеме 180 штук.",
			KindID:           "",
			OkeiID:           "--",
			OkeiCode:         "--",
			Quantity:         "1",
			Price:            "100,00",
			AmountWithoutVat: "100,00",
			Excise:           "без акциза",
			Vat:              "22%",
			AmountVat:        "22,00",
			AmountWithVat:    "122,00",
			CountryID:        "--",
			CountryName:      "--",
			Gtd:              "--",
		},
	}

	details = slices.Repeat(details, detailsNum)

	upd := UPD{
		Status:                       "1",
		SfNum:                        "1234567890",
		SfDate:                       time.Now().Format("02.01.2006"),
		LkID:                         "LkID",
		OrgPrintName:                 "ОБЩЕСТВО С ОГРАНИЧЕННОЙ ОТВЕТСТВЕННОСТЬЮ \"РВБ\"",
		OrgPrintAddress:              "142181, Московская обл, г.о. Подольск, д Коледино, тер. Индустриальный парк Коледино, д. 6, стр. 1",
		OrgInnKpp:                    "9714053621/507401001",
		ShipperPrintNameAddress:      "---",
		ConsigneePrintNameAddress:    "---",
		PaymentAndSettlementDocument: "",
		ShippingDocuments:            "Универсальный передаточный документ № 225656479 от 26.02.2025 г.",
		SupplierPrintName:            "ИП Матвеев Андрей Юрьевич",
		SupplierPrintAddress:         "115088, г. Москва, ул. Новоостаповская, д.8, кв. 65",
		SupplierInnKpp:               "212408738537/",
		CurrencyNameCode:             "Российский рубль, 643",
		AmountWithoutVatTotal:        "300,00",
		AmountVatTotal:               "66,00",
		AmountWithVatTotal:           "366,00",
		OrgChiefName:                 "Мирзоян Роберт Георгиевич",
		OrgAccountantName:            "Сарыкова Наталья Викторовна",
		Contract:                     "Договор оферты №б/н от 05.08.2024 г.",
		StoreKeeperPosition:          "Заместитель главного бухгалтера",
		StoreKeeperName:              "Сарыкова Наталья Викторовна",
		DocSendDate:                  time.Now().Format("02.01.2006"),
		SenderChiefPosition:          "Заместитель главного бухгалтера",
		SenderChiefName:              "Сарыкова Наталья Викторовна",
		RecipientPosition:            "",
		RecipientName:                "",
		RecipientChiefPosition:       "",
		RecipientChiefName:           "",
		DocReceiveDate:               "",
		OrgSignImage:                 "/1691194.png",
		Details:                      details,
		DetailsOrdered:               make([]string, 16),
	}

	return upd
}

func (upd UPD) OrderedRow(index int) []string {
	upd.DetailsOrdered[0] = upd.Details[index].Code
	upd.DetailsOrdered[1] = upd.Details[index].Number
	upd.DetailsOrdered[2] = upd.Details[index].Title
	upd.DetailsOrdered[3] = upd.Details[index].KindID
	upd.DetailsOrdered[4] = upd.Details[index].OkeiID
	upd.DetailsOrdered[5] = upd.Details[index].OkeiCode
	upd.DetailsOrdered[6] = upd.Details[index].Quantity
	upd.DetailsOrdered[7] = upd.Details[index].Price
	upd.DetailsOrdered[8] = upd.Details[index].AmountWithoutVat
	upd.DetailsOrdered[9] = upd.Details[index].Excise
	upd.DetailsOrdered[10] = upd.Details[index].Vat
	upd.DetailsOrdered[11] = upd.Details[index].AmountVat
	upd.DetailsOrdered[12] = upd.Details[index].AmountWithVat
	upd.DetailsOrdered[13] = upd.Details[index].CountryID
	upd.DetailsOrdered[14] = upd.Details[index].CountryName
	upd.DetailsOrdered[15] = upd.Details[index].Gtd

	return upd.DetailsOrdered
}

func (upd UPD) Len() int {
	return len(upd.Details)
}
