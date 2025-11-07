# init_aws_env_vars.py
import json
import subprocess
import sys


def _get_aws_configuration_details(service_type: str, command: str, *args) -> str:
    result = subprocess.run(
        ['aws', service_type, command, *args],
        check=False,
        capture_output=True,
        text=True
    )
    if result.returncode != 0:
        print(f"Error: {result.stderr}")
        sys.exit(1)
    return result.stdout
    
def _get_aws_account_id():
    result = _get_aws_configuration_details('sts', 'get-caller-identity')
    account_info = json.loads(result)
    return account_info['Account']

def _get_aws_region():
    return _get_aws_configuration_details('configure', 'get', 'region')

def _get_aws_albs():
    result = _get_aws_configuration_details('elbv2', 'describe-load-balancers')
    lb_info = json.loads(result)
    return lb_info['LoadBalancers']

def _get_aws_ecs_clusters_arns():
    result = _get_aws_configuration_details('ecs', 'list-clusters')
    clusters = json.loads(result)
    return clusters['clusterArns']

def _get_aws_ecr_repositories():
    result = _get_aws_configuration_details('ecr', 'describe-repositories')
    repos = json.loads(result)
    return repos['repositories']

def _get_aws_ecs_service_arns():
    result = _get_aws_configuration_details('ecs', 'list-services')
    services = json.loads(result)
    return services['serviceArns']

def set_aws_env_vars():
    alb_dns_names = [
        f"http://{alb['DNSName']}"
        for alb in _get_aws_albs()
    ]

    ecs_cluster_arns = _get_aws_ecs_clusters_arns()

    ecs_service_arns = _get_aws_ecs_service_arns()

    ecr_repo_names = [
        ecr_repo['repositoryName']
        for ecr_repo in _get_aws_ecr_repositories()
    ]

    ecr_repo_uris = [
        ecr_repo['repositoryUri']
        for ecr_repo in _get_aws_ecr_repositories()
    ]

    with open('.env', 'w', encoding='utf-8') as f:
        f.write(f"AWS_ACCOUNT_ID={_get_aws_account_id()}\n")
        f.write(f"AWS_REGION={_get_aws_region}\n")
        f.write(f"ALB_DNS_NAME={alb_dns_names}\n")
        f.write(f"ECS_CLUSTER={ecs_cluster_arns}\n")
        f.write(f"ECS_SERVICE={ecs_service_arns}\n")
        f.write(f"ECR_REPO={ecr_repo_names}\n")
        f.write(f"ECR_URL={ecr_repo_uris}\n")

    print("AWS environment variables set in .env file.")

if __name__ == "__main__":
    set_aws_env_vars()