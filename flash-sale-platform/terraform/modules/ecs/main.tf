  # ECS Cluster
  resource "aws_ecs_cluster" "this" {
    name = "${var.service_name}-cluster"

    setting {
      name  = "containerInsights"
      value = "enabled" 
    }
  }

  # Task Definition
  resource "aws_ecs_task_definition" "this" {
    family                   = "${var.service_name}-task"
    requires_compatibilities = ["FARGATE"]
    network_mode             = "awsvpc"
    cpu                      = "256"
    memory                   = "512"
    execution_role_arn       = var.execution_role_arn
    task_role_arn            = var.task_role_arn

    container_definitions = jsonencode([
      {
        name      = var.service_name
        image     = var.image
        essential = true
        portMappings = [
          {
            containerPort = var.container_port
            protocol      = "tcp"
          }
        ]
        logConfiguration = {
          logDriver = "awslogs"
          options = {
            "awslogs-group"         = var.log_group_name
            "awslogs-region"        = var.region
            "awslogs-stream-prefix" = "ecs"
          }
        }
        environment = [for name, value in var.environment_variables : {
          name  = name
          value = value
        }]
      }
    ])
  }

  # ECS Service with ALB Integration
  resource "aws_ecs_service" "this" {
    name            = var.service_name
    cluster         = aws_ecs_cluster.this.id
    task_definition = aws_ecs_task_definition.this.arn
    desired_count   = var.min_capacity  # Start with min capacity
    launch_type     = "FARGATE"

    network_configuration {
      subnets          = var.subnet_ids
      security_groups  = var.security_group_ids
      assign_public_ip = false
    }

    # ALB Integration - NEW!
    load_balancer {
      target_group_arn = var.target_group_arn
      container_name   = var.service_name
      container_port   = var.container_port
    }

    # Required for auto-scaling
    lifecycle {
      ignore_changes = [desired_count]
    }
  }

  resource "aws_appautoscaling_target" "ecs_target" {
    max_capacity       = var.max_capacity
    min_capacity       = var.min_capacity
    resource_id        = "service/${aws_ecs_cluster.this.name}/${aws_ecs_service.this.name}"
    scalable_dimension = "ecs:service:DesiredCount"
    service_namespace  = "ecs"
  }

  resource "aws_appautoscaling_policy" "cpu_scaling" {
    count              = var.scaling_policy_type == "target_tracking" ? 1 : 0
    name               = "${var.service_name}-cpu-scaling"
    policy_type        = "TargetTrackingScaling"
    resource_id        = aws_appautoscaling_target.ecs_target.resource_id
    scalable_dimension = aws_appautoscaling_target.ecs_target.scalable_dimension
    service_namespace  = aws_appautoscaling_target.ecs_target.service_namespace

    target_tracking_scaling_policy_configuration {
      predefined_metric_specification {
        predefined_metric_type = "ECSServiceAverageCPUUtilization"
      }

      target_value       = var.cpu_target_value
      scale_in_cooldown  = var.scale_in_cooldown
      scale_out_cooldown = var.scale_out_cooldown
    }
  }

resource "aws_appautoscaling_policy" "ecs_requests" {
  count              = var.scaling_policy_type == "step_scaling" ? 1 : 0
  name               = "${var.service_name}-request-step-scaling-up"
  policy_type        = "StepScaling"
  resource_id        = aws_appautoscaling_target.ecs_target.resource_id
  scalable_dimension = aws_appautoscaling_target.ecs_target.scalable_dimension
  service_namespace  = aws_appautoscaling_target.ecs_target.service_namespace

  step_scaling_policy_configuration {
    adjustment_type         = "ChangeInCapacity"
    cooldown                = 60
    metric_aggregation_type = "Average"

    step_adjustment {
      scaling_adjustment          = 2
      
      metric_interval_lower_bound = 0
      metric_interval_upper_bound = 100
    }
    step_adjustment {
      scaling_adjustment          = 4
      metric_interval_lower_bound = 100
      metric_interval_upper_bound = 200
    }
    step_adjustment {
      scaling_adjustment          = 6
      metric_interval_lower_bound = 200
    }
  }
}

resource "aws_cloudwatch_metric_alarm" "step_scale_up" {
  count               = var.scaling_policy_type == "step_scaling" ? 1 : 0
  alarm_name          = "${var.service_name}-requests-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "RequestCountPerTarget"
  namespace           = "AWS/ApplicationELB"
  period              = 60
  statistic           = "Sum"
  threshold           = 1000
  
  dimensions = {
    TargetGroup  = split(":", var.target_group_arn)[5]
    LoadBalancer = var.alb_arn_suffix
  }
  
  alarm_actions = [aws_appautoscaling_policy.ecs_requests[0].arn]
}

resource "aws_appautoscaling_policy" "ecs_requests_down" {
  count              = var.scaling_policy_type == "step_scaling" ? 1 : 0
  name               = "${var.service_name}-request-step-scaling-down"
  policy_type        = "StepScaling"
  resource_id        = aws_appautoscaling_target.ecs_target.resource_id
  scalable_dimension = aws_appautoscaling_target.ecs_target.scalable_dimension
  service_namespace  = aws_appautoscaling_target.ecs_target.service_namespace

  step_scaling_policy_configuration {
    adjustment_type         = "ChangeInCapacity"
    cooldown                = 300
    metric_aggregation_type = "Average"

    step_adjustment {
      scaling_adjustment          = -1
      metric_interval_upper_bound = 0
    }
  }
}

# Scale DOWN alarm
resource "aws_cloudwatch_metric_alarm" "step_scale_down" {
  count               = var.scaling_policy_type == "step_scaling" ? 1 : 0
  alarm_name          = "${var.service_name}-requests-low"
  comparison_operator = "LessThanThreshold"
  evaluation_periods  = 2 
  metric_name         = "RequestCountPerTarget"
  namespace           = "AWS/ApplicationELB"
  period              = 60
  statistic           = "Sum"
  threshold           = 500
  
  dimensions = {
    TargetGroup  = split(":", var.target_group_arn)[5]
    LoadBalancer = var.alb_arn_suffix
  }
  
  alarm_actions = [aws_appautoscaling_policy.ecs_requests_down[0].arn]
}
