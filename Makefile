port=27021
instance=1
rs=rs1

# example: make start port=27019 instance=2
start:
	mkdir -p data/data$(instance) || true
	mongod --port $(port) --replSet $(rs) --dbpath data/data$(instance) --bind_ip localhost -f  data/mongod.conf

# init replica
init:
	mongosh --port $(port) --quiet ./data/replica-init.js
