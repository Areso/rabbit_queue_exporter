# rabbit_queue_exporter
Solution to monitor RabbitMQ queues (consumers and messages)

go mod init rmq_exporter  
go mod tidy  
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o rmq_exporter . && chmod 0744 rmq_exporter  
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o rmq_exporter_arm . && chmod 0744 rmq_exporter_arm  
//GOOS=linux GOARCH=amd64 go build -o your_binary main.go  


scp ~/git/dbaas/rabbit_queue_exporter/rmq_exporter root@server:/root/rmq_exporter_dir/rmq_exporter   
