import os

import requests
from dotenv import dotenv_values


def create_vars_env(validators_count, beacon_nodes_count, data_dir='.lighthouse/local-testnet',
                    validator_clients_count=0, seconds_per_slot=3, seconds_per_eth1_block=1):
    vars_env = {
        "DATADIR": data_dir,
        "BN_COUNT": beacon_nodes_count,
        "VALIDATOR_COUNT": validators_count,
        # "GENESIS_VALIDATOR_COUNT": validators_count,
        "VC_COUNT": validator_clients_count,
        "SECONDS_PER_SLOT": seconds_per_slot,
        "SECONDS_PER_ETH1_BLOCK": seconds_per_eth1_block,
    }
    res = requests.get('https://raw.githubusercontent.com/sigp/lighthouse/unstable/scripts/local_testnet/vars.env')
    with open("vars.env", 'wb')as file:
        file.write(res.content)
    remote_vars_env = dotenv_values("vars.env")
    print(f"remote_vars_env={remote_vars_env}")
    with open("vars.env", "w") as f:
        for k, v in remote_vars_env.items():
            if k in vars_env.keys():
                v = vars_env[k]
                print(f"k={k},v={v}")
            f.write(f"{k}={v}\n")


def setup_lighthouse(validators_count, beacon_nodes_count, main_dir, data_dir='.lighthouse/local-testnet',
                     validator_clients_count=0, seconds_per_slot=3, seconds_per_eth1_block=1):
    os.chdir(main_dir)
    lh_dir = os.path.join(main_dir, 'lighthouse')
    if not os.path.isdir(lh_dir):
        os.mkdir(lh_dir)
    os.chdir(lh_dir)
    print("creating vars.env")
    create_vars_env(validators_count, beacon_nodes_count, data_dir, validator_clients_count, seconds_per_slot,
                    seconds_per_eth1_block)
    # TODO: build a docker image with the new vars.env
    return True


def start_lighthouse():
    ## TODO: implement
    return False
