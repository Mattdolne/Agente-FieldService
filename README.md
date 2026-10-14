# Agente FieldService (`AgenteFieldService.exe`)

Agente autônomo de diagnóstico avançado de hardware, integridade e sistema operacional Windows voltado para suporte técnico de campo (*Field Service*). 
Compilado em um único arquivo `.exe`, **sem dependências externas**, projetado para execução imediata em qualquer máquina (Windows 10, 11 e Windows Server) com interface em **tabelas e fichas técnicas no terminal** e exportação para **TXT**, **CSV** e **PDF**.

---

## 📍 Localização do Projeto

O projeto reside na pasta Documentos:
`C:\Users\mattheus.pereira.spt\Documents\Agente FieldService`

---

## 🎨 Novo Layout Visual Estruturado

O `AgenteFieldService.exe` abandona textos corridos e apresenta os dados em 3 camadas visuais limpas:

1. **Formulário de Hardware e Sistema**: Grade em duas colunas com alinhamento preciso de Hostname, Domínio, Serial, Versão do Windows, CPU, Memória RAM (uso/total com %), Armazenamento e BIOS.
2. **Tabela Grid de Reinicializações Recentes**: Colunas delimitadas (`DATA E HORA | EVENTO | CLASSIFICAÇÃO / MOTIVO`) para fácil identificação de reboots forçados (Event ID 41/6008) vs boots limpos (6005).
3. **Fichas Técnicas de Incidente (Diagnostic Cards)**: Cada problema encontrado é encapsulado em uma caixa de inspeção (`╔═══╗`), contendo:
   - Severidade e Título
   - 📍 Origem do log / arquivo
   - 📋 Evidências coletadas
   - 🔬 Interpretação técnica (Causa Raiz)
   - 🛠️ Lista numerada de tarefas recomendadas para resolução

---

## 🔤 Resolução Definitiva de Caracteres e Acentos (`Ç`, `Ã`, etc.)

Para eliminar caracteres corrompidos (`?` ou símbolos estranhos), a aplicação implementa duas arquiteturas conjuntas:

1. **Separação de Coleta e Apresentação**:
   - O PowerShell em segundo plano **apenas extrai dados brutos, códigos e números** (ex: `6005`, `98`, `1074`, timestamps).
   - O Go traduz e monta 100% dos textos em português nativo (`Inicialização`, `Corrupção`, `Verificação`, etc.).
2. **API Nativa do Windows (`WriteConsoleW`)**:
   - O pacote `pkg/console` utiliza a chamada `WriteConsoleW` da `kernel32.dll`.
   - Isso envia caracteres diretamente em **UTF-16 (Unicode nativo)** para o terminal, contornando qualquer bug de Code Page (CP850/CP1252) do prompt clássico do Windows.

---

## ⚙️ Como Compilar o Binário `.exe`

No PowerShell:

```powershell
# 1. Navegue até o diretório do projeto
cd "$HOME\Documents\Agente FieldService"

# 2. Compile o executável estático otimizado
& "C:\Program Files\Go\bin\go.exe" build -ldflags="-s -w" -o AgenteFieldService.exe main.go
```

O arquivo gerado `AgenteFieldService.exe` tem ~4.0 MB e está pronto para ser copiado para um pendrive ou compartilhado via rede.

---

## 🖥️ Como Executar

1. Abra o Prompt de Comando ou PowerShell **como Administrador**.
2. Execute:
   ```cmd
   AgenteFieldService.exe
   ```
3. Acompanhe a varredura e utilize o menu interativo para salvar em **TXT**, **CSV** ou **PDF**.
