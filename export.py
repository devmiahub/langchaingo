import os
import shutil # Não é mais usado, mas vou deixar generate_project_structure_text que pode evoluir
import argparse

# ============================================================================
# CONFIGURAÇÃO DE EXCEÇÕES PERSONALIZADAS
# ============================================================================
# Adicione aqui os arquivos e pastas específicos que você quer ignorar
CUSTOM_IGNORE_FILES = {
    # # 'curl.go',  # Exemplo: ignora o arquivo curl.go
    # # 'data.go',
    # 'tools_utilitarias.go',
    # 'tools_imoveis.go',
    # "dynamic_filters.go",
    # # "tools_imoveis.go",
    # # "tools_chatwoot.go",
    # # "webhook_external.go",

}

CUSTOM_IGNORE_DIRS = {
    'vectorstores',  # Exemplo: ignora a pasta pkg/chatbotai
    'util',
    'tools',
    'textsplitter',
    'testing',
    'outputparser',
    'httputil',
    'examples',
    'embeddings',
    'documentloaders',
    'docs',
    'llms/openai'
    'llms/bedrock',
    'llms/cloudflare',
    'llms/cohere',
    'llms/compliance',
    'llms/ernie',
    'llms/fake',
    'llms/huggingface',
    'llms/llamafile',
    'llms/local',
    'llms/maritaca',
    'llms/mistral',
    'llms/ollama',
    'llms/watsonx',
}

# ============================================================================
# CONFIGURAÇÃO DE PASTA ESPECÍFICA (NOVA OPÇÃO)
# ============================================================================
# Se você quiser exportar apenas uma pasta específica, defina o caminho aqui.
# Deixe como None ou string vazia para exportar o projeto completo.
CUSTOM_FOLDER_PATH = None  # Exemplo: 'pkg/agents' ou 'internal/handlers'

# Função auxiliar para verificar se um caminho deve ser ignorado
def should_ignore_path(relative_path, custom_ignore_dirs):
    """
    Verifica se um caminho relativo deve ser ignorado baseado nas exceções personalizadas.
    """
    normalized_path = relative_path.replace(os.sep, '/')
    for ignore_dir in custom_ignore_dirs:
        if normalized_path.startswith(ignore_dir + '/') or normalized_path == ignore_dir:
            return True
    return False

def generate_project_structure_text(source_folder, custom_ignore_files=None, custom_ignore_dirs=None):
    """
    Gera uma representação textual da estrutura de diretórios e arquivos .go.
    """
    if custom_ignore_files is None:
        custom_ignore_files = set()
    if custom_ignore_dirs is None:
        custom_ignore_dirs = set()
        
    structure_lines = ["Estrutura do Projeto e Arquivos .go Encontrados:\n"]
    
    # Adiciona informações sobre exceções se houver
    if custom_ignore_files or custom_ignore_dirs:
        structure_lines.append("\n--- EXCEÇÕES APLICADAS ---\n")
        if custom_ignore_files:
            structure_lines.append(f"Arquivos ignorados: {', '.join(sorted(custom_ignore_files))}\n")
        if custom_ignore_dirs:
            structure_lines.append(f"Pastas ignoradas: {', '.join(sorted(custom_ignore_dirs))}\n")
        structure_lines.append("-" * 30 + "\n\n")
    
    base_level = source_folder.count(os.sep)
    # Pastas e arquivos comuns a ignorar na listagem da estrutura
    ignore_items = {'.git', '.vscode', 'node_modules', '__pycache__', '.idea', 'vendor', 'go.mod', 'go.sum'} 

    for root, dirs, files in os.walk(source_folder, topdown=True):
        relative_root = os.path.relpath(root, source_folder)
        
        # Verifica se o diretório atual deve ser ignorado pelas exceções personalizadas
        if relative_root != "." and should_ignore_path(relative_root, custom_ignore_dirs):
            dirs.clear()  # Para não processar subdiretórios
            continue
        
        # Modifica dirs IN-PLACE para pular os diretórios ignorados
        dirs[:] = [d for d in dirs if d not in ignore_items and not d.startswith('.')]
        
        # Filtra também pelos diretórios personalizados ignorados
        dirs[:] = [d for d in dirs if not should_ignore_path(
            os.path.relpath(os.path.join(root, d), source_folder), custom_ignore_dirs
        )]
        
        # Remove arquivos ignorados da listagem de arquivos
        current_files = [f for f in files if f not in ignore_items and not f.startswith('.')]
        # Remove arquivos personalizados ignorados
        current_files = [f for f in current_files if f not in custom_ignore_files]

        # Ignora o processamento se a raiz atual estiver dentro de uma pasta ignorada
        path_parts = root.split(os.sep)
        if any(ignored_part in path_parts for ignored_part in ignore_items):
            continue
        if any(part.startswith('.') for part in path_parts if part != os.path.basename(source_folder) and part != '.' and part != '..'):
            continue
        
        level = root.count(os.sep) - base_level
        indent = '    ' * level
        
        if relative_root == ".":
            structure_lines.append(f"{os.path.basename(os.path.abspath(source_folder))}/\n")
        else:
            # Apenas mostra o diretório se ele contém arquivos .go ou subdiretórios que podem conter
            go_files_in_dir_or_subdir = any(f.endswith(".go") for f in current_files) or \
                                        any(os.path.exists(os.path.join(root, d)) and \
                                            any(f.endswith(".go") for _, _, s_files in os.walk(os.path.join(root, d)) for f in s_files) \
                                            for d in dirs)
            if go_files_in_dir_or_subdir or any(f.endswith(".go") for f in current_files):
                 structure_lines.append(f"{indent}└── {os.path.basename(root)}/\n")
            else:
                continue
        
        sub_indent = '    ' * (level + 1)
        
        # Lista apenas arquivos .go
        go_files = sorted([f for f in current_files if f.endswith(".go")])
        for i, f_name in enumerate(go_files):
            prefix = "└──" if i == len(go_files) - 1 else "├──"
            structure_lines.append(f"{sub_indent}{prefix} {f_name}\n")
                
    structure_lines.append("\n" + "="*80 + "\n\n")
    return "".join(structure_lines)

