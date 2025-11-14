import subprocess
import sys
from dotenv import load_dotenv
import os

def main():
    # Load .env file
    load_dotenv()
    
    cluster_name = os.getenv("CLUSTER_NAME")
    service_name = os.getenv("SERVICE_NAME")
    desired_count = sys.argv[1] if len(sys.argv) > 1 else "1"
    
    if not cluster_name or not service_name:
        print("Error: CLUSTER_NAME and SERVICE_NAME must be set in .env")
        sys.exit(1)
    
    print(f"Resetting {service_name} to {desired_count} tasks...")
    
    result = subprocess.run([
        "aws", "ecs", "update-service",
        "--cluster", cluster_name,
        "--service", service_name,
        "--desired-count", desired_count
    ], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    
    if result.returncode == 0:
        print("Task count updated successfully")
    else:
        print("Failed to update task count")
    
    sys.exit(result.returncode)

if __name__ == "__main__":
    main()