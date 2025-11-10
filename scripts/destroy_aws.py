import subprocess
import os
from pathlib import Path
from dotenv import load_dotenv

project_root = Path(__file__).parent.parent
load_dotenv(dotenv_path=project_root / '.env')

def main():
    env = os.environ.copy()

    env['TF_VAR_db_name'] = os.getenv('TF_VAR_db_name', 'flashsale')
    env['TF_VAR_db_username'] = os.getenv('TF_VAR_db_username', 'admin')
    env['TF_VAR_db_password'] = os.getenv('TF_VAR_db_password', 'SecurePassword123!')
    
    print("Destroying Terraform...")
    subprocess.run([
        "terraform", 
        "-chdir=flash-sale-platform/terraform", 
        "destroy",
        "-auto-approve"
    ], env=env, check=True)
    
    print("AWS infrastructure Destroyed!")

if __name__ == "__main__":
    main()