#!/bin/bash

work=$(pwd)
function generate_genesis() {
    deploy_output=${work}/contracts/docker/deploymentOutput/deploy_output.json
    genesis=${work}/contracts/docker/deploymentOutput/genesis.json
#     create_rollup_output=${work}/contracts/docker/deploymentOutput/create_rollup_output.json

#     l1Config.polygonZkEVMAddress ==> rollupAddress @ create_rollup_output.json
#     l1Config.polygonRollupManagerAddress ==> polygonRollupManager @ deploy_output.json
#     l1Config.polTokenAddress ==> polTokenAddress @ deploy_output.json
#     l1Config.polygonZkEVMGlobalExitRootAddress ==> polygonZkEVMGlobalExitRootAddress @ deploy_output.json
#     rollupCreationBlockNumber ==> createRollupBlock @ create_rollup_output.json
#     rollupManagerCreationBlockNumber ==> deploymentBlockNumber @ deploy_output.json
#     root ==> root @ genesis.json
#     genesis ==> genesis @ genesis.json

    rollupAddress=$(cat "$create_rollup_output" | jq -r '.rollupAddress')
    createRollupBlock=$(cat "$create_rollup_output" | jq -r '.createRollupBlock')


    polygonRollupManager=$(cat "$deploy_output" | jq -r '.polygonRollupManagerAddress')
    echo "polygonRollupManager: ${polygonRollupManager}"
    polTokenAddress=$(cat "$deploy_output" | jq -r '.polTokenAddress')
    polygonZkEVMGlobalExitRootAddress=$(cat "$deploy_output" | jq -r '.polygonZkEVMGlobalExitRootAddress')
    deploymentBlockNumber=$(cat "$deploy_output" | jq -r '.deploymentRollupManagerBlockNumber')

    root=$(cat "$genesis" | jq -r '.root')
    genesis=$(cat "$genesis" | jq -c '.genesis')

    update_genesis_json 'rollupCreationBlockNumber' "$createRollupBlock"
    update_genesis_json 'l1Config.polygonZkEVMAddress' "$rollupAddress"

    update_genesis_json 'l1Config.polygonRollupManagerAddress' "$polygonRollupManager"
    update_genesis_json 'l1Config.polTokenAddress' "$polTokenAddress"
    update_genesis_json 'l1Config.polygonZkEVMGlobalExitRootAddress' "$polygonZkEVMGlobalExitRootAddress"

    update_genesis_json 'rollupManagerCreationBlockNumber' "$deploymentBlockNumber"
    update_genesis_json 'root' "$root"
    update_genesis_json_obj 'genesis' "$genesis"

}

update_genesis_json() {
    genesis_cfg=${work}/validium-node/test/config/test.genesis.config.json
    local key="$1"
    local value="$2"
    echo "Updating $key to $value in $genesis_cfg"
    jq --arg val "$value" ".$key = \$val" "$genesis_cfg" > temp_genesis.json && mv temp_genesis.json "$genesis_cfg"
}

update_genesis_json_obj() {
    genesis_cfg=${work}/validium-node/test/config/test.genesis.config.json
    local key="$1"
    local value="$2"
    echo "Updating $key to $value in $genesis_cfg"
    jq --argjson val "$value" ".$key = \$val" "$genesis_cfg" > temp_genesis.json && mv temp_genesis.json "$genesis_cfg"
}

generate_genesis
