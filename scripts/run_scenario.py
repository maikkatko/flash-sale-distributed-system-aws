import time
import os
import sys
import subprocess
import json

import yaml

from dotenv import load_dotenv

load_dotenv()

def run_scenario(scenario_name: str):
    with open('scenarios.yaml', 'r', encoding='utf-8') as f:
        scenarios = yaml.safe_load(f)

    if scenario_name not in scenarios:
        print(f"Scenario {scenario_name} not found!")
        return

    config = scenarios[scenario_name]

    env = os.environ.copy()
    env['API_HOST'] = os.getenv('ALB_DNS_NAME', '')
    env['WORKERS'] = str(config.get('workers', 1))
    env['USERS'] = str(config.get('users', 1))
    env['SPAWN_RATE'] = str(config.get('spawn_rate', 1))
    env['RUN_TIME'] = str(config.get('run_time', 60))
    env['GET_WEIGHT'] = str(config.get('get_weight', 1))
    env['POST_WEIGHT'] = str(config.get('post_weight', 1))
    env['TEST_NAME'] = scenario_name
    env['LOCUSTFILE'] = 'locustfile.py'
    env['SERVICE_NAME'] = os.getenv('ECS_SERVICE', '')
    env['TEST_RESULTS_FILE_NAME'] = f"{scenario_name}_test_results.json"

    print(f"\n{'='*60}")
    print(f"Running {scenario_name} scenario:")
    print(f"{'='*60}\n")

    subprocess.run(['docker-compose', 'up', '-d', '--build'],
                   check=False, env=env)

    # Wait for test to complete
    run_time = config.get('run_time', 60)

    print(f"Test running for {run_time} seconds...")

    try:
        time.sleep(run_time + 5)
    except (subprocess.CalledProcessError, KeyboardInterrupt):
        print("\nStopping test...")
        sys.exit(1)
    finally:
        subprocess.run(['docker-compose', 'down'], check=False, env=env)
        print(f"Test complete! Results saved to {config.get('test_file')}")


if __name__ == "__main__":
    run_scenario(sys.argv[1])
