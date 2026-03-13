package main

import (
	"fmt"
	"os"

	"github.com/xuri/excelize/v2"
)

func main() {

	os.MkdirAll("templates_excel", 0755)

	generateExportContractsTemplate()
	generateBalanceSheetTemplate()
	generateFinancialResultsTemplate()
	generateCreditAgreementsTemplate()

	fmt.Println("✅ Все шаблоны успешно сгенерированы в папке 'templates_excel/'")
}

func generateExportContractsTemplate() {

	f := excelize.NewFile()
	sheet := "Экспортные контракты"

	index, _ := f.NewSheet(sheet)
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	headers := []string{
		"Номер контракта", "Дата контракта", "Страна импортёра", "Код ТН ВЭД",
		"Объём (т)", "Цена ($/т)", "Валюта", "Срок оплаты (дней)",
		"Дата отгрузки", "Курс на дату контракта",
	}

	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, header)
		f.SetCellStyle(sheet, cell, cell, getHeaderStyle(f))
	}

	exampleData := [][]interface{}{
		{"KC-2026-001", "01.02.2026", "Китай", "3104", 350000, 285.00, "USD", 60, "15.02.2026", 3.25},
		{"KC-2026-002", "05.02.2026", "Бразилия", "3104", 220000, 282.00, "USD", 45, "20.02.2026", 3.26},
	}

	for r, row := range exampleData {
		for c, val := range row {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+2)
			f.SetCellValue(sheet, cell, val)
		}
	}

	// ---- страна импортёра
	start, _ := excelize.CoordinatesToCellName(3, 2)
	end, _ := excelize.CoordinatesToCellName(3, 1000)

	dv := excelize.NewDataValidation(true)
	dv.Sqref = start + ":" + end
	dv.SetDropList([]string{
		"Китай", "Бразилия", "Индия", "США",
		"Германия", "Польша", "Турция",
		"Египет", "Вьетнам", "Другие",
	})

	f.AddDataValidation(sheet, dv)

	// ---- валюта
	start, _ = excelize.CoordinatesToCellName(7, 2)
	end, _ = excelize.CoordinatesToCellName(7, 1000)

	dv2 := excelize.NewDataValidation(true)
	dv2.Sqref = start + ":" + end
	dv2.SetDropList([]string{"USD", "EUR", "CNY", "BYN"})

	f.AddDataValidation(sheet, dv2)

	f.SaveAs("templates_excel/export_contracts_template.xlsx")
}

func generateBalanceSheetTemplate() {

	f := excelize.NewFile()
	sheet := "Финансовый баланс"

	index, _ := f.NewSheet(sheet)
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	headers := []string{
		"Дата отчёта", "ДС в BYN", "ДС в USD", "КФВ", "Дебиторка", "НДС к получению",
		"Запасы", "Сырьё", "НЗП", "ГП", "ОС", "НМА", "Кредиторка", "НДС к уплате",
		"ЗП", "Краткосрочные кредиты", "Долгосрочные кредиты", "УК", "НП", "Комментарий",
	}

	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, header)
		f.SetCellStyle(sheet, cell, cell, getHeaderStyle(f))
	}

	f.SaveAs("templates_excel/balance_sheet_template.xlsx")
}

func generateFinancialResultsTemplate() {

	f := excelize.NewFile()
	sheet := "Отчёт о ФР"

	index, _ := f.NewSheet(sheet)
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	headers := []string{
		"Дата отчёта", "Период с", "Период по", "Выручка", "Экспорт", "Внутр.", "Прочие",
		"Итого выручка", "С/С", "Сырьё", "Энергия", "ЗП", "Амортизация", "Прочие затраты",
		"Итого С/С", "Комрассходы", "Упррасходы", "Прочие расходы", "Валовая прибыль",
		"Оперприбыль", "Прибыль до налога", "Налог", "Чистая прибыль", "EBITDA",
		"Опермаржа %", "Чистая маржа %",
	}

	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, header)
		f.SetCellStyle(sheet, cell, cell, getHeaderStyle(f))
	}

	f.SaveAs("templates_excel/financial_results_template.xlsx")
}

func generateCreditAgreementsTemplate() {

	f := excelize.NewFile()
	sheet := "Кредитные договоры"

	index, _ := f.NewSheet(sheet)
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	headers := []string{
		"Номер договора", "Дата заключения", "Банк", "Страна банка", "Сумма", "Валюта",
		"Ставка %", "Тип ставки", "Дата начала", "Дата погашения",
		"Срок (мес.)", "Тип обеспечения", "Статус",
	}

	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, header)
		f.SetCellStyle(sheet, cell, cell, getHeaderStyle(f))
	}

	// ---- валюта
	start, _ := excelize.CoordinatesToCellName(6, 2)
	end, _ := excelize.CoordinatesToCellName(6, 1000)

	dv := excelize.NewDataValidation(true)
	dv.Sqref = start + ":" + end
	dv.SetDropList([]string{"USD", "EUR", "BYN"})

	f.AddDataValidation(sheet, dv)

	f.SaveAs("templates_excel/credit_agreements_template.xlsx")
}

func getHeaderStyle(f *excelize.File) int {

	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Color:  "FFFFFF",
			Family: "Arial",
			Size:   10,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"4472C4"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
	})

	return style
}