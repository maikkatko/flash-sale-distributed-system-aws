import os
import subprocess
import sys


def run_command(cmd, shell=False):
    """Run a command and handle errors"""
    result = subprocess.run(
        cmd, shell=shell, capture_output=True, check=False, text=True)
    if result.returncode != 0:
        print(f"Error: {result.stderr}")
        sys.exit(1)
    return result.stdout

def _update_ecs_service(cluster_name, service_name):
    print("Updating ECS service...")
    run_command([
        'aws', 'ecs', 'update-service',
        '--cluster', cluster_name,
        '--service', service_name,
        '--force-new-deployment'
    ])

    print("Waiting for service to stabilize...")
    run_command([
        'aws', 'ecs', 'wait', 'services-stable',
        '--cluster', cluster_name,
        '--services', service_name
    ])

    print(f"{service_name} updated successfully!")

def main(): 
    ecr_url = os.getenv('ECR_REPO_URL')
    ecs_cluster = os.getenv('ECS_CLUSTER_NAME')
    ecs_service = os.getenv('ECS_SERVICE_NAME')

    if not ecr_url:
        print("Error: Could not find ECR URL for Server!")
        sys.exit(1)

    _update_ecs_service(ecs_cluster, ecs_service)