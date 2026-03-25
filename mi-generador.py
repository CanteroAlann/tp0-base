import sys
import yaml
import os

def generate_compose(filename, count):
    current_uid = os.getuid()
    current_gid = os.getgid()

    if not os.path.exists("data"):
        os.makedirs("data")

    PATH_STORAGE = "./server/bets.csv"
    if not os.path.exists(PATH_STORAGE):
        open(PATH_STORAGE, 'a').close()

    


    compose_data = {
        "name": "tp0",
        "services": {
            "server": {
                "container_name": "server",
                "image": "server:latest",
                "user": f"{current_uid}:{current_gid}",
                "entrypoint": "python3 /main.py",
                "environment": [
                    "PYTHONUNBUFFERED=1"
                ],
                "networks": ["testing_net"],
                "volumes": [
                    "./server/config.ini:/config.ini:ro",
                    f"{PATH_STORAGE}:/bets.csv:rw"
                ]
            }
        },
        "networks": {
            "testing_net": {
                "ipam": {
                    "driver": "default",
                    "config": [
                        {"subnet": "172.25.125.0/24"}
                    ]
                }
            }
        }
    }

    for i in range(1, int(count) + 1):
        name = f"client{i}"
        compose_data["services"][name] = {
            "image": "client:latest",
            "container_name": name,
            "env_file": f"./envs/client{i}.env",
            "entrypoint": "/client",
            "user": f"{current_uid}:{current_gid}",
            "networks": ["testing_net"],
            "environment": [
                f"CLI_ID={i}",
                "DATA_PATH=/data/agency.csv"
            ],
            "volumes": [
                f"./data/{name}:/app/data",
                "./client/config.yaml:/config.yaml:ro",
                f"./.data/agency-{i}.csv:/data/agency.csv:ro"
            ],
            "depends_on": ["server"]
        }

    with open(filename, 'w') as f:
        yaml.dump(compose_data, f, default_flow_style=False, sort_keys=False)

if __name__ == "__main__":
    if len(sys.argv) < 3:
        print("Faltan argumentos")
        sys.exit(1)
    generate_compose(sys.argv[1], sys.argv[2])