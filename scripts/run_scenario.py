from pathlib import Path
import time
import os
import sys
import subprocess
import boto3
import yaml
from dotenv import load_dotenv

project_root = Path(__file__).parent.parent
load_dotenv(dotenv_path=project_root / '.env')

def run_scenario(scenario_name: str):
    with open('scenarios.yaml', 'r', encoding='utf-8') as f:
        scenarios = yaml.safe_load(f)

    if scenario_name not in scenarios:
        print(f"Scenario {scenario_name} not found!")
        print(f"Available scenarios: {', '.join(scenarios.keys())}")
        return

    config = scenarios[scenario_name]

    env = os.environ.copy()
    env['API_HOST'] = os.getenv('ALB_DNS_NAME', '')
    env['WORKERS'] = str(config.get('workers', 1))
    env['USERS'] = str(config.get('users', 1))
    env['SPAWN_RATE'] = str(config.get('spawn_rate', 1))
    env['RUN_TIME'] = str(config.get('run_time', 60))
    env['USER_CLASS'] = config.get('user_class', 'NormalUser')
    env['TEST_NAME'] = scenario_name
    env['LOCUSTFILE'] = 'locustfile.py'
    env['SERVICE_NAME'] = os.getenv('SERVICE_NAME', '')
    env['TEST_RESULTS_FILE_NAME'] = f"{scenario_name}_test_results.json"
    env['TF_VAR_db_name'] = os.getenv('TF_VAR_db_name', 'flashsale')
    env['TF_VAR_db_username'] = os.getenv('TF_VAR_db_username', 'admin')
    env['TF_VAR_db_password'] = os.getenv('TF_VAR_db_password', 'SecurePassword123!')

    print(f"\n{'='*60}")
    print(f"Running {scenario_name} scenario:")
    print(f"Description: {str(config.get('description', 'N/A'))})")
    print(f"Scaling Policy: {str(config.get('scaling_policy'))}")
    print(f"User Class: {env['USER_CLASS']}")
    print(f"Users: {env['USERS']}, Spawn Rate: {env['SPAWN_RATE']}")
    print(f"Workers: {env['WORKERS']}, Run Time: {env['RUN_TIME']}s")
    print(f"{'='*60}\n")

    subprocess.run(['docker-compose', 'up', '-d', '--build'],
                check=False, env=env)

    run_time = config.get('run_time', 60)
    print(f"Test running for {run_time} seconds...")

    try:
        time.sleep(run_time + 10)  # Extra buffer for cleanup
    except KeyboardInterrupt:
        print("\nStopping test...")
    finally:
        subprocess.run(['docker-compose', 'down'], check=False, env=env)
        print(f"\nTest complete! Results saved to results/raw_locust_data/{scenario_name}_test_results.json")


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python run_scenario.py <scenario_name>")
        sys.exit(1)
    run_scenario(sys.argv[1])