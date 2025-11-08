.PHONY: setup test-mysql test-dynamodb run-tests analyze clean help

setup:
	@if not exist results mkdir results
	@if not exist results\data mkdir results\data
	@if not exist results\reports mkdir results\reports
	@if not exist results\charts mkdir results\charts
	@pip install -r requirements.txt || pip3 install -r requirements.txt
	@terraform -chdir=terraform plan
	@terraform -chdir=terraform apply -auto-approve
	@python scripts/init_aws_env_vars.py
	@echo Environment ready for testing

setup-aws:
	@terraform -chdir=terraform plan
	@terraform -chdir=terraform apply -auto-approve
	@python scripts/init_aws_env_vars.py
	@echo AWS infrastructure setup complete

init-aws-vars:
	@python scripts/init_aws_env_vars.py

update-tf:
	@terraform -chdir=terraform apply -auto-approve
	@echo AWS infrastructure updated

update-server:
	@echo === Updating Server Service ===
	@python scripts/update_ecr_ecs.py server

restart-server:
	@echo === Restarting Server Service ===
	@powershell.exe -Command "aws ecs update-service --cluster cs6650l2-cluster --service cs6650l2 --force-new-deployment" > $null
	@echo Server service restarted

test-basic:
	@echo === MySQL Load Test ===
	@python scripts/run_scenario.py basic-test

test-flash-sale:
	@echo === DynamoDB Load Test ===
	@python scripts/run_scenario.py flash-sale

combine-results:
	@echo === Combining Test Results ===
	@python scripts\combine_test_results.py || python3 scripts\combine_test_results.py

run-tests:
	@echo === Running All Tests ===
	@make test-mysql
	@make test-dynamodb
	@make combine-results
	@echo === All tests completed! ===

analyze:
	@echo === Analyzing Results ===
	@python scripts\combine_db_test_results.py || python3 scripts\combine_db_test_results.py
	@python scripts\analyze_mysql_vs_dynamodb.py || python3 scripts\analyze_mysql_vs_dynamodb.py

clean:
	@docker-compose down 2>nul || echo.
	@docker system prune -f
	@terraform -chdir=terraform destroy -auto-approve
	@if exist terraform\lambda-worker.zip del terraform\lambda-worker.zip
	@echo Cleanup complete and AWS infrastructure destroyed!

help:
	@echo.
	@echo   Load Testing Commands
	@echo.
	@echo   make setup     	 			 - Setup environment and AWS infra
	@echo   make test-mysql        - MySQL load test
	@echo   make test-dynamodb     - DynamoDB load test
	@echo   make run-tests         - Run both tests with analysis
	@echo   make analyze           - Analyze test results
	@echo   make update-api        - Update API service only
	@echo   make update-lambda     - Update Lambda worker only
	@echo   make clean             - Stop, cleanup Docker, and destroy AWS infra
	@echo.