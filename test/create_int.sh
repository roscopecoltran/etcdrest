curl -v -H "Content-Type: application/json" -X PUT -d "$(cat ./test_int.json)" 127.0.0.1:8080/hosts/test.example.com/interfaces/eth0
