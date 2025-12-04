import subprocess
import sys
from dotenv import load_dotenv
import os

def main():
    # Load .env file
    load_dotenv()
    
    # Get scaling policy from command line arg
    scaling_policy = sys.argv[1] if len(sys.argv) > 1 else "target_tracking"
    
    # Run terraform with the variable as a command line arg (more reliable)
    result = subprocess.run([
        "terraform",
        "-chdir=flash-sale-platform/terraform",
        "apply",
        "-auto-approve",
        f"-var=scaling_policy_type={scaling_policy}"
    ], env=os.environ)
    
    sys.exit(result.returncode)

if __name__ == "__main__":
    main()