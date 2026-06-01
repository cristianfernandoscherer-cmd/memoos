# MemoOS

MemoOS é uma infraestrutura de memória semântica local para agentes e LLMs.

## Visão Geral

- Memória persistente em SQLite
- Busca semântica via embeddings
- Embeddings locais (Ollama e ONNX)
- Integração MCP (Model Context Protocol)
- Suporte multi-projeto
- Categorização contextual

## Quick Start

### Instalação

```bash
# Clonar o repositório
git clone https://github.com/cristian-scherer/memoos.git
cd memoos

# Compilar binários
make build

# Ou executar diretamente (build automático)
make run        # Executa servidor MCP
make run-cli    # Executa interface CLI
```

### Opcional: Bibliotecas ONNX

Para embeddings locais (sem depender de Ollama):

1. **Baixe as bibliotecas ONNX** (não versionadas no Git):
   ```bash
   make download-libs  # Baixa para Linux/macOS/Windows
   ```

2. **As bibliotecas são salvas em** `libs/` e ignoradas pelo Git

3. **Se não baixar**, o sistema usará:
   - Ollama (se disponível) para embeddings
   - Ou não terá embeddings locais

### Configuração

```bash
# Copiar configuração de exemplo
cp configs/config.example.yaml ~/.config/memoos/config.yaml

# Editar configuração
vim ~/.config/memoos/config.yaml
```

### Uso

#### CLI

```bash
# Compilar (se não fez antes)
make build

# Salvar memória
./bin/memoos-cli add \
  --cwd /home/user/projects/whmcs \
  --category payments \
  --content "Refund Pix reutiliza e2eid"

# Buscar memórias
./bin/memoos-cli search \
  --cwd /home/user/projects/whmcs \
  --query "como funciona estorno pix?"
```

#### MCP Server

Configurar no Cursor em `.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "memoos": {
      "command": "bash",
      "args": [
        "-c",
        "cd /home/cristian-scherer/Documentos/projetos/memoos && exec ./bin/memoos-server"
      ]
    }
  }
}
```

**Nota:** A configuração acima usa `bash` para iniciar o servidor a partir do diretório do projeto, garantindo que o caminho absoluto `/home/cristian-scherer/Documentos/projetos/memoos` seja usado corretamente.

## Ferramentas MCP

As ferramentas MCP permitem salvar e recuperar memórias usando linguagem natural. Aqui estão exemplos de como usar:

### Exemplo 1: Salvar uma memória

**Formas simples de pedir para salvar:**

> "Salve essa informação no memoos: No projeto WHMCS, o processamento de estorno Pix reutiliza o e2eid."
> 
> "Guarde isso: O e2eid é reutilizado nos processamentos de estorno Pix do WHMCS."
>
> "Memorize isso: Ref Pix no WHMCS reutiliza e2eid, categoria pagamentos."

**A ferramenta `memory_save` será chamada automaticamente com:**
```json
{
  "cwd": "/caminho/do/projeto/whmcs",
  "category": "pagamentos",
  "content": "No projeto WHMCS, o processamento de estorno Pix reutiliza o e2eid.",
  "metadata": {
    "fonte": "usuário",
    "contexto": "desenvolvimento"
  }
}
```

### Exemplo 2: Buscar memórias relacionadas

**Prompt para o assistente:**
> "Procure por informações sobre como funciona o estorno Pix no projeto WHMCS."

**A ferramenta `memory_search` será chamada com:**
```json
{
  "cwd": "/caminho/do/projeto/whmcs",
  "query": "como funciona estorno Pix",
  "category": "pagamentos",
  "limit": 5,
  "min_score": 0.7
}
```

### Exemplo 3: Listar memórias recentes

**Prompt para o assistente:**
> "Mostra as memórias recentes sobre pagamentos no projeto atual."

**A ferramenta `memory_list` será chamada com:**
```json
{
  "cwd": "/caminho/do/projeto/whmcs",
  "category": "pagamentos",
  "limit": 10
}
```

### Exemplo 4: Busca sem categoria específica

**Prompt para o assistente:**
> "Lembra de algo sobre integração com bancos?"

**A ferramenta `memory_search` será chamada sem filtro de categoria:**
```json
{
  "cwd": "/caminho/do/projeto/whmcs",
  "query": "integração com bancos",
  "limit": 10
}
```

### Como funciona

1. **Fale naturalmente** com o assistente sobre o que você quer salvar ou buscar
2. **O assistente identifica** automaticamente:
   - O diretório do projeto (`cwd`)
   - A categoria relevante (se mencionada)
   - O conteúdo principal da memória
3. **As ferramentas são chamadas** com os parâmetros apropriados
4. **Os resultados são apresentados** em linguagem natural

### Formas simples de pedir

**Para salvar:**
- "Salve isso: [informação]"
- "Guarde isso: [informação]" 
- "Memorize: [informação]"
- "Lembre disso: [informação]"
- "Salva essa info: [informação]"
- "Memorize o contexto atual: [informação]"  (util para documentar o que está fazendo)

**Exemplos práticos com "contexto atual":**
- "Memorize o contexto atual: Estou implementando o login OAuth2 no módulo auth"
- "Lembre disso: Nesta branch, adicionamos suporte a PostgreSQL"
- "Guarde isso: O problema estava no cache do Redis, resolvido com `FLUSHALL`"
- "Memorize: Debug foi feito no arquivo user.go, linha 42"

**Para buscar:**
- "Lembra de algo sobre [assunto]?"
- "Procure por informações sobre [assunto]"
- "Tem algo sobre [tema]?"
- "Busca por [assunto] no memoos"

**Para listar:**
- "Mostra tudo que você guardou"
- "Quais informações você tem?"
- "Lista as memórias recentes"
- "O que você lembra sobre [categoria]?"

### Dicas de uso

- **Seja específico** ao descrever o conteúdo
- **Use categorias** para organizar melhor as informações
- **Mencione o projeto** se estiver trabalhando em múltiplos locais
- **O contexto atual** é automaticamente detectado com base no `cwd`

```json
{
  "cwd": "/home/user/projects/whmcs",
  "category": "payments",
  "limit": 10
}
```

## Desenvolvimento

```bash
# Instalar dependências
go mod tidy

# Executar testes
make test

# Executar com coverage
make test-cov

# Build
make build

# Executar servidor MCP
make run

# Executar CLI
make run-cli ARGS="--help"
```

## Licença

MIT