package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"win-diag-agent/pkg/collector"
	"win-diag-agent/pkg/console"
	"win-diag-agent/pkg/exporter"
)

func main() {
	console.Println("================================================================================")
	console.Println("             AGENTE FIELDSERVICE - DIAGNÓSTICO DO SISTEMA                       ")
	console.Println("================================================================================")
	console.Println("[*] Inicializando varredura detalhada da máquina...\r\n")

	console.Print("[1/4] Identificando processador, memória, discos e BIOS... ")
	info := collector.CollectSystemInfo()
	console.Println("CONCLUÍDO")

	console.Print("[2/4] Lendo histórico das últimas 5 reinicializações e energia... ")
	raw := collector.CollectRawDiagnostics()
	console.Println("CONCLUÍDO")

	console.Print("[3/4] Inspecionando falhas de DLL, atualizações e despejos de memória... ")
	console.Println("CONCLUÍDO")

	console.Print("[4/4] Analisando integridade física (WHEA, SMART, limites térmicos)... ")
	issues := collector.AnalyzeDiagnostics(raw)
	console.Println("CONCLUÍDO\r\n")

	// Renderiza o novo relatório em tabelas, formulários e fichas técnicas
	reportContent := exporter.FormatConsoleReport(info, raw.RebootsRaw, issues)

	// Menu interativo de exportação
	reader := bufio.NewReader(os.Stdin)
	console.Println("\nDeseja exportar este relatório?")
	console.Println("  [1] Salvar em TXT (.txt)")
	console.Println("  [2] Salvar em CSV (.csv - dados tabulares para planilhas)")
	console.Println("  [3] Salvar em PDF (.pdf - documento formal formatado)")
	console.Println("  [4] Exportar em TODOS os formatos (TXT, CSV e PDF)")
	console.Println("  [0] Sair sem salvar")
	console.Print("\nEscolha uma opção [0-4]: ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	timestamp := time.Now().Format("20060102_150405")
	baseName := fmt.Sprintf("Diagnostico_%s_%s", info.Hostname, timestamp)

	switch choice {
	case "1":
		txtFile := baseName + ".txt"
		if err := exporter.ExportTXT(txtFile, reportContent); err == nil {
			abs, _ := filepath.Abs(txtFile)
			console.Printf("\n[OK] Relatório TXT gerado com sucesso em:\n    %s\n", abs)
		}
	case "2":
		csvFile := baseName + ".csv"
		if err := exporter.ExportCSV(csvFile, info, issues); err == nil {
			abs, _ := filepath.Abs(csvFile)
			console.Printf("\n[OK] Relatório CSV gerado com sucesso em:\n    %s\n", abs)
		}
	case "3":
		pdfFile := baseName + ".pdf"
		if err := exporter.ExportPDF(pdfFile, info, raw.RebootsRaw, issues); err == nil {
			abs, _ := filepath.Abs(pdfFile)
			console.Printf("\n[OK] Relatório PDF gerado com sucesso em:\n    %s\n", abs)
		}
	case "4":
		txtFile := baseName + ".txt"
		csvFile := baseName + ".csv"
		pdfFile := baseName + ".pdf"
		exporter.ExportTXT(txtFile, reportContent)
		exporter.ExportCSV(csvFile, info, issues)
		exporter.ExportPDF(pdfFile, info, raw.RebootsRaw, issues)
		abs, _ := filepath.Abs(baseName)
		console.Printf("\n[OK] Todos os arquivos gerados com sucesso:\n    %s.*\n", abs)
	default:
		console.Println("\nNenhum arquivo de exportação foi gerado.")
	}

	console.Println("\nPressione Enter para encerrar...")
	reader.ReadString('\n')
}
