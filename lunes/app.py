import os
import subprocess
import threading
from flask import Flask

# 环境变量配置；填入 ARGO_DOMAIN 和 ARGO_AUTH 使用固定隧道
os.environ.setdefault('FILE_PATH', '.cache')
os.environ.setdefault('UUID', '0937e1b1-9065-4b70-8562-0fe44ae558da')
os.environ.setdefault('ARGO_DOMAIN', '')
os.environ.setdefault('ARGO_AUTH', '')
os.environ.setdefault('ARGO_PORT', '8001')
os.environ.setdefault('CFIP', 'saas.sin.fan')
os.environ.setdefault('CFPORT', '443')
os.environ.setdefault('NAME', '')

app = Flask(__name__)

@app.route('/')
def home():
    return "Service is running perfectly!"

def start_go_binary():
    go_bin = os.path.join(os.getcwd(), 'main')
    if os.path.exists(go_bin):
        try:
            os.chmod(go_bin, 0o775)
            # 继承环境变量并启动 Go 二进制
            subprocess.Popen([go_bin], env=os.environ.copy())
            print("Successfully launched ./main Go binary in background!")
        except Exception as e:
            print(f"Error launching ./main: {e}")

# 在后台守护线程中启动 Go 二进制
threading.Thread(target=start_go_binary, daemon=True).start()

if __name__ == '__main__':
    port = int(os.environ.get('PORT', '3000'))
    app.run(host='0.0.0.0', port=port)