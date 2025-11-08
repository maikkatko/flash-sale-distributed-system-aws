.PHONY: setup test-baseline test-high-contention test-thundering-herd test-sustained test-chaos test-fairness run-exp1 run-exp2 run-exp3 run-exp4 run-all analyze clean help

setup:
	@if not exist results mkdir results
	@if not exist results\formatted_locust_data mkdir results\formatted_locust_data
	@if not exist results\cloudwatch_metrics mkdir results\cloudwatch_metrics
	@if not exist results\raw_locust_data mkdir results\raw_locust_data
	@if not exist results\locust_reports mkdir results\locust_reports
	@if not exist results\charts mkdir results\charts
	@pip install -r requirements.txt || pip3 install -r requirements.txt
	@terraform -chdir=flash-sale-platform/terraform plan
	@terraform -chdir=flash-sale-platform/terraform apply -auto-approve
	@python scripts/init_aws_env_vars.py
	@echo Environment ready for testing

setup-aws:
	@terraform -chdir=flash-sale-platform/terraform plan
	@terraform -chdir=flash-sale-platform/terraform apply -auto-approve
	@python scripts/init_aws_env_vars.py
	@echo AWS infrastructure setup complete

initialize-aws-env-vars:
	@python scripts/init_aws_env_vars.py

# Individual test scenarios
test-baseline:
	@echo === Baseline Test ===
	@python scripts/run_scenario.py baseline

test-high-contention:
	@echo === High Contention Test (Exp 1) ===
	@python scripts/run_scenario.py high_contention

test-thundering-herd:
	@echo === Thundering Herd Test (Exp 2) ===
	@python scripts/run_scenario.py thundering_herd

test-sustained:
	@echo === Sustained Load Test (Exp 2) ===
	@python scripts/run_scenario.py sustained_load

test-chaos:
	@echo === Chaos Testing (Exp 3) ===
	@python scripts/run_scenario.py chaos_baseline

test-fairness:
	@echo === Fairness Test (Exp 4) ===
	@python scripts/run_scenario.py fairness_10x_demand

# Run experiments by group
run-exp1:
	@echo === Experiment 1: High-Contention Inventory ===
	@make test-baseline
	@make test-high-contention
	@echo === Experiment 1 Complete ===

run-exp2:
	@echo === Experiment 2: Autoscaling Analysis ===
	@make test-baseline
	@make test-thundering-herd
	@make test-sustained
	@echo === Experiment 2 Complete ===

run-exp3:
	@echo === Experiment 3: Failure Recovery ===
	@make test-baseline
	@echo NOTE: Inject failures manually during next test
	@make test-chaos
	@echo === Experiment 3 Complete ===

run-exp4:
	@echo === Experiment 4: Fairness Evaluation ===
	@make test-baseline
	@echo Running WITHOUT rate limiting...
	@make test-fairness
	@echo === Experiment 4 Complete (run twice: with/without FIFO) ===

run-all:
	@echo === Running All Experiments ===
	@make run-exp1
	@make run-exp2
	@make run-exp3
	@make run-exp4
	@echo === All experiments completed! ===

analyze:
	@echo === Analyzing Results ===
	@python scripts/analyze_results.py || python3 scripts/analyze_results.py

update-server:
	@echo === Updating Server Service ===
	@python scripts/update_ecr_ecs.py server

restart-server:
	@echo === Restarting Server Service ===
	@powershell.exe -Command "aws ecs update-service --cluster flash-sale-cluster --service flash-sale-api --force-new-deployment" > $null
	@echo Server service restarted

clean:
	@docker-compose down 2>nul || echo.
	@docker system prune -f
	@terraform -chdir=flash-sale-platform/terraform destroy -auto-approve
	@echo Cleanup complete and AWS infrastructure destroyed!

help:
	@echo.
	@echo   Load Testing Commands
	@echo.
	@echo   make setup              - Setup environment and AWS infra
	@echo   make test-baseline      - Run baseline test
	@echo   make test-high-contention - Run high contention test
	@echo   make test-thundering-herd - Run thundering herd test
	@echo   make test-sustained     - Run sustained load test
	@echo   make test-chaos         - Run chaos testing
	@echo   make test-fairness      - Run fairness test
	@echo.
	@echo   make run-exp1           - Run Experiment 1 (High Contention)
	@echo   make run-exp2           - Run Experiment 2 (Autoscaling)
	@echo   make run-exp3           - Run Experiment 3 (Failure Recovery)
	@echo   make run-exp4           - Run Experiment 4 (Fairness)
	@echo   make run-all            - Run all experiments
	@echo.
	@echo   make analyze            - Analyze test results
	@echo   make update-server      - Update API service only
	@echo   make clean              - Stop, cleanup Docker, and destroy AWS infra
	@echo.