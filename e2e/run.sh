#!/bin/bash

# Step 1: Start the beacon_proxy and ssv-node services
docker compose up -d --build beacon_proxy ssv-node-1 ssv-node-2 ssv-node-3 ssv-node-4

# Step 2: Run logs_catcher in Mode Slashing
docker compose run --build logs_catcher logs-catcher --mode Slashing

# Step 3: Stop the services
docker compose down

# Step 4. Run share_update for non leader
docker compose run --build share_update share-update --validator-pub-key "8c5801d7a18e27fae47dfdd99c0ac67fbc6a5a56bb1fc52d0309626d805861e04eaaf67948c18ad50c96d63e44328ab0" --db-path "/ssv-node-2-data/db" --operator-private-key "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcFFJQkFBS0NBUUVBd291VzR1SFRSMUQrdnYvYUU0Q09JeFQ1T0x5SGIwR0dieE01MDhQci82aW1XakRFCmxUUWR5Z2ZKblFld05Vb2FzdmthNEYzU2NxaTZTVGZDUG00a1V1RkY5YllsdXl3Q1ZKcUlpdjNVMW5acmlIbVQKTVRWNEpxaWxxVktIeU9HYlJKZUErQXdNZGRUR1pKTHRjenRXRTY4SUlzWkcwN1ZQQnpjQlJPdXVCWDNtQzNoSApjcVdjbzRFRlV0RHZGbXNKbEo2ZHB6SzYzcWREaFNIUXVWak5UbnhmRHpEV1BqK2RkMERzSWh6NCtDc3pEMzVCCnFFQ293a0REMCtiaXhkWEF5NTZsVi9JcUgxc3RCclEwQ2JnczJSbzlLbEVmc3VLbTNWdlFyejVQanhaUU9kc2cKMXhLS3JsZU8vQUw0T2lkeWV1aXBVM3IxQy90Yk9USjF1S3hSS1FJREFRQUJBb0lCQVFDSzlSKzZRT2tqaUhQZApRMnlsLzI0SEd1VUVwSXpzWjlZNUluZHNqZ1hVbjhicXB1alRWZDF0UC9DL0xBMnRrcGZOZkdhNUdlckdvVVFtCkppQ2xiUkNlN20rRkdTeU1LOXdpU0JyOWhGN3hMTGFVVFpwWVRNUGNnUnVLL1BzbC9oZGtmLzdMcmZkOGRwV2EKb3VQZUtlVEt2SHZJTXUzR0xEd2RnQ2wwN0E1cHRvdEh5WnpnckZuUlhpb1hRdmVqdnphYTJ2RDNKVWJKRk5QQQpsczZhejRqZWMrV2pyTnBQRnh5SWJHYzBNb2NZRVZVVC9CTlhZQmliNXc4M21kczJkM2x2RWpNNkJ2eWhBWGRTCkV2dGIyTTAwK0hUdzJOTlpjMXA3QXptUmZabG1wRHZmYVZCR0gwcTE4Q1R3eXZ0Q1JqZmkxWjNRZTFROTZCa3UKVFFFS3B3emRBb0dCQU5PaUgrcnlZWWxvKytTVzRRTlFKZytLQVh5Y0hJNzJtbTlkaTZwcElPbGQyc0FvaWM4UQpRbUV3VkhEUlZiMGRJNEdRYXlXeERYcGRTMFkrcWREV2JWTlN4YnY1ZFFKRTE1b0FKbm9sNWpqZ05la2RzVFdoCkFoU1pCb1ZXN2djc0ExVGJSa0VnamRIUzFTQmRoT3g3Wm5TV3BiT1BaMmttcXA4bGk3TStpV0kzQW9HQkFPdFUKWTZsbGNoZW1IODROOG02RXhxRU5nbDhJSERuUkxvUjNhc0R3dDBWOHJYbi85ditXMmVTRmtiZlRGRkNnMkZ3dQpiazRmZkhhV2g2dkNCYXdRNnI2RVJnMENEZGNTNzErZkU2ai9hNWZiWlNGR2V2b2J1bWtTN0pkSlNqYUdTNU5oCjFaN253YVU3OFErbDlTd20zeXJ3RitKYlN4bWM2dU9EL0NuL2RiZWZBb0dCQUx2d3hoZUhtRWJIREtzN3NpZVgKRGJYUEFQTUFUL3hGMDNBQ3l2MVN6djl2Y2N3a00vM0dtcXhrbHhoNVRvTGJWYVRCOCtWTkRvTVVScnppK1R1VQpzUkhGK0FPdXpOSnZBR2lxcVlEZ0YwdDdFV1VzRVN0bkNNbngrM0IrZW5PMENtRlpPVktzN2tUZnpwVW5kOXZxCjJsbS9UdmZlNmg3ZlQ3WjFTVktzdnFTUkFvR0FBYUJxcy9BTWt0ZEdId0Yvckgza2RaYUhVU3JZTHhvZ0RUQmEKSDQxS1p3T09tMnBHaGN2QUk5RThpWjIrNVRQSGF4T3pGWDBvT2hXZVNIU2wzMk9haThpVVIyQzlRY0JTd1VGegpQRmJQb3BRVXBkODcyR0M2c0NFK1cybFpSdmswcW9jaGwrQ1lPUkVxQUdhd1JDYmNvZ3BZeitxN29TaXhndk1WCm1pQzI2cGNDZ1lFQWpGejZVczhzczFPTmp1VEJ0MVozZFZXT0d4aVBSd1l2VTd4RWhJcWgzNUFZdWY4TWVSTzMKNDdwcThSc2Q0TnZiWnE0NE1JTmtZVXFsdUY4dmZCWE1oYWVURGtLNmhpSWo1TGNZOUNEbndtRFJ6a2kyTU9sMQo3UkdpUFhYNE05NVVUeEJwME5MZC9sTDdETVAxTUpMU2F0VmhHUEpmODZ0ci9kYk9NbFgrOHh3PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo="

