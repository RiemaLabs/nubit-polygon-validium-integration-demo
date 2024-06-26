#!/bin/bash

work=$(pwd)
image_tag="latest"

function check_zkevmnode(){
    cd ${work}
    image_name="zkevm-node"
    image_exists=$(docker image ls --format "{{.Repository}}:{{.Tag}}" | grep "^${image_name}:${image_tag}$")

    if [ -n "${image_exists}" ]; then
        echo "Docker image ${image_name}:${image_tag} exists."
    else
       echo "Docker image ${image_name}:${image_tag} does not exist."

       cd  ./cdk-validium-node && make build-docker
    fi
}



function check_geth_evm(){
    cd ${work}
    name="hermeznetwork/geth-zkevm-contracts"

    geth_exists=$(docker image ls --format "{{.Repository}}:{{.Tag}}" | grep "^${name}:${image_tag}$")

    if [ -n "${geth_exists}" ]; then
        echo "Docker image ${name}:${image_tag} exists."
    else
       echo "Docker image ${name}:${image_tag} does not exist."

       check_node


       cd ./contracts && npm install --save-dev hardhat && npm i && npx hardhat compile && npm run docker:contracts

       sleep 2

       docker image tag hermeznetwork/geth-zkevm-contracts hermeznetwork/geth-cdk-validium-contracts:20240420



    fi
}



check_node() {
    if ! command -v node &> /dev/null; then
        echo "Node.js is not installed. Installing v20..."
        install_node_v20
    else
        echo "Node.js is installed. Checking version..."

        major_version=$(echo "$node_version" | cut -d '.' -f 1)

        if [ "$major_version" != "v20" ]; then
            echo "Node.js version is not v20 (current version: $node_version). Installing v20..."
            install_node_v20
        else
            echo "Node.js version is v20."
        fi
    fi
}

check_nvm() {
    if  ! command -v nvm &> /dev/null; then
         echo "NVM is not installed. Installing NVM..."
         install_nvm
    else
        echo "NVM is installed."
    fi
}

install_nvm() {
    curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.1/install.sh | bash

    # wget -qO- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.1/install.sh | bash

    export NVM_DIR="$HOME/.nvm"
    [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
    [ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"
    sleep 2
    if command -v nvm &> /dev/null; then
        echo "Successfully installed NVM."
    else
        echo "Failed to install NVM. Please check the logs and try again manually."
        exit 1
    fi
}

install_node_v20() {
    check_nvm

    nvm_path=$(command -v nvm)

    echo "nvm_path: ${nvm_path}"
    "${nvm_path}" install 20
}


function run(){
    git clone https://github.com/RiemaLabs/cdk-validium-node.git ./cdk-validium-node
    git clone https://github.com/RiemaLabs/zkevm-contracts.git ./contracts
    check_zkevmnode
    check_geth_evm
}

run
