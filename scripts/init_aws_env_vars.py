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
    result = _get_aws_configuration_details('configure', 'get', 'region').strip()
    return result

def _get_aws_albs():
    result = _get_aws_configuration_details('elbv2', 'describe-load-balancers')
    lb_info = json.loads(result)
    return lb_info['LoadBalancers']

def _get_aws_target_groups():
    result = _get_aws_configuration_details('elbv2', 'describe-target-groups')
    tg_info = json.loads(result)
    return tg_info['TargetGroups']

def _get_aws_ecs_clusters():
    # Get cluster ARNs
    result = _get_aws_configuration_details('ecs', 'list-clusters')
    cluster_arns = json.loads(result)['clusterArns']
    
    if not cluster_arns:
        return []
    
    # Get detailed cluster info
    result = _get_aws_configuration_details('ecs', 'describe-clusters', '--clusters', *cluster_arns)
    clusters = json.loads(result)
    return clusters['clusters']

def _get_aws_ecs_services(cluster_arn):
    # Get service ARNs for a specific cluster
    result = _get_aws_configuration_details('ecs', 'list-services', '--cluster', cluster_arn)
    service_arns = json.loads(result)['serviceArns']
    
    if not service_arns:
        return []
    
    # Get detailed service info
    result = _get_aws_configuration_details('ecs', 'describe-services', '--cluster', cluster_arn, '--services', *service_arns)
    services = json.loads(result)
    return services['services']

def _get_aws_rds_instances():
    result = _get_aws_configuration_details('rds', 'describe-db-instances')
    instances = json.loads(result)
    return instances['DBInstances']

def _get_aws_ecr_repositories():
    result = _get_aws_configuration_details('ecr', 'describe-repositories')
    repos = json.loads(result)
    return repos['repositories']

def _extract_load_balancer_name(alb_arn):
    """
    Extract load balancer name from ARN
    ARN format: arn:aws:elasticloadbalancing:region:account-id:loadbalancer/app/my-alb/1234567890abcdef
    Returns: app/my-alb/1234567890abcdef
    """
    return '/'.join(alb_arn.split(':')[-1].split('/')[1:])

def _extract_target_group_name(tg_arn):
    """
    Extract target group name from ARN
    ARN format: arn:aws:elasticloadbalancing:region:account-id:targetgroup/my-tg/1234567890abcdef
    Returns: targetgroup/my-tg/1234567890abcdef
    """
    parts = tg_arn.split(':')[-1].split('/')
    return f"targetgroup/{parts[1]}/{parts[2]}"

def _extract_name_from_arn(arn):
    return arn.split('/')[-1]

def set_aws_env_vars():
    # Get all AWS resources
    albs = _get_aws_albs()
    target_groups = _get_aws_target_groups()
    clusters = _get_aws_ecs_clusters()
    rds_instances = _get_aws_rds_instances()
    ecr_repos = _get_aws_ecr_repositories()
    region = _get_aws_region()
    
    # Extract ALB info
    alb_dns_names = [f"http://{alb['DNSName']}" for alb in albs]
    alb_arns = [alb['LoadBalancerArn'] for alb in albs]
    alb_names = [_extract_load_balancer_name(arn) for arn in alb_arns]
    
    # Extract Target Group info
    tg_arns = [tg['TargetGroupArn'] for tg in target_groups]
    tg_names = [_extract_target_group_name(arn) for arn in tg_arns]
    
    # Extract ECS Cluster info
    cluster_arns = [cluster['clusterArn'] for cluster in clusters]
    cluster_names = [_extract_name_from_arn(arn) for arn in cluster_arns]
    
    # Extract ECS Service info (from first cluster if multiple exist)
    service_arns = []
    service_names = []
    if cluster_arns:
        services = _get_aws_ecs_services(cluster_arns[0])
        service_arns = [service['serviceArn'] for service in services]
        service_names = [_extract_name_from_arn(arn) for arn in service_arns]
    
    # Extract RDS info
    rds_instance_ids = [db['DBInstanceIdentifier'] for db in rds_instances]
    
    # Extract ECR info
    ecr_repo_names = [repo['repositoryName'] for repo in ecr_repos]
    ecr_repo_uris = [repo['repositoryUri'] for repo in ecr_repos]
    
    # Write to .env file
    with open('.env', 'w', encoding='utf-8') as f:
        f.write(f"AWS_ACCOUNT_ID={_get_aws_account_id()}\n")
        f.write(f"AWS_REGION={region}\n")
        
        # ALB info
        f.write(f"ALB_DNS_NAME={alb_dns_names[0] if alb_dns_names else ''}\n")
        f.write(f"ALB_ARN={alb_arns[0] if alb_arns else ''}\n")
        f.write(f"LOAD_BALANCER_NAME={alb_names[0] if alb_names else ''}\n")
        
        # Target Group info
        f.write(f"TARGET_GROUP_ARN={tg_arns[0] if tg_arns else ''}\n")
        f.write(f"TARGET_GROUP_NAME={tg_names[0] if tg_names else ''}\n")
        
        # ECS Cluster info
        f.write(f"ECS_CLUSTER_ARN={cluster_arns[0] if cluster_arns else ''}\n")
        f.write(f"CLUSTER_NAME={cluster_names[0] if cluster_names else ''}\n")
        
        # ECS Service info
        f.write(f"ECS_SERVICE_ARN={service_arns[0] if service_arns else ''}\n")
        f.write(f"SERVICE_NAME={service_names[0] if service_names else ''}\n")
        
        # RDS info
        f.write(f"RDS_INSTANCE_ID={rds_instance_ids[0] if rds_instance_ids else ''}\n")
        
        # ECR info
        f.write(f"ECR_REPO_NAME={ecr_repo_names[0] if ecr_repo_names else ''}\n")
        f.write(f"ECR_REPO_URL={ecr_repo_uris[0] if ecr_repo_uris else ''}\n")

        # RDS DB Info
        f.write(f"TF_VAR_db_name=flashsale\n")
        f.write(f"TF_VAR_db_username=admin\n")
        f.write(f"TF_VAR_db_password=SecurePassword123!\n")

    print("AWS environment variables set in .env file.")
    print("\n=== CloudWatch Metrics Configuration ===")
    print(f"Region: {region}")
    print(f"Cluster Name: {cluster_names[0] if cluster_names else 'N/A'}")
    print(f"Service Name: {service_names[0] if service_names else 'N/A'}")
    print(f"Load Balancer Name: {alb_names[0] if alb_names else 'N/A'}")
    print(f"Target Group Name: {tg_names[0] if tg_names else 'N/A'}")
    print(f"RDS Instance ID: {rds_instance_ids[0] if rds_instance_ids else 'N/A'}")

if __name__ == "__main__":
    set_aws_env_vars()