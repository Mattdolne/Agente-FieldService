package exporter

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"win-diag-agent/pkg/collector"
	"win-diag-agent/pkg/console"
	"win-diag-agent/pkg/model"

	"github.com/jung-kurt/gofpdf"
)

const totalBoxWidth = 98
const contentWidth = 94 // 98 - 4 (para "║ " e " ║")

// pad ajusta o tamanho da string considerando caracteres Unicode reais (runes)
func pad(s string, width int) string {
	runes := []rune(s)
	if len(runes) >= width {
		return string(runes[:width])
	}
	return s + strings.Repeat(" ", width-len(runes))
}

// wrapText quebra um texto longo respeitando o limite de largura e palavras
func wrapText(text string, maxWidth int) []string {
	var lines []string
	paragraphs := strings.Split(text, "\n")

	for _, p := range paragraphs {
		words := strings.Fields(p)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}

		currentLine := ""
		currentLen := 0

		for _, w := range words {
			wLen := len([]rune(w))
			if currentLen == 0 {
				currentLine = w
				currentLen = wLen
			} else if currentLen+1+wLen <= maxWidth {
				currentLine += " " + w
				currentLen += 1 + wLen
			} else {
				lines = append(lines, currentLine)
				currentLine = w
				currentLen = wLen
			}
		}
		if currentLen > 0 {
			lines = append(lines, currentLine)
		}
	}
	return lines
}

// cardLine formata uma linha única dentro da caixa fechada
func cardLine(text string) string {
	return fmt.Sprintf("║ %s ║\n", pad(text, contentWidth))
}

// cardParagraph formata parágrafos com quebra de linha e recuo francês elegante
func cardParagraph(firstPrefix string, text string) string {
	var sb strings.Builder
	firstIndentLen := len([]rune(firstPrefix))
	subsequentPrefix := strings.Repeat(" ", firstIndentLen)
	availWidth := contentWidth - firstIndentLen
	lines := wrapText(text, availWidth)

	for i, l := range lines {
		p := firstPrefix
		if i > 0 {
			p = subsequentPrefix
		}
		sb.WriteString(fmt.Sprintf("║ %s%s ║\n", p, pad(l, availWidth)))
	}
	return sb.String()
}

// cardEmptyLine gera uma linha em branco com bordas fechadas
func cardEmptyLine() string {
	return fmt.Sprintf("║%s║\n", strings.Repeat(" ", totalBoxWidth-2))
}

