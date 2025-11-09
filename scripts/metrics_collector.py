import os
import boto3
import json
import time
from dotenv import load_dotenv
from datetime import datetime
import pandas as pd

load_dotenv()

cloudwatch = boto3.client('cloudwatch', region_name='us-east-1')

METRICS_CONFIG = {
    'cluster_name': os.getenv('CLUSTER_NAME'),
    'service_name': os.getenv('SERVICE_NAME'),
    'target_group_name': os.getenv('TARGET_GROUP_NAME'),
    'load_balancer_name': os.getenv('LOAD_BALANCER_NAME'),
    'rds_instance_id': os.getenv('RDS_INSTANCE_ID'),
    'region': os.getenv('AWS_REGION')
}

class AWSMetricsCollector:
    def __init__(self, config):
        """
        config = {
            'cluster_name': 'your-cluster',
            'service_name': 'your-service',
            'target_group_name': 'targetgroup/...',
            'load_balancer_name': 'app/your-alb/...',
            'rds_instance_id': 'your-db-instance',
            'region': 'us-east-1'
        }
        """
        self.config = config
        self.test_start_time = None
        self.test_end_time = None
        
    def start_collection(self):
        """Mark the start of the test"""
        self.test_start_time = datetime.utcnow()
        print(f"[METRICS] Starting collection at {self.test_start_time}")
    
    def stop_collection(self):
        """Mark the end of the test"""
        self.test_end_time = datetime.utcnow()
        print(f"[METRICS] Stopping collection at {self.test_end_time}")
    
    def _get_metric(self, namespace, metric_name, dimensions, statistics=['Average'], period=60):
        """Helper to fetch a single metric"""
        try:
            response = cloudwatch.get_metric_statistics(
                Namespace=namespace,
                MetricName=metric_name,
                Dimensions=dimensions,
                StartTime=self.test_start_time,
                EndTime=self.test_end_time,
                Period=period,
                Statistics=statistics
            )
            return sorted(response['Datapoints'], key=lambda x: x['Timestamp'])
        except Exception as e:
            print(f"[ERROR] Failed to get {metric_name}: {e}")
            return []
    
    def _get_ecs_metrics(self):
        """Collect all ECS-related metrics"""
        print("[METRICS] Collecting ECS metrics...")
        
        metrics = {
            'task_cpu': self._get_metric(
                namespace='ECS/ContainerInsights',
                metric_name='TaskCpuUtilization',
                dimensions=[
                    {'Name': 'ServiceName', 'Value': self.config['service_name']},
                    {'Name': 'ClusterName', 'Value': self.config['cluster_name']}
                ],
                statistics=['Average', 'Maximum']
            ),
            'task_memory': self._get_metric(
                namespace='ECS/ContainerInsights',
                metric_name='TaskMemoryUtilization',
                dimensions=[
                    {'Name': 'ServiceName', 'Value': self.config['service_name']},
                    {'Name': 'ClusterName', 'Value': self.config['cluster_name']}
                ],
                statistics=['Average', 'Maximum']
            ),
            'desired_task_count': self._get_metric(
                namespace='ECS/ContainerInsights',
                metric_name='DesiredTaskCount',
                dimensions=[
                    {'Name': 'ServiceName', 'Value': self.config['service_name']},
                    {'Name': 'ClusterName', 'Value': self.config['cluster_name']}
                ],
                statistics=['Average']
            ),
            'running_task_count': self._get_metric(
                namespace='ECS/ContainerInsights',
                metric_name='RunningTaskCount',
                dimensions=[
                    {'Name': 'ServiceName', 'Value': self.config['service_name']},
                    {'Name': 'ClusterName', 'Value': self.config['cluster_name']}
                ],
                statistics=['Average']
            )
        }
        
        return metrics
    
    def _get_alb_metrics(self):
        """Collect all ALB-related metrics"""
        print("[METRICS] Collecting ALB metrics...")
        
        metrics = {
            'request_count': self._get_metric(
                namespace='AWS/ApplicationELB',
                metric_name='RequestCount',
                dimensions=[
                    {'Name': 'LoadBalancer', 'Value': self.config['load_balancer_name']}
                ],
                statistics=['Sum']
            ),
            'target_response_time': self._get_metric(
                namespace='AWS/ApplicationELB',
                metric_name='TargetResponseTime',
                dimensions=[
                    {'Name': 'LoadBalancer', 'Value': self.config['load_balancer_name']}
                ],
                statistics=['Average', 'Maximum']
            ),
            'healthy_host_count': self._get_metric(
                namespace='AWS/ApplicationELB',
                metric_name='HealthyHostCount',
                dimensions=[
                    {'Name': 'TargetGroup', 'Value': self.config['target_group_name']},
                    {'Name': 'LoadBalancer', 'Value': self.config['load_balancer_name']}
                ],
                statistics=['Average']
            ),
            'unhealthy_host_count': self._get_metric(
                namespace='AWS/ApplicationELB',
                metric_name='UnHealthyHostCount',
                dimensions=[
                    {'Name': 'TargetGroup', 'Value': self.config['target_group_name']},
                    {'Name': 'LoadBalancer', 'Value': self.config['load_balancer_name']}
                ],
                statistics=['Average']
            ),
            'http_2xx_count': self._get_metric(
                namespace='AWS/ApplicationELB',
                metric_name='HTTPCode_Target_2XX_Count',
                dimensions=[
                    {'Name': 'LoadBalancer', 'Value': self.config['load_balancer_name']}
                ],
                statistics=['Sum']
            ),
            'http_4xx_count': self._get_metric(
                namespace='AWS/ApplicationELB',
                metric_name='HTTPCode_Target_4XX_Count',
                dimensions=[
                    {'Name': 'LoadBalancer', 'Value': self.config['load_balancer_name']}
                ],
                statistics=['Sum']
            ),
            'http_5xx_count': self._get_metric(
                namespace='AWS/ApplicationELB',
                metric_name='HTTPCode_Target_5XX_Count',
                dimensions=[
                    {'Name': 'LoadBalancer', 'Value': self.config['load_balancer_name']}
                ],
                statistics=['Sum']
            )
        }
        
        return metrics
    
    def _get_rds_metrics(self):
        """Collect all RDS-related metrics"""
        print("[METRICS] Collecting RDS metrics...")
        
        metrics = {
            'cpu_utilization': self._get_metric(
                namespace='AWS/RDS',
                metric_name='CPUUtilization',
                dimensions=[
                    {'Name': 'DBInstanceIdentifier', 'Value': self.config['rds_instance_id']}
                ],
                statistics=['Average', 'Maximum']
            ),
            'database_connections': self._get_metric(
                namespace='AWS/RDS',
                metric_name='DatabaseConnections',
                dimensions=[
                    {'Name': 'DBInstanceIdentifier', 'Value': self.config['rds_instance_id']}
                ],
                statistics=['Average', 'Maximum']
            ),
            'read_iops': self._get_metric(
                namespace='AWS/RDS',
                metric_name='ReadIOPS',
                dimensions=[
                    {'Name': 'DBInstanceIdentifier', 'Value': self.config['rds_instance_id']}
                ],
                statistics=['Average', 'Maximum']
            ),
            'write_iops': self._get_metric(
                namespace='AWS/RDS',
                metric_name='WriteIOPS',
                dimensions=[
                    {'Name': 'DBInstanceIdentifier', 'Value': self.config['rds_instance_id']}
                ],
                statistics=['Average', 'Maximum']
            ),
            'read_latency': self._get_metric(
                namespace='AWS/RDS',
                metric_name='ReadLatency',
                dimensions=[
                    {'Name': 'DBInstanceIdentifier', 'Value': self.config['rds_instance_id']}
                ],
                statistics=['Average', 'Maximum']
            ),
            'write_latency': self._get_metric(
                namespace='AWS/RDS',
                metric_name='WriteLatency',
                dimensions=[
                    {'Name': 'DBInstanceIdentifier', 'Value': self.config['rds_instance_id']}
                ],
                statistics=['Average', 'Maximum']
            ),
            'network_receive_throughput': self._get_metric(
                namespace='AWS/RDS',
                metric_name='NetworkReceiveThroughput',
                dimensions=[
                    {'Name': 'DBInstanceIdentifier', 'Value': self.config['rds_instance_id']}
                ],
                statistics=['Average']
            ),
            'network_transmit_throughput': self._get_metric(
                namespace='AWS/RDS',
                metric_name='NetworkTransmitThroughput',
                dimensions=[
                    {'Name': 'DBInstanceIdentifier', 'Value': self.config['rds_instance_id']}
                ],
                statistics=['Average']
            ),
            'free_storage_space': self._get_metric(
                namespace='AWS/RDS',
                metric_name='FreeStorageSpace',
                dimensions=[
                    {'Name': 'DBInstanceIdentifier', 'Value': self.config['rds_instance_id']}
                ],
                statistics=['Average', 'Minimum']
            )
        }
        
        return metrics
    
    def export_all_metrics(self, output_format='json'):
        """Export all metrics after test completes"""
        print("[METRICS] Exporting all metrics...")
        
        start_iso = self.test_start_time.isoformat() if self.test_start_time else None
        end_iso = self.test_end_time.isoformat() if self.test_end_time else None
        duration_seconds = (
            (self.test_end_time - self.test_start_time).total_seconds()
            if self.test_start_time and self.test_end_time
            else None
        )

        all_metrics = {
            'test_info': {
                'start_time': start_iso,
                'end_time': end_iso,
                'duration_seconds': duration_seconds,
                'config': self.config
            },
            'ecs': self._get_ecs_metrics(),
            'alb': self._get_alb_metrics(),
            'rds': self._get_rds_metrics()
        }
        
        timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
        
        if output_format == 'json':
            filename = f'results/cloudwatch_metrics/cloudwatch_metrics_{timestamp}.json'
            with open(filename, 'w') as f:
                json.dump(all_metrics, f, indent=2, default=str)
            print(f"[METRICS] Exported to {filename}")
        
        elif output_format == 'csv':
            self._export_to_csv(all_metrics, timestamp)
        
        return all_metrics
    
    def _export_to_csv(self, all_metrics, timestamp):
        """Export metrics to separate CSV files for easier analysis"""
        
        # ECS Metrics CSV
        ecs_rows = []
        for metric_name, datapoints in all_metrics['ecs'].items():
            for point in datapoints:
                ecs_rows.append({
                    'timestamp': point['Timestamp'],
                    'metric': metric_name,
                    'average': point.get('Average'),
                    'maximum': point.get('Maximum'),
                    'unit': point.get('Unit')
                })
        if ecs_rows:
            df = pd.DataFrame(ecs_rows)
            df.to_csv(f'ecs_metrics_{timestamp}.csv', index=False)
            print(f"[METRICS] Exported ECS metrics to ecs_metrics_{timestamp}.csv")
        
        # ALB Metrics CSV
        alb_rows = []
        for metric_name, datapoints in all_metrics['alb'].items():
            for point in datapoints:
                alb_rows.append({
                    'timestamp': point['Timestamp'],
                    'metric': metric_name,
                    'average': point.get('Average'),
                    'maximum': point.get('Maximum'),
                    'sum': point.get('Sum'),
                    'unit': point.get('Unit')
                })
        if alb_rows:
            df = pd.DataFrame(alb_rows)
            df.to_csv(f'alb_metrics_{timestamp}.csv', index=False)
            print(f"[METRICS] Exported ALB metrics to alb_metrics_{timestamp}.csv")
        
        # RDS Metrics CSV
        rds_rows = []
        for metric_name, datapoints in all_metrics['rds'].items():
            for point in datapoints:
                rds_rows.append({
                    'timestamp': point['Timestamp'],
                    'metric': metric_name,
                    'average': point.get('Average'),
                    'maximum': point.get('Maximum'),
                    'minimum': point.get('Minimum'),
                    'unit': point.get('Unit')
                })
        if rds_rows:
            df = pd.DataFrame(rds_rows)
            df.to_csv(f'rds_metrics_{timestamp}.csv', index=False)
            print(f"[METRICS] Exported RDS metrics to rds_metrics_{timestamp}.csv")

metrics_collector = AWSMetricsCollector(METRICS_CONFIG)