# Step 5. Run share_update for leader
docker compose run --build share_update share-update --validator-pub-key "a238aa8e3bd1890ac5def81e1a693a7658da491ac087d92cee870ab4d42998a184957321d70cbd42f9d38982dd9a928c" --db-path "/ssv-node-2-data/db" --operator-private-key "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcFFJQkFBS0NBUUVBd291VzR1SFRSMUQrdnYvYUU0Q09JeFQ1T0x5SGIwR0dieE01MDhQci82aW1XakRFCmxUUWR5Z2ZKblFld05Vb2FzdmthNEYzU2NxaTZTVGZDUG00a1V1RkY5YllsdXl3Q1ZKcUlpdjNVMW5acmlIbVQKTVRWNEpxaWxxVktIeU9HYlJKZUErQXdNZGRUR1pKTHRjenRXRTY4SUlzWkcwN1ZQQnpjQlJPdXVCWDNtQzNoSApjcVdjbzRFRlV0RHZGbXNKbEo2ZHB6SzYzcWREaFNIUXVWak5UbnhmRHpEV1BqK2RkMERzSWh6NCtDc3pEMzVCCnFFQ293a0REMCtiaXhkWEF5NTZsVi9JcUgxc3RCclEwQ2JnczJSbzlLbEVmc3VLbTNWdlFyejVQanhaUU9kc2cKMXhLS3JsZU8vQUw0T2lkeWV1aXBVM3IxQy90Yk9USjF1S3hSS1FJREFRQUJBb0lCQVFDSzlSKzZRT2tqaUhQZApRMnlsLzI0SEd1VUVwSXpzWjlZNUluZHNqZ1hVbjhicXB1alRWZDF0UC9DL0xBMnRrcGZOZkdhNUdlckdvVVFtCkppQ2xiUkNlN20rRkdTeU1LOXdpU0JyOWhGN3hMTGFVVFpwWVRNUGNnUnVLL1BzbC9oZGtmLzdMcmZkOGRwV2EKb3VQZUtlVEt2SHZJTXUzR0xEd2RnQ2wwN0E1cHRvdEh5WnpnckZuUlhpb1hRdmVqdnphYTJ2RDNKVWJKRk5QQQpsczZhejRqZWMrV2pyTnBQRnh5SWJHYzBNb2NZRVZVVC9CTlhZQmliNXc4M21kczJkM2x2RWpNNkJ2eWhBWGRTCkV2dGIyTTAwK0hUdzJOTlpjMXA3QXptUmZabG1wRHZmYVZCR0gwcTE4Q1R3eXZ0Q1JqZmkxWjNRZTFROTZCa3UKVFFFS3B3emRBb0dCQU5PaUgrcnlZWWxvKytTVzRRTlFKZytLQVh5Y0hJNzJtbTlkaTZwcElPbGQyc0FvaWM4UQpRbUV3VkhEUlZiMGRJNEdRYXlXeERYcGRTMFkrcWREV2JWTlN4YnY1ZFFKRTE1b0FKbm9sNWpqZ05la2RzVFdoCkFoU1pCb1ZXN2djc0ExVGJSa0VnamRIUzFTQmRoT3g3Wm5TV3BiT1BaMmttcXA4bGk3TStpV0kzQW9HQkFPdFUKWTZsbGNoZW1IODROOG02RXhxRU5nbDhJSERuUkxvUjNhc0R3dDBWOHJYbi85ditXMmVTRmtiZlRGRkNnMkZ3dQpiazRmZkhhV2g2dkNCYXdRNnI2RVJnMENEZGNTNzErZkU2ai9hNWZiWlNGR2V2b2J1bWtTN0pkSlNqYUdTNU5oCjFaN253YVU3OFErbDlTd20zeXJ3RitKYlN4bWM2dU9EL0NuL2RiZWZBb0dCQUx2d3hoZUhtRWJIREtzN3NpZVgKRGJYUEFQTUFUL3hGMDNBQ3l2MVN6djl2Y2N3a00vM0dtcXhrbHhoNVRvTGJWYVRCOCtWTkRvTVVScnppK1R1VQpzUkhGK0FPdXpOSnZBR2lxcVlEZ0YwdDdFV1VzRVN0bkNNbngrM0IrZW5PMENtRlpPVktzN2tUZnpwVW5kOXZxCjJsbS9UdmZlNmg3ZlQ3WjFTVktzdnFTUkFvR0FBYUJxcy9BTWt0ZEdId0Yvckgza2RaYUhVU3JZTHhvZ0RUQmEKSDQxS1p3T09tMnBHaGN2QUk5RThpWjIrNVRQSGF4T3pGWDBvT2hXZVNIU2wzMk9haThpVVIyQzlRY0JTd1VGegpQRmJQb3BRVXBkODcyR0M2c0NFK1cybFpSdmswcW9jaGwrQ1lPUkVxQUdhd1JDYmNvZ3BZeitxN29TaXhndk1WCm1pQzI2cGNDZ1lFQWpGejZVczhzczFPTmp1VEJ0MVozZFZXT0d4aVBSd1l2VTd4RWhJcWgzNUFZdWY4TWVSTzMKNDdwcThSc2Q0TnZiWnE0NE1JTmtZVXFsdUY4dmZCWE1oYWVURGtLNmhpSWo1TGNZOUNEbndtRFJ6a2kyTU9sMQo3UkdpUFhYNE05NVVUeEJwME5MZC9sTDdETVAxTUpMU2F0VmhHUEpmODZ0ci9kYk9NbFgrOHh3PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo="

# Step 6: Start the beacon_proxy and ssv-nodes again
docker compose up -d beacon_proxy ssv-node-1 ssv-node-2 ssv-node-3 ssv-node-4

# Step 7: Run logs_catcher in Mode BlsVerification for non leader
docker compose run --build logs_catcher logs-catcher --mode BlsVerification --leader 1

# Step 8: Run logs_catcher in Mode BlsVerification for leader
# TODO: Fix
#docker compose run --build logs_catcher logs-catcher --mode BlsVerification --leader 2