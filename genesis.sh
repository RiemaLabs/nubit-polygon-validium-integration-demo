#!/bin/bash

work=$(pwd)
function generate_genesis() {
    check_jq
    deploy_output=${work}/contracts/docker/deploymentOutput/deploy_output.json
    genesis=${work}/contracts/docker/deploymentOutput/genesis.json
    create_rollup_output=${work}/contracts/docker/deploymentOutput/create_rollup_output.json

#     l1Config.polygonZkEVMAddress ==> rollupAddress @ create_rollup_output.json
#     l1Config.polygonRollupManagerAddress ==> polygonRollupManager @ deploy_output.json
#     l1Config.polTokenAddress ==> polTokenAddress @ deploy_output.json
#     l1Config.polygonZkEVMGlobalExitRootAddress ==> polygonZkEVMGlobalExitRootAddress @ deploy_output.json
#     rollupCreationBlockNumber ==> createRollupBlock @ create_rollup_output.json
#     rollupManagerCreationBlockNumber ==> deploymentBlockNumber @ deploy_output.json
#     root ==> root @ genesis.json
#     genesis ==> genesis @ genesis.json

    rollupAddress=$(cat "$create_rollup_output" | jq -r '.rollupAddress')
    createRollupBlock=$(cat "$create_rollup_output" | jq -r '.createRollupBlockNumber')


    polygonRollupManager=$(cat "$deploy_output" | jq -r '.polygonRollupManagerAddress')
    echo "polygonRollupManager: ${polygonRollupManager}"
    polTokenAddress=$(cat "$deploy_output" | jq -r '.polTokenAddress')
    polygonZkEVMGlobalExitRootAddress=$(cat "$deploy_output" | jq -r '.polygonZkEVMGlobalExitRootAddress')
    deploymentBlockNumber=$(cat "$deploy_output" | jq -r '.deploymentRollupManagerBlockNumber')

    root=$(cat "$genesis" | jq -r '.root')
    genesis=$(cat "$genesis" | jq -c '.genesis')

    update_genesis_json_number 'rollupCreationBlockNumber' "$createRollupBlock"
    update_genesis_json 'l1Config.polygonZkEVMAddress' "$rollupAddress"

    update_genesis_json 'l1Config.polygonRollupManagerAddress' "$polygonRollupManager"
    update_genesis_json 'l1Config.polTokenAddress' "$polTokenAddress"
    update_genesis_json 'l1Config.polygonZkEVMGlobalExitRootAddress' "$polygonZkEVMGlobalExitRootAddress"

    update_genesis_json_number 'rollupManagerCreationBlockNumber' "$deploymentBlockNumber"
    update_genesis_json 'root' "$root"
    update_genesis_json_obj 'genesis' "$genesis"

    sleep 3

    echo "Please enter the nubit rpcURL:"
    read rpcURL
    echo "Please enter the new modularAppName:"
    read modularAppName
    echo "Please enter the new authKey:"
    read authKey

    echo "$rpcURL"
    echo "$modularAppName"
    echo "$authKey"

    if [ -n "$rpcURL" ]; then
        update_nubit_config 'rpcURL' "$rpcURL"
    else
        echo "rpcURL cannot be empty."
    fi

    sleep 2

    if [ -n "$modularAppName" ]; then
        update_nubit_config 'modularAppName' "$modularAppName"
    else
        echo "modularAppName cannot be empty."
    fi

    sleep 2

    if [ -n "$authKey" ]; then
        update_nubit_config 'authKey' "$authKey"
    else
        echo "authKey cannot be empty."
    fi

    sleep 3

    cd ./cdk-validium-node/test/ && make run

}

update_nubit_config() {
    local key="$1"
    local value="$2"
    local json_file=${work}/cdk-validium-node/test/config/test.nubit.config.json

    jq --arg key "$key" --arg value "$value" '.[$key] = $value' "$json_file" > tmp.json && \
        mv tmp.json "$json_file" && \
        echo "Value for key '$key' successfully updated to '$value'" || \
        echo "Failed to update the value for key '$key'"
}

update_genesis_json_number() {
    genesis_cfg=${work}/cdk-validium-node/test/config/test.genesis.config.json
    local key="$1"
    local value="$2"
    echo "Updating $key to $value in $genesis_cfg"
    jq --arg val "$value" ".$key = (\$val | tonumber)" "$genesis_cfg" > temp_genesis.json && mv temp_genesis.json "$genesis_cfg"
}

update_genesis_json() {
    genesis_cfg=${work}/cdk-validium-node/test/config/test.genesis.config.json
    local key="$1"
    local value="$2"
    echo "Updating $key to $value in $genesis_cfg"
    jq --arg val "$value" ".$key = \$val" "$genesis_cfg" > temp_genesis.json && mv temp_genesis.json "$genesis_cfg"
}

update_genesis_json_obj() {
    genesis_cfg=${work}/cdk-validium-node/test/config/test.genesis.config.json
    local key="$1"
    local value="$2"
    echo "Updating $key to $value in $genesis_cfg"
    jq --argjson val "$value" ".$key = \$val" "$genesis_cfg" > temp_genesis.json && mv temp_genesis.json "$genesis_cfg"
}

check_jq() {
    if ! command -v jq &> /dev/null; then
        echo "jq is not installed. Installing now..."
        sudo apt-get update

        sudo apt-get install -y jq

        if [ $? -eq 0 ]; then
            echo "jq has been successfully installed."
        else
            echo "Failed to install jq. Please check your internet connection or package repository."
            exit 1
        fi
    else
        echo "jq is already installed."
    fi
}



if [[ "$0" == "$BASH_SOURCE" ]]; then
    generate_genesis
fi
