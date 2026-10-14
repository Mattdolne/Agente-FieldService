package model

// SystemInfo armazena os dados do cabeçalho da máquina
type SystemInfo struct {
	Hostname       string `json:"hostname"`
	Domain         string `json:"domain"`
	SerialNumber   string `json:"serial_number"`
	WindowsVersion string `json:"windows_version"`
	Processor      string `json:"processor"`
	RAMTotal       string `json:"ram_total"`
	RAMUsed        string `json:"ram_used"`
	Storage        string `json:"storage"`
	BIOSVersion    string `json:"bios_version"`
}

// DiagnosticIssue representa um problema identificado pelo agente
type DiagnosticIssue struct {
	Severity       string   `json:"severity"`       // CRÍTICO, ALTO, MÉDIO, BAIXO
	Category       string   `json:"category"`       // HARDWARE, BSOD, DLL, ATUALIZACAO, DESEMPENHO, SEGURANCA
	Title          string   `json:"title"`          // Nome claro do incidente
	Source         string   `json:"source"`         // Onde foi encontrado (log, arquivo, registro)
	LogsEvidence   string   `json:"logs_evidence"`  // Quais logs e dados brutos
	Interpretation string   `json:"interpretation"` // Como foi interpretado e causa raiz
	Remediation    []string `json:"remediation"`    // Lista de passos práticos para resolver
}

// RebootEntry representa um registro de inicialização/desligamento
type RebootEntry struct {
	Timestamp string `json:"timestamp"`
	EventID   int    `json:"event_id"`
	Type      string `json:"type"`
	Reason    string `json:"reason"`
}

