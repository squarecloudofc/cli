#!/usr/bin/env sh
set -euo pipefail

# Verifica se o sistema operacional é Windows.
# Se for Windows, informa o usuário para desinstalar no WSL.
if [[ ${OS:-} = Windows_NT ]]; then
    echo 'Erro: Se você é um usuário Windows, por favor, desinstale no WSL (Windows Subsystem for Linux).'
    exit 1
fi

# Define variáveis para cores no terminal para melhor legibilidade da saída.
BOLD="$(tput bold 2>/dev/null || printf '')"
GREY="$(tput setaf 0 2>/dev/null || printf '')"
RED="$(tput setaf 1 2>/dev/null || printf '')"
GREEN="$(tput setaf 2 2>/dev/null || printf '')"
YELLOW="$(tput setaf 3 2>/dev/null || printf '')"
NO_COLOR="$(tput sgr0 2>/dev/null || printf '')"

# Função para exibir mensagens informativas.
info() {
  printf '%s\n' "${BOLD}${GREY}>${NO_COLOR} $*"
}

# Função para exibir mensagens de aviso.
warn() {
  printf '%s\n' "${YELLOW}! $*${NO_COLOR}"
}

# Função para exibir mensagens de erro e sair.
error() {
  printf '%s\n' "${RED}x $*${NO_COLOR}" >&2
  exit 1
}

# Função para exibir mensagens de sucesso.
success() {
  printf '%s\n' "${GREEN}$@ ${NO_COLOR}"
}

# Função para exibir mensagens de conclusão com um visto verde.
completed() {
  printf '%s\n' "${GREEN}✓${NO_COLOR} $*"
}

# Função para verificar se um comando existe no PATH.
has() {
  command -v "$1" 1>/dev/null 2>&1
}

# Define o caminho onde o executável squarecloud é geralmente instalado.
INSTALL_PATH="/usr/local/bin/squarecloud"

info "Iniciando a desinstalação da CLI do Square Cloud..."

# Verifica se o executável existe no caminho de instalação padrão.
if [ ! -f "${INSTALL_PATH}" ]; then
    warn "O executável 'squarecloud' não foi encontrado em '${INSTALL_PATH}'."
    warn "A CLI do Square Cloud pode não estar instalada ou está em um local diferente."
    completed "Desinstalação concluída (nenhum arquivo para remover)."
    exit 0
fi

info "Encontrado 'squarecloud' em '${INSTALL_PATH}'."

# Tenta remover o executável. Requer sudo.
if has sudo; then
    info "Removendo '${INSTALL_PATH}' (requer sudo)..."
    if sudo rm -f "${INSTALL_PATH}"; then
        completed "CLI do Square Cloud desinstalada com sucesso!"
    else
        error "Falha ao remover '${INSTALL_PATH}'. Verifique suas permissões."
    fi
else
    error "sudo não encontrado. Por favor, execute este script com privilégios de root (ex: sudo sh uninstall.sh) ou remova manualmente '${INSTALL_PATH}'."
fi
