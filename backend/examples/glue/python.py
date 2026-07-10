#!/usr/bin/env python3
# GLUE-Python 任务示例：GLUE 类型选 Python，粘贴本脚本。sys.argv[1] 为任务参数。
import sys, datetime, os

param = sys.argv[1] if len(sys.argv) > 1 else ""
print(f"[{datetime.datetime.now()}] GLUE Python 开始，参数: {param}")
print(f"Python: {sys.version.split()[0]}  PID: {os.getpid()}")
for i in range(3):
    print(f"  step {i+1}/3")
print(f"[{datetime.datetime.now()}] GLUE Python 结束")
sys.exit(0)
