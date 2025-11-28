# Script de Alternância de Imports

Este script permite alternar facilmente entre usar dependências locais (seu fork) e o langchain original.

## Uso

### Opção 1: Usar dependências LOCAIS (devmiahub)
```bash
python switch_imports.py 1
```

Isso irá:
- Alterar todos os imports de `github.com/tmc/langchaingo` para `github.com/devmiahub/langchaingo`
- Atualizar o `go.mod` para usar o módulo local
- Adicionar `replace github.com/tmc/langchaingo => ./` no `go.mod`

### Opção 2: Usar langchain ORIGINAL (tmc)
```bash
python switch_imports.py 2
```

Isso irá:
- Alterar todos os imports de `github.com/devmiahub/langchaingo` de volta para `github.com/tmc/langchaingo`
- Atualizar o `go.mod` para usar o módulo original
- Remover o `replace` do `go.mod`

## Após executar o script

Sempre execute após a conversão:
```bash
go mod tidy
```

Isso garantirá que todas as dependências estejam corretas.

## Exemplo de fluxo

```bash
# 1. Trabalhar com seu fork local
python switch_imports.py 1
go mod tidy

# ... fazer suas alterações ...

# 2. Voltar para o original (se necessário)
python switch_imports.py 2
go mod tidy
```

## Notas

- O script processa todos os arquivos `.go` no projeto
- Ignora diretórios `.git`, `vendor` e `node_modules`
- Faz backup automático através do controle de versão Git
- Sempre execute `go mod tidy` após usar o script

