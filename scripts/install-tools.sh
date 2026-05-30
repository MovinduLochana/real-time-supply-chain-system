#!/bin/bash

# install-tools.sh - Installs all required development tools
# Run this once to set up your development environment

set -e

echo "RTSCS Development Tools Installer"
echo "===================================="
echo ""

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
check_tool() {
    if command -v $1 &> /dev/null; then
        echo -e "${GREEN}✓${NC} $1 is installed"
        return 0
    else
        echo -e "${RED}✗${NC} $1 is not installed"
        return 1
    fi
}

install_tool() {
    echo -e "${YELLOW}Installing $1...${NC}"
    eval "$2"
    echo -e "${GREEN}✓ $1 installed${NC}"
}

# Check/Install Docker
echo "Checking Docker..."
if check_tool docker; then
    :
else
    echo "Please install Docker Desktop from https://www.docker.com/products/docker-desktop"
    exit 1
fi

# Check/Install kubectl
echo ""
echo "Checking kubectl..."
if ! check_tool kubectl; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        install_tool "kubectl" "brew install kubectl"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        install_tool "kubectl" "curl -LO 'https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl' && chmod +x kubectl && sudo mv kubectl /usr/local/bin/"
    fi
fi

# Check/Install Kind
echo ""
echo "Checking Kind..."
if ! check_tool kind; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        install_tool "kind" "brew install kind"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        install_tool "kind" "GO111MODULE='on' go install sigs.k8s.io/kind@latest && sudo mv $HOME/go/bin/kind /usr/local/bin/"
    fi
fi

# Check/Install Helm
echo ""
echo "Checking Helm..."
if ! check_tool helm; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        install_tool "helm" "brew install helm"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        install_tool "helm" "curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash"
    fi
fi

# Check/Install Terraform
echo ""
echo "Checking Terraform..."
if ! check_tool terraform; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        install_tool "terraform" "brew install terraform"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        install_tool "terraform" "curl https://apt.releases.hashicorp.com/gpg | sudo apt-key add - && sudo apt-add-repository 'deb [arch=amd64] https://apt.releases.hashicorp.com $(lsb_release -cs) main' && sudo apt-get update && sudo apt-get install terraform"
    fi
fi

# Check/Install Buf
echo ""
echo "Checking Buf..."
if ! check_tool buf; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        install_tool "buf" "brew install buf"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        install_tool "buf" "curl -sSL 'https://github.com/bufbuild/buf/releases/download/v1.29.0/buf-Linux-x86_64' -o /tmp/buf && chmod +x /tmp/buf && sudo mv /tmp/buf /usr/local/bin/"
    fi
fi

# Check/Install protoc
echo ""
echo "Checking protoc..."
if ! check_tool protoc; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        install_tool "protoc" "brew install protobuf"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        install_tool "protoc" "apt-get install -y protobuf-compiler"
    fi
fi

# Check language-specific tools
echo ""
echo "Checking language-specific tools..."

# Java
if ! check_tool java; then
    echo -e "${YELLOW}Please install Java 21+${NC}"
fi

# Go
if ! check_tool go; then
    echo -e "${YELLOW}Please install Go 1.22+${NC}"
fi

# Rust
if ! check_tool cargo; then
    echo -e "${YELLOW}Please install Rust (rustup)${NC}"
fi

# Python
if ! check_tool python3; then
    echo -e "${YELLOW}Please install Python 3.11+${NC}"
fi

# Node/Bun
if ! check_tool bun; then
    echo -e "${YELLOW}Installing Bun...${NC}"
    curl -fsSL https://bun.sh/install | bash
    export PATH="$HOME/.bun/bin:$PATH"
    check_tool bun
fi

echo ""
echo -e "${GREEN}===== Installation Complete =====${NC}"
echo ""
echo "Next steps:"
echo "1. Create a Kind cluster: make setup-kind"
echo "2. Deploy infrastructure: make deploy-infra"
echo "3. Generate proto code: bash scripts/gen-proto.sh"
echo ""
