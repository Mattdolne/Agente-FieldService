package collector

import (
	"fmt"
	"strings"
	"win-diag-agent/pkg/analyzer"
	"win-diag-agent/pkg/model"
)

// RawDiagnostics armazena as evidências brutas coletadas
type RawDiagnostics struct {
	RebootsRaw      string
	KernelPowerRaw  string
	MinidumpsRaw    string
	BugCheckRaw     string
	WHEARaw         string
	DiskErrorsRaw   string
	SmartHealthRaw  string
	ThermalRaw      string
	DllErrorsRaw    string
	UpdateErrorsRaw string
	SecureBootRaw   string
	DefenderRaw     string
	UACRaw          string
}

// CollectRawDiagnostics executa as varreduras retornando dados estruturados em ASCII/chaves
func CollectRawDiagnostics() RawDiagnostics {
	raw := RawDiagnostics{}

	// 1. Últimos 5 reboots / desligamentos (Retorna: DataHora;EventID;Motivo)
	raw.RebootsRaw = RunPS(`
		Get-WinEvent -FilterHashtable @{LogName='System'; Id=6005,6006,6008,41,1074} -MaxEvents 15 -ErrorAction SilentlyContinue |
		Select-Object -First 5 |
		ForEach-Object {
			$m = ""
			if ($_.Id -eq 1074 -and $_.Properties -and $_.Properties.Count -gt 4) {
				$m = $_.Properties[4].Value
			}
			$_.TimeCreated.ToString("dd/MM/yyyy HH:mm:ss") + ";" + $_.Id + ";" + $m
		}
	`)

	// 2. Incidentes de Kernel-Power (Event ID 41) específicos
	raw.KernelPowerRaw = RunPS(`
		Get-WinEvent -FilterHashtable @{LogName='System'; Id=41} -MaxEvents 5 -ErrorAction SilentlyContinue |
		ForEach-Object { $_.TimeCreated.ToString("dd/MM/yyyy HH:mm:ss") + ";41;KERNEL_POWER" }
	`)

	// 3. Arquivos de Minidump (C:\Windows\Minidump)
	raw.MinidumpsRaw = RunPS(`
		Get-ChildItem -Path "C:\Windows\Minidump" -Filter "*.dmp" -ErrorAction SilentlyContinue |
		Sort-Object LastWriteTime -Descending |
		Select-Object -First 5 |
		ForEach-Object { $_.LastWriteTime.ToString("dd/MM/yyyy HH:mm:ss") + " | " + $_.Name + " (" + [math]::Round($_.Length/1KB,0) + " KB)" }
	`)

	// 4. Detalhes de BugCheck no log System (Event ID 1001)
	raw.BugCheckRaw = RunPS(`
		Get-WinEvent -FilterHashtable @{LogName='System'; ProviderName='Microsoft-Windows-WER-SystemErrorReporting'} -MaxEvents 3 -ErrorAction SilentlyContinue |
		ForEach-Object { $_.TimeCreated.ToString("dd/MM/yyyy HH:mm:ss") + " | " + ($_.Message.Split([char]10)[0].Trim()) }
	`)

	// 5. WHEA-Logger (Event IDs 17, 18, 19, 47)
	raw.WHEARaw = RunPS(`
		Get-WinEvent -FilterHashtable @{LogName='System'; ProviderName='Microsoft-Windows-WHEA-Logger'} -MaxEvents 5 -ErrorAction SilentlyContinue |
		ForEach-Object { $_.TimeCreated.ToString("dd/MM/yyyy HH:mm:ss") + ";" + $_.Id }
	`)

	// 6. Erros de Disco / Controladora (Event IDs 7, 11, 153, 129, 55, 98)
	raw.DiskErrorsRaw = RunPS(`
		Get-WinEvent -FilterHashtable @{LogName='System'; Id=7,11,153,129,55,98} -MaxEvents 5 -ErrorAction SilentlyContinue |
		ForEach-Object { $_.TimeCreated.ToString("dd/MM/yyyy HH:mm:ss") + ";" + $_.Id }
	`)

	// 7. Status SMART
	raw.SmartHealthRaw = RunPS(`
		Get-PhysicalDisk -ErrorAction SilentlyContinue |
		ForEach-Object { $_.FriendlyName + " | Status: " + $_.OperationalStatus + " | Saude: " + $_.HealthStatus }
	`)

	// 8. Desempenho Térmico (Event ID 37)
	raw.ThermalRaw = RunPS(`
		Get-WinEvent -FilterHashtable @{LogName='System'; ProviderName='Microsoft-Windows-Kernel-Processor-Power'; Id=37} -MaxEvents 3 -ErrorAction SilentlyContinue |
		ForEach-Object { $_.TimeCreated.ToString("dd/MM/yyyy HH:mm:ss") + ";37;THROTTLING" }
	`)

	// 9. Erros de DLL / Aplicação (Event ID 1000)
	raw.DllErrorsRaw = RunPS(`
		Get-WinEvent -FilterHashtable @{LogName='Application'; Id=1000} -MaxEvents 5 -ErrorAction SilentlyContinue |
		ForEach-Object {
			$msg = $_.Message.Split([char]10)
			$app = ($msg | Where-Object { $_ -match "Nome do aplicativo com falha|Faulting application name" }) -replace ".*:\s*", ""
			$mod = ($msg | Where-Object { $_ -match "Nome do módulo com falha|Faulting module name" }) -replace ".*:\s*", ""
			$code = ($msg | Where-Object { $_ -match "Código de exceção|Exception code" }) -replace ".*:\s*", ""
			$_.TimeCreated.ToString("dd/MM/yyyy HH:mm:ss") + ";" + $app.Trim() + ";" + $mod.Trim() + ";" + $code.Trim()
		}
	`)

	// 10. Falhas do Windows Update (Event ID 20 e 25)
	raw.UpdateErrorsRaw = RunPS(`
		Get-WinEvent -FilterHashtable @{LogName='Microsoft-Windows-WindowsUpdateClient/Operational'; Id=20,25} -MaxEvents 4 -ErrorAction SilentlyContinue |
		ForEach-Object { $_.TimeCreated.ToString("dd/MM/yyyy HH:mm:ss") + ";" + $_.Id }
	`)

	// 11. Secure Boot
	raw.SecureBootRaw = RunPS(`
		$sb = Confirm-SecureBootUEFI -ErrorAction SilentlyContinue
		if ($sb -eq $true) { "Ativado" } elseif ($sb -eq $false) { "Desativado" } else { "Legacy" }
	`)

	// 12. Windows Defender
	raw.DefenderRaw = RunPS(`
		$d = Get-MpComputerStatus -ErrorAction SilentlyContinue
		if ($d) {
			"RealTimeProtection=" + $d.RealTimeProtectionEnabled + "; Antivirus=" + $d.AntivirusEnabled
		} else { "N/D" }
	`)

	// 13. UAC
	raw.UACRaw = RunPS(`
		(Get-ItemProperty "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System" -Name EnableLUA -ErrorAction SilentlyContinue).EnableLUA
	`)

	return raw
}

