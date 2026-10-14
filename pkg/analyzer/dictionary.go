package analyzer

// BSODRule descreve um Stop Code conhecido de Tela Azul
type BSODRule struct {
	Code           string
	Name           string
	Interpretation string
	Remediation    []string
}

// UpdateErrorRule descreve um código HRESULT de falha do Windows Update
type UpdateErrorRule struct {
	Code           string
	Description    string
	Remediation    []string
}

// HardwareRule descreve eventos físicos de hardware
type HardwareRule struct {
	EventID        int
	Source         string
	Title          string
	Interpretation string
	Remediation    []string
}

// LookupBSOD busca no dicionário extensivo pelo StopCode
func LookupBSOD(codeHex string) BSODRule {
	rules := map[string]BSODRule{
		"0x0000009F": {
			Code: "0x0000009F",
			Name: "DRIVER_POWER_STATE_FAILURE",
			Interpretation: "Um driver de dispositivo não respondeu a uma transição de energia (hibernação, suspensão ACPI ou desligamento dentro do tempo limite).",
			Remediation: []string{
				"Atualizar os drivers de rede sem fio (Wi-Fi) e placa gráfica para as versões oficiais do fabricante.",
				"Desativar a Inicialização Rápida (Fast Startup) nas opções de energia do Painel de Controle.",
				"Verificar se há atualizações de firmware da BIOS para compatibilidade ACPI.",
			},
		},
		"0x00000124": {
			Code: "0x00000124",
			Name: "WHEA_UNCORRECTABLE_ERROR",
			Interpretation: "Falha de hardware irrecuperável detectada pela arquitetura WHEA (geralmente ligada à CPU, barramento PCIe ou sub-tensão).",
			Remediation: []string{
				"Verificar temperaturas de CPU e GPU sob carga para descartar superaquecimento.",
				"Restaurar configurações de overclock ou perfil XMP na BIOS para o padrão (Default).",
				"Testar módulos de memória RAM individualmente com o Windows Memory Diagnostic.",
				"Atualizar a BIOS da placa-mãe para o microcódigo mais recente do processador.",
			},
		},
		"0x000000D1": {
			Code: "0x000000D1",
			Name: "DRIVER_IRQL_NOT_LESS_OR_EQUAL",
			Interpretation: "Um driver em modo kernel tentou acessar um endereço de memória inválido em um nível de requisição de interrupção (IRQL) inadequado.",
			Remediation: []string{
				"Identificar o driver causador (.sys) nos arquivos de Minidump e reinstalá-lo.",
				"Executar 'sfc /scannow' para validar os arquivos do subsistema do Windows.",
				"Desinstalar softwares de virtualização ou antivírus de terceiros em caso de conflito recente.",
			},
		},
		"0x0000000A": {
			Code: "0x0000000A",
			Name: "IRQL_NOT_LESS_OR_EQUAL",
			Interpretation: "Acesso ilegal à memória de paginação por um processo do sistema, indicativo comum de corrupção de memória RAM ou driver instável.",
			Remediation: []string{
				"Executar diagnóstico de memória RAM (mdsched.exe).",
				"Remover drivers recém-instalados ou reverter para ponto de restauração anterior.",
			},
		},
		"0x0000003B": {
			Code: "0x0000003B",
			Name: "SYSTEM_SERVICE_EXCEPTION",
			Interpretation: "Exceção em rotina de serviço do sistema durante a execução do código de usuário/kernel (muito frequente em drivers de vídeo ou antivírus).",
			Remediation: []string{
				"Atualizar ou reinstalar de forma limpa o driver da GPU (utilizando DDU se necessário).",
				"Verificar se o aplicativo que disparou o erro possui atualização de compatibilidade.",
			},
		},
		"0x0000001A": {
			Code: "0x0000001A",
			Name: "MEMORY_MANAGEMENT",
			Interpretation: "Erro grave na gerência de memória do Windows. Quase invariavelmente associado a pente de RAM físico defeituoso ou setor com defeito onde reside o pagefile.sys.",
			Remediation: []string{
				"Executar 'mdsched.exe' ou MemTest86 completo por pelo menos 2 ciclos.",
				"Limpar os contatos dos pentes de memória e alternar os slots na placa-mãe.",
				"Recriar o arquivo de paginação (pagefile.sys) em outra unidade íntegra.",
			},
		},
		"0x00000050": {
			Code: "0x00000050",
			Name: "PAGE_FAULT_IN_NONPAGED_AREA",
			Interpretation: "O sistema operacional requisitou dados em uma área da memória que deveria existir fisicamente, mas a leitura falhou.",
			Remediation: []string{
				"Testar integridade da memória RAM física.",
				"Verificar saúde do disco rígido/SSD com 'chkdsk /f /r'.",
			},
		},
		"0x0000007B": {
			Code: "0x0000007B",
			Name: "INACCESSIBLE_BOOT_DEVICE",
			Interpretation: "O Windows perdeu o acesso à partição de sistema durante o processo de inicialização (driver de controladora AHCI/NVMe corrompido ou partição com falha).",
			Remediation: []string{
				"Verificar modo de operação SATA/NVMe na BIOS (AHCI vs RAID/VMD).",
				"Executar reparo de inicialização do Windows via ambiente WinRE.",
			},
		},
		"0x00000133": {
			Code: "0x00000133",
			Name: "DPC_WATCHDOG_VIOLATION",
			Interpretation: "Uma chamada de procedimento adiada (DPC) travou a CPU além do tempo limite aceitável (comum em SSDs com firmware desatualizado ou placas Wi-Fi).",
			Remediation: []string{
				"Atualizar o firmware do SSD via software do fabricante (Samsung Magician, Crucial Storage Executive, etc.).",
				"Atualizar os drivers do controlador SATA AHCI padrão e da placa Wi-Fi.",
			},
		},
		"0x00000116": {
			Code: "0x00000116",
			Name: "VIDEO_TDR_FAILURE",
			Interpretation: "O mecanismo Timeout Detection and Recovery (TDR) do Windows tentou resetar o driver de vídeo travado, mas a GPU não respondeu.",
			Remediation: []string{
				"Reinstalar o driver de vídeo mais recente de forma limpa.",
				"Verificar a integridade física da placa de vídeo e temperaturas de operação.",
			},
		},
	}

	if rule, found := rules[codeHex]; found {
		return rule
	}

	// Regra genérica para outros stop codes
	return BSODRule{
		Code:           codeHex,
		Name:           "GENERIC_BUGCHECK",
		Interpretation: "Despejo de memória por erro crítico de sistema registrado no log.",
		Remediation: []string{
			"Analisar o arquivo minidump específico com WinDbg ou BlueScreenView.",
			"Executar 'sfc /scannow' e 'dism /online /cleanup-image /restorehealth'.",
		},
	}
}