def export_go_code_to_single_txt(source_folder, output_txt_file, custom_ignore_files=None, custom_ignore_dirs=None):
    """
    Exporta a estrutura de pastas e o conteúdo de todos os arquivos .go
    de uma pasta de origem para um único arquivo .txt.

    Args:
        source_folder (str): O caminho para a pasta de origem (raiz do projeto Go).
        output_txt_file (str): O caminho para o arquivo .txt de saída.
        custom_ignore_files (set): Conjunto de nomes de arquivos para ignorar.
        custom_ignore_dirs (set): Conjunto de caminhos de diretórios para ignorar.
    """
    if custom_ignore_files is None:
        custom_ignore_files = set()
    if custom_ignore_dirs is None:
        custom_ignore_dirs = set()
        
    if not os.path.isdir(source_folder):
        print(f"Erro: A pasta de origem '{source_folder}' não existe ou não é um diretório.")
        return

    # Garante que o diretório do arquivo de saída exista
    output_dir = os.path.dirname(output_txt_file)
    if output_dir and not os.path.exists(output_dir):
        try:
            os.makedirs(output_dir)
            print(f"Diretório de saída '{output_dir}' criado.")
        except OSError as e:
            print(f"Erro ao criar diretório de saída '{output_dir}': {e}")
            return

    found_go_files = False
    ignored_files_count = 0
    ignored_dirs_count = 0

    try:
        with open(output_txt_file, 'w', encoding='utf-8') as outfile:
            print(f"Gerando estrutura de pastas para '{source_folder}'...")
            
            # Mostra as exceções que serão aplicadas
            if custom_ignore_files or custom_ignore_dirs:
                print("\n--- EXCEÇÕES APLICADAS ---")
                if custom_ignore_files:
                    print(f"Arquivos ignorados: {', '.join(sorted(custom_ignore_files))}")
                if custom_ignore_dirs:
                    print(f"Pastas ignoradas: {', '.join(sorted(custom_ignore_dirs))}")
                print("-" * 30)
            
            structure_text = generate_project_structure_text(source_folder, custom_ignore_files, custom_ignore_dirs)
            outfile.write(structure_text)
            print(f"Estrutura de pastas escrita em '{os.path.basename(output_txt_file)}'.")

            print(f"\nProcurando e exportando arquivos .go de '{source_folder}'...\n")
            
            # Pastas comuns a ignorar na busca por arquivos .go
            ignore_dirs_for_content = {'.git', '.vscode', 'node_modules', '__pycache__', '.idea', 'vendor'}

            for root, dirs, files in os.walk(source_folder):
                relative_root = os.path.relpath(root, source_folder)
                
                # Verifica se o diretório atual deve ser ignorado pelas exceções personalizadas
                if relative_root != "." and should_ignore_path(relative_root, custom_ignore_dirs):
                    ignored_dirs_count += 1
                    dirs.clear()  # Para não processar subdiretórios
                    continue
                
                # Modifica dirs IN-PLACE para pular os diretórios ignorados
                dirs[:] = [d for d in dirs if d not in ignore_dirs_for_content and not d.startswith('.')]
                
                # Filtra também pelos diretórios personalizados ignorados
                dirs[:] = [d for d in dirs if not should_ignore_path(
                    os.path.relpath(os.path.join(root, d), source_folder), custom_ignore_dirs
                )]
                
                # Ignora o processamento de arquivos se a raiz atual estiver dentro de uma pasta ignorada
                path_parts = root.split(os.sep)
                if any(ignored_part in path_parts for ignored_part in ignore_dirs_for_content):
                    continue
                if any(part.startswith('.') for part in path_parts if part != os.path.basename(source_folder) and part != '.' and part != '..'):
                     continue

                for filename in sorted(files):
                    if filename.endswith(".go"):
                        # Verifica se o arquivo deve ser ignorado pelas exceções personalizadas
                        if filename in custom_ignore_files:
                            ignored_files_count += 1
                            print(f"Ignorando arquivo: {os.path.relpath(os.path.join(root, filename), source_folder)}")
                            continue
                            
                        found_go_files = True
                        file_path = os.path.join(root, filename)
                        relative_path = os.path.relpath(file_path, source_folder)

                        header = f"--- Arquivo: {relative_path.replace(os.sep, '/')} ---\n"
                        print(f"Adicionando: {relative_path}")

                        outfile.write(header)
                        outfile.write("```go\n")

                        try:
                            with open(file_path, 'r', encoding='utf-8', errors='ignore') as infile:
                                outfile.write(infile.read())
                        except Exception as e:
                            outfile.write(f"\n[ERRO AO LER ARQUIVO: {e}]\n")
                        
                        outfile.write("\n```\n")
                        outfile.write("\n" + "="*80 + "\n\n")

            # Relatório final
            print(f"\n--- RELATÓRIO FINAL ---")
            if found_go_files:
                print(f"✅ Código Go exportado para '{os.path.abspath(output_txt_file)}'.")
            else:
                message = f"⚠️  Nenhum arquivo .go encontrado em '{source_folder}' (após aplicar filtros e exceções)."
                print(message)
                outfile.write(message)
            
            if ignored_files_count > 0:
                print(f"📄 Arquivos .go ignorados: {ignored_files_count}")
            if ignored_dirs_count > 0:
                print(f"📁 Diretórios ignorados: {ignored_dirs_count}")

    except IOError as e:
        print(f"Erro de E/S ao manusear o arquivo '{output_txt_file}': {e}")
    except Exception as e:
        print(f"Ocorreu um erro inesperado: {e}")

