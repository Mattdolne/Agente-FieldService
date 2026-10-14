# Base de Conhecimento em Go (Golang) para o Agente FieldService

Este documento serve como guia de estudo e referência técnica detalhada para você entender como a linguagem Go funciona e como ela foi utilizada para construir a arquitetura do **Agente FieldService**.

---

## 1. O que é Go e por que ele é perfeito para este agente?

Go (ou Golang) é uma linguagem criada pelo Google com três pilares centrais:
1. **Velocidade de Execução**: É compilada diretamente para código de máquina x64 (como C/C++), sem passar por máquinas virtuais lentas.
2. **Independência de Runtimes**: Não requer instalação de Python, Node.js nem frameworks externos.
3. **Binário Único Autossuficiente**: O compilador embuta dentro do executável `.exe` todas as bibliotecas e o próprio runtime. Você coloca o arquivo em um pendrive e roda em qualquer computador com Windows.

---

## 2. Conceitos Essenciais da Linguagem

### A. Pacotes (`package`) e Funções
Todo arquivo Go começa declarando seu pacote:
- `package main`: Identifica o ponto de entrada de um programa executável.
- Funções que começam com **letra maiúscula** (ex: `Print`, `CollectSystemInfo`, `ExportPDF`) são públicas (exportadas).
- Funções com **letra minúscula** (ex: `pad`, `runPS`) são privadas daquele pacote.

### B. Structs (Agrupamento de Dados sem Classes)
Go não usa classes nem herança. Em vez disso, agrupa propriedades com `struct`:
```go
type SystemInfo struct {
    Hostname       string
    Domain         string
    SerialNumber   string
    WindowsVersion string
    RAMTotal       string
}
```

### C. Tratamento de Erros sem `try/catch`
Em Go não existem exceções. As funções retornam o erro como um valor comum:
```go
arquivo, err := os.Create("relatorio.txt")
if err != nil {
    fmt.Println("Erro ao criar arquivo:", err)
    return
}
defer arquivo.Close() // Fecha o arquivo automaticamente ao final da função!
```

---

## 3. O Desafio dos Acentos no Windows e a Solução Técnica

### Por que `Ç`, `Ã` e acentos viravam `?`?
No Windows, existem duas armadilhas históricas de codificação que provocavam o problema:

1. **A Armadilha do PowerShell (`$OutputEncoding`)**:
   - No Windows PowerShell 5.1 nativo, quando um script gera uma string contendo `INICIALIZAÇÃO` e envia para um pipe capturado pelo Go, o PowerShell converte o texto usando sua variável interna `$OutputEncoding`, que por padrão é **US-ASCII** ou **CP850 (OEM)**. Caracteres que não cabem no padrão viram `?`.
   - **Solução Arquitetural Aplicada**: Mudamos a responsabilidade! O PowerShell agora apenas coleta dados atômicos (números de evento como `6005`, `98`, `1074` e timestamps). Toda a tradução para português humano (`Inicialização`, `Corrupção`, `Verificação`) é feita **100% dentro do código Go**. Como o Go trabalha nativamente em UTF-8, o PowerShell nunca mais toca nos acentos!

2. **A Armadilha do Console do Windows (`WriteFile` vs `WriteConsoleW`)**:
   - As funções comuns de impressão do Go (`fmt.Println` ou `os.Stdout.Write`) usam internamente a API `WriteFile` do Windows, que envia bytes brutos.
   - O console do Windows (`conhost.exe`) tem falhas históricas ao converter bytes UTF-8 para fontes antigas do Prompt, gerando distorções.
   - **Solução Definitiva com `WriteConsoleW`**:
     Criamos o pacote `pkg/console`, que acessa a API nativa de caracteres largos do Windows:
     ```go
     var (
         kernel32          = syscall.NewLazyDLL("kernel32.dll")
         procGetStdHandle  = kernel32.NewProc("GetStdHandle")
         procWriteConsoleW = kernel32.NewProc("WriteConsoleW")
     )
     ```
     Convertendo o texto em Go para UTF-16 (`utf16.Encode([]rune(s))`) e chamando `WriteConsoleW`, o Windows renderiza qualquer caractere Unicode diretamente do motor de fontes, com **zero dependência de Code Pages e zero distorção**.

---

## 4. Como Construir Tabelas e Caixas Alinhadas no Terminal

Para que tabelas e caixas desenhadas com caracteres gráficos (`┌─┐`, `╔═╗`) não fiquem tortas quando uma linha contém acentos, não podemos usar a função `len(s)` do Go:
- Em UTF-8, a letra `á` ou `ç` ocupa **2 bytes**, mas na tela ocupa apenas **1 coluna visual**!
- Se usarmos `len("Maçã")`, o Go dirá que o tamanho é **6 bytes**, e o alinhamento da borda direita da tabela vai entortar.

### A Solução com `runes` (`pad`):
No Go, um caractere Unicode individual é chamado de **`rune`**:
```go
func pad(s string, width int) string {
    runes := []rune(s) // Converte a string para lista de caracteres reais visuais
    if len(runes) >= width {
        return string(runes[:width])
    }
    return s + strings.Repeat(" ", width-len(runes))
}
```
Com essa função, medimos a largura visual real da palavra e preenchemos com os espaços exatos, garantindo que as bordas `│` e `║` fiquem rigorosamente alinhadas na vertical.

### B. Quebra Automática de Linhas (`wrapText`) e Recuo Francês
Quando um texto de diagnóstico técnico possui centenas de caracteres, colocá-lo diretamente dentro de uma caixa estouraria o terminal ou deixaria a borda direita aberta.

Para resolver isso, criamos o algoritmo `wrapText`:
1. Divide o parágrafo em palavras (`strings.Fields`).
2. Mede a largura acumulada em `runes`.
3. Quando a próxima palavra ultrapassa a largura máxima útil (`contentWidth = 94`), quebra para a próxima linha.
4. Na função `cardParagraph`, aplicamos **recuo francês (hanging indent)**: a primeira linha recebe o marcador (ex: `   [3] `) e as linhas seguintes recebem espaços equivalentes (`       `), mantendo o alinhamento visual perfeito.
5. Cada linha é delimitada por `║ ` e ` ║`, garantindo caixas 100% fechadas e simétricas.

---

## 5. Como Compilar para Produção

Na pasta do código fonte:
```powershell
go build -ldflags="-s -w" -o AgenteFieldService.exe main.go
```
- `-s`: Remove símbolos de depuração (*symbol table*).
- `-w`: Desativa informações de debug DWARF.
- O resultado é um executável de ~4.0 MB, ultrarrápido e estático.
