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
go install github.com/cristian-scherer/memoos/cmd/memoos-cli@latest
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
# Salvar memória
memoos-cli save \
  --cwd /home/user/projects/whmcs \
  --category payments \
  --content "Refund Pix reutiliza e2eid"

# Buscar memórias
memoos-cli search \
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

### memory_save

Salva uma memória semanticamente.

```json
{
  "cwd": "/home/user/projects/whmcs",
  "category": "payments",
  "content": "Refund Pix reutiliza e2eid",
  "tags": ["pix", "refund"]
}
```

### memory_search

Busca memórias por similaridade semântica.

```json
{
  "cwd": "/home/user/projects/whmcs",
  "query": "como funciona estorno pix?",
  "category": "payments",
  "limit": 5
}
```

### memory_list

Lista memórias recentes.

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