#!/usr/bin/env python3
"""
Script para alternar entre imports locais e remotos do langchaingo.

Uso:
    python switch_imports.py 1  # Usar dependências locais (devmiahub)
    python switch_imports.py 2  # Usar langchain original (tmc)
"""

import os
import re
import sys
from pathlib import Path


# Configurações
OLD_MODULE = "github.com/tmc/langchaingo"
NEW_MODULE = "github.com/devmiahub/langchaingo"
GO_MOD_FILE = "go.mod"


def find_go_files(root_dir="."):
    """Encontra todos os arquivos .go no diretório."""
    go_files = []
    for root, dirs, files in os.walk(root_dir):
        # Ignorar diretórios comuns
        dirs[:] = [d for d in dirs if d not in [".git", "vendor", "node_modules"]]
        
        for file in files:
            if file.endswith(".go"):
                go_files.append(os.path.join(root, file))
    return go_files


def update_imports_in_file(file_path, old_module, new_module):
    """Atualiza imports em um arquivo .go."""
    try:
        with open(file_path, "r", encoding="utf-8") as f:
            content = f.read()
        
        original_content = content
        
        # Padrão para encontrar imports do módulo
        # Captura: import "github.com/tmc/langchaingo/..."
        pattern = rf'({re.escape(old_module)}/[^"]*)'
        
        def replace_import(match):
            full_import = match.group(0)
            # Substitui apenas a parte do módulo
            new_import = full_import.replace(old_module, new_module, 1)
            return new_import
        
        # Substitui em imports individuais
        content = re.sub(
            rf'"{re.escape(old_module)}/([^"]*)"',
            rf'"{new_module}/\1"',
            content
        )
        
        # Substitui em blocos de import
        content = re.sub(
            rf'\({re.escape(old_module)}/([^"]*)\)',
            rf'({new_module}/\1)',
            content
        )
        
        if content != original_content:
            with open(file_path, "w", encoding="utf-8") as f:
                f.write(content)
            return True
        return False
    except Exception as e:
        print(f"Erro ao processar {file_path}: {e}")
        return False


def update_go_mod(option):
    """Atualiza o go.mod conforme a opção escolhida."""
    try:
        with open(GO_MOD_FILE, "r", encoding="utf-8") as f:
            lines = f.readlines()
        
        original_lines = lines.copy()
        content = ''.join(lines)
        
        if option == 1:
            # Opção 1: Usar dependências locais
            # Atualiza o módulo para devmiahub se necessário
            for i, line in enumerate(lines):
                if line.startswith('module '):
                    if OLD_MODULE in line:
                        lines[i] = line.replace(OLD_MODULE, NEW_MODULE)
                    break
            
            # Adiciona replace no final se não existir
            replace_exists = any('replace github.com/tmc/langchaingo' in line for line in lines)
            if not replace_exists:
                # Remove linhas vazias no final
                while lines and lines[-1].strip() == '':
                    lines.pop()
                # Adiciona replace
                lines.append('\n')
                lines.append('replace github.com/tmc/langchaingo => ./\n')
        
        elif option == 2:
            # Opção 2: Usar langchain original
            # Atualiza o módulo de volta para tmc se necessário
            for i, line in enumerate(lines):
                if line.startswith('module '):
                    if NEW_MODULE in line:
                        lines[i] = line.replace(NEW_MODULE, OLD_MODULE)
                    break
            
            # Remove o replace se existir
            new_lines = []
            skip_next = False
            for i, line in enumerate(lines):
                if 'replace github.com/tmc/langchaingo' in line:
                    skip_next = True
                    continue
                if skip_next and line.strip() == '':
                    skip_next = False
                    continue
                new_lines.append(line)
            lines = new_lines
        
        new_content = ''.join(lines)
        if new_content != content:
            with open(GO_MOD_FILE, "w", encoding="utf-8") as f:
                f.write(new_content)
            return True
        return False
    except Exception as e:
        print(f"Erro ao atualizar go.mod: {e}")
        import traceback
        traceback.print_exc()
        return False


def main():
    if len(sys.argv) < 2:
        print("Uso: python switch_imports.py <opção>")
        print("  Opção 1: Usar dependências locais (github.com/devmiahub/langchaingo)")
        print("  Opção 2: Usar langchain original (github.com/tmc/langchaingo)")
        sys.exit(1)
    
    try:
        option = int(sys.argv[1])
    except ValueError:
        print("Erro: Opção deve ser um número (1 ou 2)")
        sys.exit(1)
    
    if option not in [1, 2]:
        print("Erro: Opção deve ser 1 ou 2")
        sys.exit(1)
    
    if option == 1:
        print("🔄 Alternando para dependências LOCAIS (devmiahub)...")
        old_module = OLD_MODULE
        new_module = NEW_MODULE
    else:
        print("🔄 Alternando para langchain ORIGINAL (tmc)...")
        old_module = NEW_MODULE
        new_module = OLD_MODULE
    
    # Encontra todos os arquivos .go
    print("📁 Procurando arquivos .go...")
    go_files = find_go_files()
    print(f"   Encontrados {len(go_files)} arquivos")
    
    # Atualiza imports nos arquivos
    print("✏️  Atualizando imports...")
    updated_files = 0
    for file_path in go_files:
        if update_imports_in_file(file_path, old_module, new_module):
            updated_files += 1
            print(f"   ✓ {file_path}")
    
    print(f"\n   {updated_files} arquivos atualizados")
    
    # Atualiza go.mod
    print("📝 Atualizando go.mod...")
    if update_go_mod(option):
        print("   ✓ go.mod atualizado")
    else:
        print("   - go.mod não precisou de alterações")
    
    print("\n✅ Concluído!")
    print("\n⚠️  Próximos passos:")
    print("   1. Execute: go mod tidy")
    print("   2. Execute: go mod download (se necessário)")


if __name__ == "__main__":
    main()

