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

def _wait_for_ecs_stability(service_name, cluster_name, timeout=300):
    """Wait for ECS service to reach steady state after scaling policy change"""
    ecs = boto3.client('ecs')
    print(f"Waiting for ECS service {service_name} to stabilize...")
    
    waiter = ecs.get_waiter('services_stable')
    
    try:
        waiter.wait(
            cluster=cluster_name,
            services=[service_name],
            WaiterConfig={
                'Delay': 15,
                'MaxAttempts': timeout // 15
            }
        )
        print("ECS service is stable and ready")
    except Exception as e:
        print(f"Warning: Service stabilization check failed: {e}")
        print("Proceeding anyway after 30s buffer...")
        time.sleep(30)

def _configure_scaling_policy(policy_type):
    """Configure autoscaling policy before test"""
    print(f"Applying scaling policy: {policy_type}...")
    
    result = subprocess.run([
        "terraform",
        "-chdir=flash-sale-platform/terraform", 
        "apply",
        "-auto-approve",
        f"-var=scaling_policy_type={policy_type}"
    ], check=True, capture_output=True, text=True)
    
    print("Terraform apply complete")
    
    service_name = os.getenv('SERVICE_NAME', 'flash-sale-platform')
    cluster_name = f"{service_name}-cluster"
    _wait_for_ecs_stability(service_name, cluster_name)

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

    _configure_scaling_policy(str(config.get('scaling_policy', 'target_tracking')))

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
        print(f"\nTest complete! Results saved to results/locust_data/{scenario_name}_test_results.json")


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python run_scenario.py <scenario_name>")
        sys.exit(1)
    run_scenario(sys.argv[1])