def export_single_folder_to_txt(folder_path, output_txt_file, custom_ignore_files=None):
    """
    Exporta apenas uma pasta específica (sem subpastas) e o conteúdo de todos os arquivos .go
    dessa pasta para um único arquivo .txt.

    Args:
        folder_path (str): O caminho para a pasta específica a ser exportada.
        output_txt_file (str): O caminho para o arquivo .txt de saída.
        custom_ignore_files (set): Conjunto de nomes de arquivos para ignorar.
    """
    if custom_ignore_files is None:
        custom_ignore_files = set()
        
    if not os.path.isdir(folder_path):
        print(f"Erro: A pasta '{folder_path}' não existe ou não é um diretório.")
        return

    # Garante que o diretório do arquivo de saída exista
    output_dir = os.path.dirname(output_txt_file)
    if output_dir and not os.path.exists(output_dir):
        try:
            os.makedirs(output_dir)
            print(f"Diretório de saída '{output_dir}' criado.")
        except OSError as e:
            print(f"Erro ao criar diretório de saída '{output_dir}': {e}")
            return

    found_go_files = False
    ignored_files_count = 0

    try:
        with open(output_txt_file, 'w', encoding='utf-8') as outfile:
            folder_name = os.path.basename(os.path.abspath(folder_path))
            
            # Cabeçalho
            outfile.write(f"Exportação da Pasta: {folder_name}\n")
            outfile.write(f"Caminho: {os.path.abspath(folder_path)}\n")
            
            # Mostra as exceções que serão aplicadas
            if custom_ignore_files:
                outfile.write(f"\n--- EXCEÇÕES APLICADAS ---\n")
                outfile.write(f"Arquivos ignorados: {', '.join(sorted(custom_ignore_files))}\n")
                outfile.write("-" * 30 + "\n\n")
                print(f"Arquivos ignorados: {', '.join(sorted(custom_ignore_files))}")

            print(f"Exportando arquivos .go da pasta '{folder_path}'...")
            
            # Lista apenas os arquivos da pasta (sem subpastas)
            try:
                files = os.listdir(folder_path)
                go_files = sorted([f for f in files if f.endswith('.go') and os.path.isfile(os.path.join(folder_path, f))])
                
                if not go_files:
                    message = f"⚠️  Nenhum arquivo .go encontrado na pasta '{folder_path}'."
                    print(message)
                    outfile.write(message + "\n")
                    return
                
                # Lista os arquivos encontrados
                outfile.write(f"Arquivos .go encontrados na pasta:\n")
                for go_file in go_files:
                    if go_file not in custom_ignore_files:
                        outfile.write(f"  - {go_file}\n")
                    else:
                        ignored_files_count += 1
                        print(f"Ignorando arquivo: {go_file}")
                
                outfile.write("\n" + "="*80 + "\n\n")
                
                # Processa cada arquivo .go
                for filename in go_files:
                    if filename in custom_ignore_files:
                        continue
                        
                    found_go_files = True
                    file_path = os.path.join(folder_path, filename)

                    header = f"--- Arquivo: {filename} ---\n"
                    print(f"Adicionando: {filename}")

                    outfile.write(header)
                    outfile.write("```go\n")

                    try:
                        with open(file_path, 'r', encoding='utf-8', errors='ignore') as infile:
                            outfile.write(infile.read())
                    except Exception as e:
                        outfile.write(f"\n[ERRO AO LER ARQUIVO: {e}]\n")
                    
                    outfile.write("\n```\n")
                    outfile.write("\n" + "="*80 + "\n\n")
                
            except OSError as e:
                print(f"Erro ao acessar a pasta '{folder_path}': {e}")
                return

            # Relatório final
            print(f"\n--- RELATÓRIO FINAL ---")
            if found_go_files:
                print(f"✅ Código Go da pasta '{folder_name}' exportado para '{os.path.abspath(output_txt_file)}'.")
            else:
                message = f"⚠️  Nenhum arquivo .go válido encontrado na pasta '{folder_path}' (após aplicar filtros)."
                print(message)
                outfile.write(message)
            
            if ignored_files_count > 0:
                print(f"📄 Arquivos .go ignorados: {ignored_files_count}")

    except IOError as e:
        print(f"Erro de E/S ao manusear o arquivo '{output_txt_file}': {e}")
    except Exception as e:
        print(f"Ocorreu um erro inesperado: {e}")

