package main

import (
	"fmt"
	"os"
	"slices"
	"time"
)

func main() {
	details := []HTMLDetail{
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

	details = slices.Repeat(details, 50000)

	upd := HTML{
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
	}

	t := time.Now()

	printForm, err := upd.FillTemplate()
	if err != nil {
		fmt.Println(err)

		return
	}

	fmt.Printf("duration: %s", time.Since(t).String())

	output, err := os.Create("output.pdf")
	if err != nil {
		fmt.Println(err)

		return
	}

	defer output.Close()

	_, err = output.Write(printForm)
	if err != nil {
		fmt.Println(err)

		return
	}
}
