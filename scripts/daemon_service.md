sudo nano /etc/systemd/system/swisstronikd.service

# Write the followings by replacing <your-user> with the one of your instance.
```
[Unit]
Description=Swisstronik Daemon (cosmovisor)
After=network-online.target

[Service]
User=<your-user>
ExecStart=/home/<your-user>/go/bin/cosmovisor run start --pruning="nothing" --rpc.laddr "tcp://0.0.0.0:26657" --enclave.address "0.0.0.0:8999" --json-rpc.address 0.0.0.0:8545 --json-rpc.ws-address 0.0.0.0:8546
Restart=always
RestartSec=3
LimitNOFILE=65536
ProtectHome = true;
ProtectSystem = "strict";
PrivateTmp = true;
ProtectHostname = true;
ProtectKernelTunables = true;
ProtectKernelModules = true;
ProtectKernelLogs = true;
ProtectControlGroups = true;
NoNewPrivileges = true;
RestrictRealtime = true;
RestrictSUIDSGID = true;
RemoveIPC = true;
PrivateMounts = true;
Environment="DAEMON_NAME=swisstronikd"
Environment="DAEMON_HOME=/home/<your-user>/.swisstronik"
Environment="DAEMON_ALLOW_DOWNLOAD_BINARIES=false"
Environment="DAEMON_RESTART_AFTER_UPGRADE=true"
Environment="DAEMON_LOG_BUFFER_SIZE=512"

[Install]
WantedBy=multi-user.target
```