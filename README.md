# Agente FieldService (`AgenteFieldService.exe`)

O **Agente FieldService** é uma ferramenta autônoma de diagnóstico técnico de hardware, integridade e sistema operacional Windows voltada para analistas e técnicos de suporte de campo (*Field Service*).

Projetado como um executável único (`.exe`), **totalmente portátil e sem necessidade de instalação de dependências**, o agente pode ser inserido em qualquer computador (Windows 10, 11 e Windows Server) via pendrive, rede local ou sistema de gerenciamento remoto (RMM), gerando uma varredura completa da máquina em poucos segundos.

---

## 🚀 Principais Funcionalidades

### 1. Cabeçalho de Inventário da Máquina
Ao iniciar, o agente captura instantaneamente os dados fundamentais do dispositivo:
- **Hostname** e **Domínio** (Active Directory ou Workgroup);
- **Tag / Número de Série** (Chassi e Placa-Mãe);
- **Versão do Windows** (Edição, Versão e Número da Build);
- **Processador** (Modelo completo e frequência);
- **Memória RAM** (Volume em uso, capacidade total e porcentagem);
- **Armazenamento** (Espaço livre, capacidade total da unidade do SO e taxa de ocupação);
- **Versão da BIOS** (Fabricante, versão do SMBIOS e data de lançamento).

### 2. Diagnóstico das Últimas 5 Reinicializações
- Leitura cronológica dos últimos ciclos de boot e desligamento no log `System`:
  - `Event ID 6005`: Inicialização limpa do sistema;
  - `Event ID 6006`: Desligamento normal e ordenado;
  - `Event ID 6008`: Desligamento inesperado registrado no boot seguinte;
  - `Event ID 41`: *Kernel-Power* (reinicialização abrupta por perda de força ou crash de kernel);
  - `Event ID 1074`: Reinicialização/desligamento formal ordenado por usuário, serviço ou Windows Update.

### 3. Detecção e Tradução de Telas Azuis (BSOD)
- Varredura de arquivos em `C:\Windows\Minidump` e análise do Event ID `1001` (*Windows Error Reporting*).
- Dicionário extensivo embutido com tradução técnica imediata de Stop Codes como:
  - `0x0000009F` (`DRIVER_POWER_STATE_FAILURE`)
  - `0x00000124` (`WHEA_UNCORRECTABLE_ERROR`)
  - `0x000000D1` (`DRIVER_IRQL_NOT_LESS_OR_EQUAL`)
  - `0x0000001A` (`MEMORY_MANAGEMENT`)
  - `0x00000050` (`PAGE_FAULT_IN_NONPAGED_AREA`)
  - `0x00000133` (`DPC_WATCHDOG_VIOLATION`)
  - `0x00000116` (`VIDEO_TDR_FAILURE`), entre outros.

### 4. Saúde Física de Hardware e Armazenamento
- **Arquitetura WHEA (*Windows Hardware Error Architecture*)**: Rastreamento de erros corrigidos e fatais de barramento PCI Express, CPU e memória RAM (Event IDs `17`, `18`, `19` e `47`).
- **Discos e Controladoras**: Detecção de timeouts de E/S repetidos (Event ID `153`), erros de controladora (Event ID `11`), setores defeituosos (*Bad Blocks* - Event ID `7`) e corrupção de estrutura NTFS (Event IDs `55` e `98`).
- **Status SMART**: Avaliação da saúde física reportada pelo firmware das unidades de disco.
- **Controle Térmico**: Detecção de *Thermal Throttling* imposto pelo firmware por superaquecimento do processador (Event ID `37`).

### 5. Aplicações, Módulos DLL e Atualizações
- **Falhas de Aplicativo (Event ID `1000`)**: Identificação do executável, módulo DLL causador e código de exceção de memória.
- **Windows Update (Event IDs `20` e `25`)**: Rastreamento de falhas persistentes na instalação ou download de pacotes de segurança.

### 6. Integridade de Segurança
- Status da Inicialização Segura (*Secure Boot*);
- Status da Proteção em Tempo Real do Antivírus / Windows Defender;
- Status do Controle de Conta de Usuário (UAC - *EnableLUA*).

---

## 🎨 Interface no Terminal

O agente organiza a apresentação das informações no prompt em três blocos estruturados:

1. **Formulário de Hardware**: Tabela delimitada em duas colunas com alinhamento visual de dados do computador e especificações físicas.
2. **Grade de Reinicializações**: Linha do tempo tabular com data, hora, código do evento e motivo detalhado.
3. **Fichas Técnicas de Diagnóstico**: Caixas simétricas de inspeção para cada incidente encontrado, contendo:
   - Severidade do alerta (`CRÍTICO`, `ALTO`, `MÉDIO`, `BAIXO`);
   - Origem exata do incidente (log, arquivo ou chave do registro);
   - Evidências brutas coletadas;
   - Interpretação técnica da causa raiz;
   - Lista numerada de tarefas recomendadas para resolução.

---

## 💾 Exportação de Relatórios

Ao término da varredura, o programa disponibiliza um menu interativo com as opções:
- **`[1] TXT`**: Salva o relatório textual idêntico ao console para rápida anexação em chamados e tickets de suporte.
- **`[2] CSV`**: Gera uma planilha estruturada com separador `;` para importação direta em ferramentas de inventário e planilhas.
- **`[3] PDF`**: Gera um documento formatado e diagramado com tabelas e cabeçalhos coloridos por nível de criticidade.
- **`[4] Todos`**: Gera os três formatos simultaneamente.

---

## 🖥️ Como Utilizar em Produção

1. Copie o executável `AgenteFieldService.exe` para o computador alvo.
2. Abra o Prompt de Comando ou PowerShell **como Administrador** (necessário para acesso a logs do sistema e telemetria de disco).
3. Execute:
   ```cmd
   AgenteFieldService.exe
   ```
4. Acompanhe a varredura no terminal e escolha o formato de exportação desejado.
