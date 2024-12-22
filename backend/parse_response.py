import json
import sys

response = json.load(sys.stdin)
print(response['url'])