// FormatRebootsTable traduz com segurança os reboots em Go para exibição formatada
func FormatRebootsTable(rawReboots string) []string {
	if rawReboots == "" || rawReboots == "N/D" {
		return nil
	}

	var formatted []string
	lines := strings.Split(rawReboots, "\n")
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		parts := strings.Split(l, ";")
		if len(parts) >= 2 {
			dt := parts[0]
			id := parts[1]
			motivo := ""
			if len(parts) >= 3 {
				motivo = parts[2]
			}

			desc := "Evento de Energia"
			switch id {
			case "6005":
				desc = "Inicialização do Sistema (Boot limpo)"
			case "6006":
				desc = "Desligamento Normal (Clean Shutdown)"
			case "6008":
				desc = "Desligamento Inesperado (Queda de energia/Crash)"
			case "41":
				desc = "Kernel-Power (Reinício abrupto sem desligamento)"
			case "1074":
				if motivo != "" {
					desc = "Reinicialização Solicitada (" + motivo + ")"
				} else {
					desc = "Reinicialização Solicitada (Usuário/Sistema)"
				}
			}
			formatted = append(formatted, fmt.Sprintf("%s | Event ID %s | %s", dt, id, desc))
		} else {
			formatted = append(formatted, l)
		}
	}
	return formatted
}

