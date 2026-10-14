package collector

import (
	"fmt"
	"os/exec"
	"strings"
	"win-diag-agent/pkg/model"
)

// RunPS executa um comando no PowerShell de forma silenciosa e retorna a saída limpa em UTF-8
func RunPS(cmd string) string {
	fullCmd := fmt.Sprintf("$OutputEncoding = [Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false); $ErrorActionPreference='SilentlyContinue'; %s", cmd)
	out, err := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", fullCmd).Output()
	if err != nil {
		return "N/D"
	}
	return strings.TrimSpace(string(out))
}

// CollectSystemInfo obtém os dados do cabeçalho de hardware e sistema
func CollectSystemInfo() model.SystemInfo {
	info := model.SystemInfo{}

	// Hostname e Domínio
	info.Hostname = RunPS("$env:COMPUTERNAME")
	domain := RunPS("(Get-CimInstance Win32_ComputerSystem).Domain")
	partOfDomain := RunPS("(Get-CimInstance Win32_ComputerSystem).PartOfDomain")
	if partOfDomain == "True" {
		info.Domain = domain + " (AD)"
	} else if domain != "" && domain != "N/D" {
		info.Domain = domain + " (Workgroup)"
	} else {
		info.Domain = "WORKGROUP"
	}

	// Service Tag / Serial da Placa-Mãe ou Chassi
	serial := RunPS("(Get-CimInstance Win32_BIOS).SerialNumber")
	if serial == "" || serial == "To be filled by O.E.M." || serial == "Default string" {
		serialBase := RunPS("(Get-CimInstance Win32_BaseBoard).SerialNumber")
		if serialBase != "" && serialBase != "N/D" {
			serial = serialBase
		}
	}
	info.SerialNumber = serial

	// Versão do Windows (Nome Comercial + Build)
	osCaption := RunPS("(Get-CimInstance Win32_OperatingSystem).Caption")
	osBuild := RunPS("(Get-CimInstance Win32_OperatingSystem).BuildNumber")
	osDisplayVersion := RunPS("(Get-ItemProperty 'HKLM:\\SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion').DisplayVersion")
	if osDisplayVersion != "" && osDisplayVersion != "N/D" {
		info.WindowsVersion = fmt.Sprintf("%s %s (Build %s)", osCaption, osDisplayVersion, osBuild)
	} else {
		info.WindowsVersion = fmt.Sprintf("%s (Build %s)", osCaption, osBuild)
	}

	// Processador
	info.Processor = RunPS("(Get-CimInstance Win32_Processor | Select-Object -First 1).Name")

	// Memória RAM (Total e Usada com porcentagem)
	ramRaw := RunPS(`
		$os = Get-CimInstance Win32_OperatingSystem;
		$tot = [math]::Round($os.TotalVisibleMemorySize/1MB, 1);
		$free = [math]::Round($os.FreePhysicalMemory/1MB, 1);
		$used = [math]::Round($tot - $free, 1);
		$pct = [math]::Round(($used / $tot) * 100, 0);
		"$used;$tot;$pct"
	`)
	ramParts := strings.Split(ramRaw, ";")
	if len(ramParts) == 3 {
		info.RAMTotal = ramParts[1] + " GB"
		info.RAMUsed = fmt.Sprintf("%s GB (%s%%)", ramParts[0], ramParts[2])
	} else {
		info.RAMTotal = "N/D"
		info.RAMUsed = "N/D"
	}

	// Armazenamento (Unidade C: disponível e total)
	diskRaw := RunPS(`
		$d = Get-CimInstance Win32_LogicalDisk -Filter "DeviceID='C:'";
		$free = [math]::Round($d.FreeSpace/1GB, 1);
		$total = [math]::Round($d.Size/1GB, 1);
		$usedPct = [math]::Round((($total - $free)/$total)*100, 0);
		"$free;$total;$usedPct"
	`)
	diskParts := strings.Split(diskRaw, ";")
	if len(diskParts) == 3 {
		info.Storage = fmt.Sprintf("C: %s GB livres / %s GB (%s%%)", diskParts[0], diskParts[1], diskParts[2])
	} else {
		info.Storage = "C: N/D"
	}

	// Versão da BIOS
	biosRaw := RunPS(`
		$b = Get-CimInstance Win32_BIOS;
		$vendor = $b.Manufacturer;
		$ver = $b.SMBIOSBIOSVersion;
		$date = if ($b.ReleaseDate) { $b.ReleaseDate.ToString('dd/MM/yyyy') } else { 'N/D' };
		"$vendor $ver ($date)"
	`)
	info.BIOSVersion = biosRaw

	return info
}