// LookupUpdateError busca no dicionário de códigos de erro do Windows Update
func LookupUpdateError(hresult string) UpdateErrorRule {
	rules := map[string]UpdateErrorRule{
		"0x80070002": {
			Code:        "0x80070002",
			Description: "ERROR_FILE_NOT_FOUND: O arquivo de instalação não foi encontrado ou está corrompido no cache local.",
			Remediation: []string{
				"Parar serviços wuauserv e bits via CMD.",
				"Renomear ou excluir a pasta C:\\Windows\\SoftwareDistribution.",
				"Reiniciar os serviços e disparar nova busca de atualizações.",
			},
		},
		"0x80070490": {
			Code:        "0x80070490",
			Description: "ERROR_NOT_FOUND: Corrupção nos metadados da loja de componentes do Windows (CBS Store).",
			Remediation: []string{
				"Executar 'dism /online /cleanup-image /restorehealth' conectado à internet.",
				"Executar 'sfc /scannow' em seguida para reparar manifestos corrompidos.",
			},
		},
		"0x80070643": {
			Code:        "0x80070643",
			Description: "ERROR_INSTALL_FAILURE: Falha fatal de execução de script do pacote (frequente em partições WinRE pequenas ou .NET Framework).",
			Remediation: []string{
				"Verificar tamanho da partição de recuperação (WinRE); se menor que 250MB livres, expandi-la.",
				"Executar a ferramenta de reparo do Microsoft .NET Framework.",
			},
		},
		"0x800f0922": {
			Code:        "0x800f0922",
			Description: "CBS_E_INSTALLERS_FAILED: Falha durante o rollback da atualização ou partição reservada do sistema (EFI) sem espaço suficiente.",
			Remediation: []string{
				"Verificar se a partição do sistema EFI tem pelo menos 50MB livres.",
				"Garantir que serviços como AppX Deployment e Windows Event Log estejam em execução.",
			},
		},
		"0x800f081f": {
			Code:        "0x800f081f",
			Description: "CBS_E_SOURCE_MISSING: Os arquivos binários de origem necessários para aplicar o pacote não foram localizados.",
			Remediation: []string{
				"Executar restauração do repositório DISM apontando para uma ISO oficial do Windows como fonte.",
			},
		},
	}

	if rule, found := rules[hresult]; found {
		return rule
	}

	return UpdateErrorRule{
		Code:        hresult,
		Description: "Código de erro retornado durante a instalação do pacote.",
		Remediation: []string{
			"Executar a Solução de Problemas do Windows Update nas configurações.",
			"Tentar baixar e instalar o pacote KB manualmente pelo Catálogo do Microsoft Update.",
		},
	}
}

