package tests

import (
	"slices"
	"time"
)

type DTO struct {
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

func NewDTO(detailsNum int) DTO {
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
		{
			Number:           "2",
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
		{
			Number:           "3",
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

	upd := DTO{
		Status:                       "1",
		SfNum:                        "1234567890",
		SfDate:                       time.Now().Format("02.01.2006"),
		LkID:                         "00000000001",
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

func (dto DTO) Section(index int) []string {
	dto.DetailsOrdered[0] = dto.Details[index].Code
	dto.DetailsOrdered[1] = dto.Details[index].Number
	dto.DetailsOrdered[2] = dto.Details[index].Title
	dto.DetailsOrdered[3] = dto.Details[index].KindID
	dto.DetailsOrdered[4] = dto.Details[index].OkeiID
	dto.DetailsOrdered[5] = dto.Details[index].OkeiCode
	dto.DetailsOrdered[6] = dto.Details[index].Quantity
	dto.DetailsOrdered[7] = dto.Details[index].Price
	dto.DetailsOrdered[8] = dto.Details[index].AmountWithoutVat
	dto.DetailsOrdered[9] = dto.Details[index].Excise
	dto.DetailsOrdered[10] = dto.Details[index].Vat
	dto.DetailsOrdered[11] = dto.Details[index].AmountVat
	dto.DetailsOrdered[12] = dto.Details[index].AmountWithVat
	dto.DetailsOrdered[13] = dto.Details[index].CountryID
	dto.DetailsOrdered[14] = dto.Details[index].CountryName
	dto.DetailsOrdered[15] = dto.Details[index].Gtd

	return dto.DetailsOrdered
}

func (dto DTO) Count() int {
	return len(dto.Details)
}
