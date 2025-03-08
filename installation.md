I suggest
rmq_exporter 0700
config.yaml 0700

`nano /etc/systemd/system/rmq_queue_exporter.service`
```
[Unit]
Description=rmq_queue_exporter
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/root/rmq_exporter_dir/
ExecStart=bash /root/rmq_exporter_dir/rmq_exporter
Restart=always

[Install]
WantedBy=multi-user.target
```

`systemctl enable rmq_queue_exporter`
`systemctl daemon-reload`
`systemctl start rmq_queue_exporter`
`systemctl status rmq_queue_exporter`