// FormatConsoleReport renderiza o relatório com caixas 100% fechadas e alinhadas
func FormatConsoleReport(info model.SystemInfo, rawReboots string, issues []model.DiagnosticIssue) string {
	var sb strings.Builder

	// Limpa o prefixo 'Microsoft ' para caber com elegância no formulário
	winVer := strings.TrimPrefix(info.WindowsVersion, "Microsoft ")

	// 1. FORMULÁRIO DO COMPUTADOR (CABEÇALHO)
	h := fmt.Sprintf(`
┌──────────────────────────────────────────────────────────────────────────────────────────────────┐
│                               AGENTE FIELDSERVICE - RELATÓRIO DO SISTEMA                         │
├────────────────────────────────────────┬─────────────────────────────────────────────────────────┤
│ DADOS DO COMPUTADOR                    │ ESPECIFICAÇÕES DE HARDWARE                              │
├────────────────────────────────────────┼─────────────────────────────────────────────────────────┤
│ Hostname:       %s │ Processador:    %s │
│ Domínio:        %s │ Memória RAM:    %s │
│ Tag / Serial:   %s │ Armazenamento:  %s │
│ Versão SO:      %s │ Versão BIOS:    %s │
└────────────────────────────────────────┴─────────────────────────────────────────────────────────┘
`,
		pad(info.Hostname, 22), pad(info.Processor, 39),
		pad(info.Domain, 22), pad(info.RAMUsed+" / "+info.RAMTotal, 39),
		pad(info.SerialNumber, 22), pad(info.Storage, 39),
		pad(winVer, 22), pad(info.BIOSVersion, 39),
	)

	sb.WriteString(h)
	console.Print(h)

	// 2. TABELA DOS ÚLTIMOS REBOOTS
	rebootsHeader := `
┌──────────────────────────────────────────────────────────────────────────────────────────────────┐
│                     HISTÓRICO RECENTE DE INICIALIZAÇÕES E EVENTOS DE ENERGIA                     │
├─────────────────────┬─────────────────┬──────────────────────────────────────────────────────────┤
│ DATA E HORA         │ EVENTO          │ CLASSIFICAÇÃO / MOTIVO                                   │
├─────────────────────┼─────────────────┼──────────────────────────────────────────────────────────┤
`
	sb.WriteString(rebootsHeader)
	console.Print(rebootsHeader)

	rebootRows := collector.FormatRebootsTable(rawReboots)
	if len(rebootRows) == 0 {
		row := "│ --/--/---- --:--:-- │ Sem Eventos     │ Nenhum evento anômalo registrado                         │\n"
		sb.WriteString(row)
		console.Print(row)
	} else {
		for _, l := range rebootRows {
			parts := strings.Split(l, "|")
			dt := "--/--/---- --:--:--"
			ev := "Evento"
			desc := l
			if len(parts) >= 3 {
				dt = strings.TrimSpace(parts[0])
				ev = strings.TrimSpace(parts[1])
				desc = strings.TrimSpace(parts[2])
			}
			row := fmt.Sprintf("│ %s │ %s │ %s │\n", pad(dt, 19), pad(ev, 15), pad(desc, 56))
			sb.WriteString(row)
			console.Print(row)
		}
	}
	tableEnd := "└─────────────────────┴─────────────────┴──────────────────────────────────────────────────────────┘\n"
	sb.WriteString(tableEnd)
	console.Print(tableEnd)

	// 3. FICHAS TÉCNICAS DE DIAGNÓSTICO (CARDS 100% FECHADOS)
	console.Println("\nDIAGNÓSTICO DETALHADO E PLANO DE AÇÃO:")
	sb.WriteString("\nDIAGNÓSTICO DETALHADO E PLANO DE AÇÃO:\n")

	if len(issues) == 0 {
		okCard := `
┌──────────────────────────────────────────────────────────────────────────────────────────────────┐
│ [STATUS: SAUDÁVEL] Nenhum problema de hardware ou erro crítico detectado no sistema!            │
└──────────────────────────────────────────────────────────────────────────────────────────────────┘
`
		sb.WriteString(okCard)
		console.Print(okCard)
	}

	for i, issue := range issues {
		var card strings.Builder

		// Topo e Cabeçalho do Card
		topBorder := fmt.Sprintf("╔%s╗\n", strings.Repeat("═", totalBoxWidth-2))
		midBorder := fmt.Sprintf("╠%s╣\n", strings.Repeat("═", totalBoxWidth-2))
		botBorder := fmt.Sprintf("╚%s╝\n", strings.Repeat("═", totalBoxWidth-2))

		titleText := fmt.Sprintf("[%s] #%d: %s", issue.Severity, i+1, issue.Title)

		card.WriteString(topBorder)
		card.WriteString(cardLine(titleText))
		card.WriteString(midBorder)
		card.WriteString(cardEmptyLine())

		// Origem
		card.WriteString(cardLine("► ONDE FOI ENCONTRADO:"))
		card.WriteString(cardParagraph("   ", issue.Source))
		card.WriteString(cardEmptyLine())

		// Evidências
		card.WriteString(cardLine("► EVIDÊNCIAS COLETADAS:"))
		evLines := strings.Split(issue.LogsEvidence, "\n")
		for _, ev := range evLines {
			ev = strings.TrimSpace(ev)
			if ev != "" {
				card.WriteString(cardParagraph("   • ", ev))
			}
		}
		card.WriteString(cardEmptyLine())

		// Causa Raiz / Interpretação
		card.WriteString(cardLine("► INTERPRETAÇÃO TÉCNICA (CAUSA RAIZ):"))
		card.WriteString(cardParagraph("   ", issue.Interpretation))
		card.WriteString(cardEmptyLine())

		// Remediação
		card.WriteString(cardLine("► TAREFAS RECOMENDADAS PARA RESOLUÇÃO:"))
		for j, task := range issue.Remediation {
			prefix := fmt.Sprintf("   [%d] ", j+1)
			card.WriteString(cardParagraph(prefix, task))
		}
		card.WriteString(cardEmptyLine())
		card.WriteString(botBorder)

		fullCard := card.String()
		sb.WriteString(fullCard)
		console.Print(fullCard)
	}

	return sb.String()
}

// ExportTXT salva o relatório em formato texto puro
func ExportTXT(filename string, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}

// ExportCSV salva os problemas diagnosticados em tabela CSV
func ExportCSV(filename string, info model.SystemInfo, issues []model.DiagnosticIssue) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.Comma = ';'
	defer writer.Flush()

	writer.Write([]string{"Hostname", "Severidade", "Categoria", "Titulo", "Onde_Encontrado", "Logs_Evidencias", "Interpretacao_Tecnica", "Tarefas_Remediacao"})

	if len(issues) == 0 {
		writer.Write([]string{info.Hostname, "OK", "SISTEMA", "Nenhum problema detectado", "Log de Sistema", "Sem falhas recentes", "Sistema operando dentro dos parâmetros normais", "Nenhuma ação necessária"})
		return nil
	}

	for _, issue := range issues {
		remediationJoined := strings.Join(issue.Remediation, " | ")
		cleanEvidence := strings.ReplaceAll(issue.LogsEvidence, "\n", " ; ")
		writer.Write([]string{
			info.Hostname,
			issue.Severity,
			issue.Category,
			issue.Title,
			issue.Source,
			cleanEvidence,
			issue.Interpretation,
			remediationJoined,
		})
	}
	return nil
}