// AnalyzeDiagnostics correlaciona as evidências e constrói a lista de problemas traduzidos em Go puro
func AnalyzeDiagnostics(raw RawDiagnostics) []model.DiagnosticIssue {
	var issues []model.DiagnosticIssue

	// Análise 1: Minidump / BSOD
	if raw.MinidumpsRaw != "" && raw.MinidumpsRaw != "N/D" {
		issues = append(issues, model.DiagnosticIssue{
			Severity:       "CRÍTICO",
			Category:       "BSOD",
			Title:          "Despejos de Memória por Tela Azul (BSOD) Recentes",
			Source:         "C:\\Windows\\Minidump e Log 'System' (Event ID 1001 WER)",
			LogsEvidence:   fmt.Sprintf("Arquivos de crashdump no disco:\n%s\nDetalhes do BugCheck:\n%s", raw.MinidumpsRaw, raw.BugCheckRaw),
			Interpretation: "O sistema operacional sofreu interrupções fatais em modo kernel (BugCheck). A criação do arquivo minidump confirma que o kernel descarregou o estado da memória antes de reiniciar devido a falhas críticas de drivers ou hardware.",
			Remediation: []string{
				"Executar diagnóstico completo de memória RAM usando 'mdsched.exe' ou MemTest86.",
				"Abrir o minidump mais recente com WinDbg ou BlueScreenView para isolar o arquivo .sys causador.",
				"Atualizar drivers essenciais de chipset, controladora de disco e placa de vídeo.",
				"Executar 'sfc /scannow' e 'dism /online /cleanup-image /restorehealth' para reparar corrupções no sistema.",
			},
		})
	}

	// Análise 2: WHEA-Logger (Erros físicos de hardware)
	if raw.WHEARaw != "" && raw.WHEARaw != "N/D" {
		var evBuilder strings.Builder
		for _, line := range strings.Split(raw.WHEARaw, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			p := strings.Split(line, ";")
			if len(p) >= 2 {
				msg := "Alerta da arquitetura de hardware"
				switch p[1] {
				case "17":
					msg = "Erro corrigido de barramento PCI Express"
				case "18":
					msg = "ERRO FATAL DE PROCESSADOR (Machine Check Exception)"
				case "19":
					msg = "Erro corrigido de processador / cache"
				case "47":
					msg = "Erro de memória física corrigido por hardware"
				}
				evBuilder.WriteString(fmt.Sprintf("%s | Event ID %s | %s\n", p[0], p[1], msg))
			}
		}

		issues = append(issues, model.DiagnosticIssue{
			Severity:       "CRÍTICO",
			Category:       "HARDWARE",
			Title:          "Alertas de Falha Física na Arquitetura de Hardware (WHEA)",
			Source:         "Log 'System' (Provider: Microsoft-Windows-WHEA-Logger)",
			LogsEvidence:   strings.TrimSpace(evBuilder.String()),
			Interpretation: "A arquitetura de erros de hardware do Windows (WHEA) interceptou falhas em nível de barramento ou circuito integrado (PCIe, CPU Machine Check ou controladora de memória).",
			Remediation: []string{
				"Verificar se o slot PCIe da GPU/NVMe está limpo e bem encaixado.",
				"Desativar qualquer perfil de overclock ou XMP/DOCP na BIOS da máquina.",
				"Atualizar a BIOS para a versão estável mais recente disponibilizada pelo fabricante.",
				"Monitorar tensões da fonte de alimentação via software ou multímetro.",
			},
		})
	}

	// Análise 3: Desligamentos Inesperados / Kernel-Power (Event ID 41)
	if raw.KernelPowerRaw != "" && raw.KernelPowerRaw != "N/D" {
		var evBuilder strings.Builder
		for _, line := range strings.Split(raw.KernelPowerRaw, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			p := strings.Split(line, ";")
			if len(p) >= 1 {
				evBuilder.WriteString(fmt.Sprintf("%s | Event ID 41 | Reinício não limpo detectado (Perda de energia/Crash)\n", p[0]))
			}
		}

		issues = append(issues, model.DiagnosticIssue{
			Severity:       "ALTO",
			Category:       "ENERGIA/HARDWARE",
			Title:          "Reinicializações Inesperadas e Perda Abrupta de Energia",
			Source:         "Log 'System' (Event ID 41 Kernel-Power / Event ID 6008)",
			LogsEvidence:   strings.TrimSpace(evBuilder.String()),
			Interpretation: "O sistema reiniciou sem passar pela sequência padrão de desligamento (ACPI Shutdown). Isso ocorre por queda de energia elétrica externa, falha na fonte (PSU) ou travamento total da CPU.",
			Remediation: []string{
				"Verificar estabilidade da rede elétrica, cabo de alimentação e nobreak/filtro de linha.",
				"Inspecionar a fonte de alimentação do computador quanto a zumbidos ou capacitores estufados.",
				"Verificar se o desligamento coincidiu com momentos de pico de uso (jogos, renderização ou compilação).",
			},
		})
	}

	// Análise 4: Falhas e Degradação de Discos / Armazenamento
	if raw.DiskErrorsRaw != "" && raw.DiskErrorsRaw != "N/D" {
		var evBuilder strings.Builder
		for _, line := range strings.Split(raw.DiskErrorsRaw, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			p := strings.Split(line, ";")
			if len(p) >= 2 {
				msg := "Alerta no subsistema de armazenamento"
				switch p[1] {
				case "7":
					msg = "SETOR DEFEITUOSO (Bad Block) detectado no disco"
				case "11":
					msg = "Erro de hardware do driver da controladora"
				case "153":
					msg = "TIMEOUT DE E/S (Operação de leitura/escrita repetida)"
				case "129":
					msg = "Reset emitido para a porta do dispositivo"
				case "55":
					msg = "CORRUPÇÃO NTFS na estrutura do sistema de arquivos"
				case "98":
					msg = "Volume precisa de verificação offline (Chkdsk)"
				}
				evBuilder.WriteString(fmt.Sprintf("%s | Event ID %s | %s\n", p[0], p[1], msg))
			}
		}

		issues = append(issues, model.DiagnosticIssue{
			Severity:       "ALTO",
			Category:       "HARDWARE_DISCO",
			Title:          "Erros de E/S ou Corrupção Detectados na Unidade de Armazenamento",
			Source:         "Log 'System' (Event IDs 7, 11, 153, 129, 55, 98)",
			LogsEvidence:   strings.TrimSpace(evBuilder.String()),
			Interpretation: "Foram identificados timeouts de comunicação (Event ID 153) ou bad blocks físicos. Quando uma operação de leitura/escrita demora mais que o limite aceitável, a controladora reenvia o comando, causando travamentos momentâneos e risco iminente de perda de dados.",
			Remediation: []string{
				"Realizar imediatamente backup de segurança de todos os dados críticos.",
				"Executar 'chkdsk C: /f /r' na próxima reinicialização para isolar setores defeituosos.",
				"Inspecionar a saúde do SSD/HD através do software do fabricante (CrystalDiskInfo, Samsung Magician).",
				"Caso utilize cabo SATA, substituir o cabo e testar outra porta da placa-mãe.",
			},
		})
	}

	// Análise 5: SMART Alertas
	if strings.Contains(raw.SmartHealthRaw, "Unhealthy") || strings.Contains(raw.SmartHealthRaw, "Warning") {
		issues = append(issues, model.DiagnosticIssue{
			Severity:       "CRÍTICO",
			Category:       "HARDWARE_DISCO",
			Title:          "Status SMART de Integridade de Disco em Alerta / Degradação",
			Source:         "WMI MSStorageDriver_FailurePredictStatus / Get-PhysicalDisk",
			LogsEvidence:   raw.SmartHealthRaw,
			Interpretation: "O firmware do disco físico disparou um alarme de integridade pré-falha (SMART). Setores realocados ou falhas de leitura mecânica/flash ultrapassaram o limiar seguro de operação.",
			Remediation: []string{
				"Substituir o disco rígido ou SSD com urgência antes que a unidade entre em modo somente-leitura ou pare totalmente.",
				"Clonar a partição para um disco novo através de ferramenta de imagem segura.",
			},
		})
	}

	// Análise 6: Thermal Throttling / Temperatura
	if raw.ThermalRaw != "" && raw.ThermalRaw != "N/D" {
		issues = append(issues, model.DiagnosticIssue{
			Severity:       "ALTO",
			Category:       "DESEMPENHO_TERMICO",
			Title:          "Limitação de Frequência do Processador por Superaquecimento (Thermal Throttling)",
			Source:         "Log 'System' (Microsoft-Windows-Kernel-Processor-Power - Event ID 37)",
			LogsEvidence:   "Event ID 37 registrado: Velocidade do processador limitada pelo firmware devido a temperatura/energia.",
			Interpretation: "O firmware da placa-mãe forçou a redução drástica da velocidade da CPU para evitar que o silício atinja a temperatura de junção máxima (Tjunction). Causa lentidão severa sob uso moderado.",
			Remediation: []string{
				"Limpar o acúmulo de poeira nas ventoinhas, dissipador e saídas de ar do gabinete/notebook.",
				"Substituir a pasta térmica da CPU e GPU por uma pasta de alta condutividade térmica.",
				"Verificar se a bomba do Water Cooler ou o cooler a ar está operando na rotação correta.",
			},
		})
	}

	// Análise 7: Falhas em Módulos de Execução e DLLs (Event ID 1000)
	if raw.DllErrorsRaw != "" && raw.DllErrorsRaw != "N/D" {
		var evBuilder strings.Builder
		for _, line := range strings.Split(raw.DllErrorsRaw, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			p := strings.Split(line, ";")
			if len(p) >= 4 {
				evBuilder.WriteString(fmt.Sprintf("%s | App: %s | Módulo com falha: %s | Código: %s\n", p[0], p[1], p[2], p[3]))
			}
		}

		issues = append(issues, model.DiagnosticIssue{
			Severity:       "MÉDIO",
			Category:       "SISTEMA_DLL",
			Title:          "Falhas em Módulos e DLLs de Aplicações do Windows",
			Source:         "Log 'Application' (Event ID 1000 - Application Error)",
			LogsEvidence:   strings.TrimSpace(evBuilder.String()),
			Interpretation: "Módulos dinâmicos compartilhados (.dll) ou executáveis encerraram repentinamente por violação de acesso à memória (Access Violation 0xc0000005) ou corrupção de biblioteca de runtime.",
			Remediation: []string{
				"Reinstalar todos os pacotes redistribuíveis do Microsoft Visual C++ (x86 e x64).",
				"Executar 'sfc /scannow' para validar a integridade de DLLs do sistema em C:\\Windows\\System32.",
				"Reinstalar ou reparar a aplicação específica citada na evidência do log.",
			},
		})
	}

	// Análise 8: Falhas no Windows Update
	if raw.UpdateErrorsRaw != "" && raw.UpdateErrorsRaw != "N/D" {
		var evBuilder strings.Builder
		for _, line := range strings.Split(raw.UpdateErrorsRaw, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			p := strings.Split(line, ";")
			if len(p) >= 2 {
				evBuilder.WriteString(fmt.Sprintf("%s | Event ID %s | Falha na instalação/download de atualização\n", p[0], p[1]))
			}
		}

		issues = append(issues, model.DiagnosticIssue{
			Severity:       "MÉDIO",
			Category:       "ATUALIZACOES",
			Title:          "Falhas Persistentes na Instalação de Atualizações de Segurança",
			Source:         "Log 'Microsoft-Windows-WindowsUpdateClient/Operational' (Event ID 20/25)",
			LogsEvidence:   strings.TrimSpace(evBuilder.String()),
			Interpretation: "O subsistema do Windows Update não conseguiu baixar ou aplicar atualizações do sistema operacional, mantendo a máquina vulnerável a brechas conhecidas ou deixando pacotes incompletos.",
			Remediation: []string{
				"Interromper os serviços 'wuauserv' e 'bits' via prompt administrativo.",
				"Renomear a pasta 'C:\\Windows\\SoftwareDistribution' para 'SoftwareDistribution.old'.",
				"Reiniciar os serviços e acionar a verificação manual de atualizações.",
				"Executar 'dism /online /cleanup-image /restorehealth' para corrigir dependências ausentes.",
			},
		})
	}

	// Análise 9: Integridade de Segurança (Secure Boot)
	if raw.SecureBootRaw == "Desativado" {
		issues = append(issues, model.DiagnosticIssue{
			Severity:       "BAIXO",
			Category:       "SEGURANCA",
			Title:          "Inicialização Segura (Secure Boot) Encontra-se Desativada",
			Source:         "Firmware UEFI (Confirm-SecureBootUEFI)",
			LogsEvidence:   "Status retornado: Desativado",
			Interpretation: "O Secure Boot impede a execução de bootkits e drivers maliciosos antes do carregamento do kernel do Windows. Estando desativado, o sistema fica vulnerável a modificações no processo de boot.",
			Remediation: []string{
				"Acessar as configurações de setup da BIOS/UEFI durante a reinicialização.",
				"Navegar até a aba 'Security' ou 'Boot' e habilitar a opção 'Secure Boot'.",
			},
		})
	}

	// Análise 10: Integridade de Segurança (UAC Desativado)
	if raw.UACRaw == "0" {
		issues = append(issues, model.DiagnosticIssue{
			Severity:       "MÉDIO",
			Category:       "SEGURANCA",
			Title:          "Controle de Conta de Usuário (UAC) Encontra-se Totalmente Desativado",
			Source:         "Registro do Windows: HKLM:\\SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Policies\\System (EnableLUA)",
			LogsEvidence:   "EnableLUA = 0",
			Interpretation: "Com o UAC desativado, qualquer executável rodando na sessão do usuário obtém privilégios administrativos completos de forma silenciosa, sem solicitar consentimento.",
			Remediation: []string{
				"Restaurar o UAC para o nível recomendado nas Configurações de Controle de Conta de Usuário ou definindo EnableLUA = 1 no registro.",
			},
		})
	}

	_ = analyzer.LookupBSOD
	return issues
}