def main():
    parser = argparse.ArgumentParser(
        description="Exporta código Go para arquivo texto",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Exemplos de uso:

1. Exportar todo o projeto (comportamento padrão):
   python export.py

2. Exportar apenas uma pasta específica:
   python export.py --folder pkg/agents

3. Exportar pasta específica com nome de arquivo personalizado:
   python export.py --folder pkg/agents --output minha_pasta.txt

4. Exportar todo o projeto com nome personalizado:
   python export.py --output projeto_completo.txt
        """
    )
    
    parser.add_argument(
        '--folder', '-f',
        type=str,
        help='Caminho da pasta específica para exportar (relativo ao diretório atual). Se não especificado, exporta todo o projeto.'
    )
    
    parser.add_argument(
        '--output', '-o',
        type=str,
        help='Nome do arquivo de saída (padrão: baseado no nome da pasta/projeto)'
    )
    
    args = parser.parse_args()
    
    # Verifica se foi configurada uma pasta específica
    if CUSTOM_FOLDER_PATH and CUSTOM_FOLDER_PATH.strip():
        # Modo: exportar apenas uma pasta específica
        current_dir = os.getcwd()
        target_folder = os.path.join(current_dir, CUSTOM_FOLDER_PATH.strip())
        
        if not os.path.isdir(target_folder):
            print(f"Erro: A pasta configurada '{CUSTOM_FOLDER_PATH}' não existe no diretório atual '{current_dir}'.")
            print(f"Caminho completo testado: '{target_folder}'")
            exit(1)
        
        # Define o nome do arquivo de saída baseado na pasta
        folder_name = os.path.basename(CUSTOM_FOLDER_PATH.strip().rstrip(os.sep))
        output_filename = f"{folder_name}_codigo.txt"
        output_file_path = os.path.join(current_dir, output_filename)
        
        print(f"🎯 MODO: Exportar pasta específica (configurada)")
        print(f"📁 Pasta configurada: '{CUSTOM_FOLDER_PATH}'")
        print(f"📍 Pasta de origem: '{target_folder}'")
        print(f"📄 Arquivo de saída: '{output_file_path}'")
        print("-" * 60)
        
        # Exporta apenas a pasta específica
        export_single_folder_to_txt(target_folder, output_file_path, CUSTOM_IGNORE_FILES)
        
    else:
        # Modo: exportar todo o projeto (comportamento original)
        source_folder = os.getcwd()
        project_name = os.path.basename(source_folder)
        output_filename = f"{project_name}_codigo_completo.txt"
        output_file_path = os.path.join(source_folder, output_filename)
        
        print(f"🚀 MODO: Exportar projeto completo")
        print(f"📁 Pasta de origem: '{source_folder}'")
        print(f"📄 Arquivo de saída: '{output_file_path}'")
        print("-" * 60)
        
        # Exporta todo o projeto
        export_go_code_to_single_txt(source_folder, output_file_path, CUSTOM_IGNORE_FILES, CUSTOM_IGNORE_DIRS)

if __name__ == "__main__":
    main()