// ExportPDF gera um documento PDF formatado e profissional
func ExportPDF(filename string, info model.SystemInfo, rawReboots string, issues []model.DiagnosticIssue) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	pdf.SetFont("Arial", "B", 16)
	pdf.SetFillColor(33, 50, 75)
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(180, 12, tr("AGENTE FIELDSERVICE - RELATÓRIO DO SISTEMA"), "0", 1, "C", true, 0, "")
	pdf.Ln(4)

	// Informações Gerais (Tabela no PDF)
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(240, 242, 245)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(180, 7, tr("  INFORMAÇÕES DO DISPOSITIVO"), "1", 1, "L", true, 0, "")

	winVer := strings.TrimPrefix(info.WindowsVersion, "Microsoft ")
	sysFields := [][]string{
		{"Hostname:", info.Hostname, "Domínio:", info.Domain},
		{"Tag / Serial:", info.SerialNumber, "Versão BIOS:", info.BIOSVersion},
		{"Versão SO:", winVer, "Processador:", info.Processor},
		{"Memória RAM:", info.RAMUsed + " / " + info.RAMTotal, "Armazenamento:", info.Storage},
	}

	for _, row := range sysFields {
		pdf.SetFont("Arial", "B", 8)
		pdf.CellFormat(28, 6, tr(row[0]), "1", 0, "L", false, 0, "")
		pdf.SetFont("Arial", "", 8)
		pdf.CellFormat(62, 6, tr(row[1]), "1", 0, "L", false, 0, "")
		pdf.SetFont("Arial", "B", 8)
		pdf.CellFormat(28, 6, tr(row[2]), "1", 0, "L", false, 0, "")
		pdf.SetFont("Arial", "", 8)
		pdf.CellFormat(62, 6, tr(row[3]), "1", 1, "L", false, 0, "")
	}
	pdf.Ln(5)

	// Histórico de Reboots no PDF
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(240, 242, 245)
	pdf.CellFormat(180, 7, tr("  HISTÓRICO RECENTE DE INICIALIZAÇÕES E ENERGIA"), "1", 1, "L", true, 0, "")
	pdf.SetFont("Arial", "", 8)

	rebootRows := collector.FormatRebootsTable(rawReboots)
	if len(rebootRows) > 0 {
		for _, l := range rebootRows {
			pdf.MultiCell(180, 5, tr("• "+l), "LR", "L", false)
		}
	} else {
		pdf.CellFormat(180, 6, tr("Nenhum evento registrado."), "LR", 1, "L", false, 0, "")
	}
	pdf.CellFormat(180, 1, "", "T", 1, "", false, 0, "")
	pdf.Ln(5)

	// Fichas no PDF
	for i, issue := range issues {
		switch issue.Severity {
		case "CRÍTICO":
			pdf.SetFillColor(217, 83, 79)
		case "ALTO":
			pdf.SetFillColor(240, 173, 78)
		default:
			pdf.SetFillColor(91, 192, 222)
		}
		pdf.SetTextColor(255, 255, 255)
		pdf.SetFont("Arial", "B", 9)
		pdf.CellFormat(180, 6, tr(fmt.Sprintf(" [%s] #%d: %s", issue.Severity, i+1, issue.Title)), "1", 1, "L", true, 0, "")

		pdf.SetTextColor(0, 0, 0)
		pdf.SetFont("Arial", "B", 8)
		pdf.CellFormat(180, 5, tr("Origem: ")+tr(issue.Source), "LR", 1, "L", false, 0, "")
		pdf.SetFont("Arial", "", 8)
		pdf.MultiCell(180, 4, tr("Evidências:\n")+tr(issue.LogsEvidence), "LR", "L", false)
		pdf.MultiCell(180, 4, tr("Interpretação:\n")+tr(issue.Interpretation), "LR", "L", false)

		pdf.SetFont("Arial", "B", 8)
		pdf.CellFormat(180, 5, tr("Tarefas de Resolução:"), "LR", 1, "L", false, 0, "")
		pdf.SetFont("Arial", "", 8)
		for _, task := range issue.Remediation {
			pdf.MultiCell(180, 4, tr("  -> ")+tr(task), "LR", "L", false)
		}
		pdf.CellFormat(180, 1, "", "T", 1, "", false, 0, "")
		pdf.Ln(3)
	}

	return pdf.OutputFileAndClose(filename)
}
