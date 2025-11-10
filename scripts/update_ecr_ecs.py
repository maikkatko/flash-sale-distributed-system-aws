# update_ecr_and_ecs.py
import sys
import subprocess
import os

from dotenv import load_dotenv

load_dotenv()


def run_command(cmd, shell=False):
    """Run a command and handle errors"""
    result = subprocess.run(
        cmd, shell=shell, capture_output=True, check=False, text=True)
    if result.returncode != 0:
        print(f"Error: {result.stderr}")
        sys.exit(1)
    return result.stdout


def _update_ecr_image(ecr_url, image_name, dockerfile_path):
    region = os.getenv('AWS_REGION', 'us-east-1')

    print(f"Building {image_name}...")
    run_command(['docker', 'build', '-f',
                dockerfile_path, '-t', image_name, 'src'])

    print(f"Tagging {image_name}...")
    run_command(['docker', 'tag', f'{image_name}:latest', ecr_url])

    print("Logging into ECR...")
    login_cmd = f'aws ecr get-login-password --region {region} | docker login --username AWS --password-stdin {ecr_url.split("/")[0]}'
    run_command(login_cmd, shell=True)

    print("Pushing to ECR...")
    run_command(['docker', 'push', ecr_url])


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
    image_name = os.getenv('ECR_REPO_NAME')
    dockerfile_path = 'flash-sale-platform/src'
    ecr_url = os.getenv('ECR_REPO_URL')
    ecs_cluster = os.getenv('ECS_CLUSTER_NAME')
    ecs_service = os.getenv('ECS_SERVICE_NAME')

    if not ecr_url:
        print("Error: Could not find ECR URL for Server!")
        sys.exit(1)

    _update_ecr_image(ecr_url, image_name, dockerfile_path)
    _update_ecs_service(ecs_cluster, ecs_service)


if __name__ == "__main__":
    main()
