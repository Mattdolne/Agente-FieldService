# Contexto Operacional para Agentes de IA (`GEMINI.md`)

Este arquivo define todas as especificações técnicas, schemas de dados, dicionários de erros e diretrizes de desenvolvimento para agentes de IA que interagem com ou evoluem o **`Agente FieldService`**.

> [!IMPORTANT]
> **Regra de Manutenção do Projeto:**
> Sempre que houver qualquer alteração na arquitetura, novos diagnósticos adicionados ou modificações nos arquivos de código, este arquivo (`GEMINI.md`), o `README.md` e o `GO_KNOWLEDGE_BASE.md` **devem ser atualizados imediatamente**.

---

## 1. Localização e Arquitetura do Projeto

- **Caminho Base**: `C:\Users\mattheus.pereira.spt\Documents\Agente FieldService`
- **Executável Final**: `AgenteFieldService.exe`

### Camadas Desacopladas:
1. **`Collector` (`pkg/collector`)**:
   - `system.go`: Extrai WMI/CIM de hardware (CPU, RAM, Discos, BIOS, Hostname, Domínio, Serial).
   - `diagnostics.go`: Executa comandos PowerShell retornando **apenas chaves atômicas, IDs de eventos e timestamps**. Não gera strings complexas em português dentro do PowerShell para evitar distorção de codificação.
2. **`Console` (`pkg/console`)**:
   - Implementa renderização direta em **UTF-16** via `WriteConsoleW` da `kernel32.dll`, garantindo que acentos (`ç`, `ã`, `é`, `ó`, etc.) e caracteres de desenho de caixas (`┌─┐`, `╔═╗`) nunca se corrompam no prompt do Windows.
3. **`Analyzer` (`pkg/analyzer`)**:
   - Dicionário extensivo de causas raízes e remediações (WHEA, BSOD, Disk Timeouts, Thermal Throttling, DLLs e Windows Update).
   - Classifica em severidades: `CRÍTICO`, `ALTO`, `MÉDIO`, `BAIXO`.
4. **`Exporter` (`pkg/exporter`)**:
   - Formata a saída do terminal em **Formulário de Hardware**, **Grid de Reboots** e **Fichas Técnicas de Diagnóstico**.
   - Exporta sob demanda para **TXT**, **CSV** e **PDF**.

---

## 2. Padrão de Codificação e Tratamento de Strings

Ao adicionar novos comandos do PowerShell no `collector`:
- **NUNCA** inclua palavras com acentuação ou cedilha no corpo do script do PowerShell.
- Em vez disso, retorne IDs simples (ex: `6005`, `98`, `THROTTLING`, etc.) delimitados por ponto e vírgula `;`.
- Deixe o código Go mapear esses IDs para as strings amigáveis em português dentro de `diagnostics.go` ou `dictionary.go`.
- Use a função `pad(string, largura)` em `exporter.go` para alinhar tabelas com base na contagem de `runes` (para que caracteres multibyte não desloquem a borda direita).

---

## 3. Diretrizes de Escopo

1. **Sem BitLocker**: Não reinserir nenhuma verificação ou alerta de BitLocker.
2. **Offline-First**: Nenhuma funcionalidade deve depender obrigatoriamente de conexão de rede ou login em contas.
3. **Execução Autônoma**: O operador de campo só interage ao final para escolher o formato do arquivo a ser salvo.
