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

Para embeddings locais:
```bash
make download-libs  # Baixa bibliotecas ONNX
```

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

O servidor MCP expõe três ferramentas principais:

### 1. `memory_save`

Salva uma memória semanticamente.

```json
{
  "cwd": "/home/user/projects/whmcs",
  "category": "payments",
  "content": "Refund Pix reutiliza e2eid",
  "metadata": {
    "autor": "joao",
    "prioridade": "alta"
  }
}
```

**Parâmetros:**
- `cwd` (requerido): Diretório do projeto para resolução de contexto
- `content` (requerido): Conteúdo da memória a ser salvo
- `category` (opcional): Categoria para separação de contexto
- `metadata` (opcional): Metadados em pares chave-valor

### 2. `memory_search`

Busca memórias por similaridade semântica.

```json
{
  "cwd": "/home/user/projects/whmcs",
  "query": "como funciona estorno pix?",
  "category": "payments",
  "limit": 10,
  "min_score": 0.5,
  "max_distance": 0.8
}
```

**Parâmetros:**
- `cwd` (requerido): Diretório do projeto
- `query` (requerido): Texto para busca semântica
- `limit` (opcional): Máximo de resultados (padrão: 10, máximo: 100)
- `min_score` (opcional): Pontuação mínima de similaridade (0.0-1.0, padrão: 0.5)
- `max_distance` (opcional): Distância euclidiana máxima
- `category` (opcional): Filtrar por categoria

### 3. `memory_list`

Lista memórias recentes.

```json
{
  "cwd": "/home/user/projects/whmcs",
  "category": "payments",
  "limit": 20,
  "offset": 0
}
```

**Parâmetros:**
- `cwd` (requerido): Diretório do projeto
- `category` (opcional): Filtrar por categoria
- `limit` (opcional): Máximo de resultados (padrão: 20, máximo: 100)
- `offset` (opcional): Offset de paginação (padrão: 0)

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