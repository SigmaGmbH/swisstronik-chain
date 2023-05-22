HOMEDIR="$HOME/.swisstronik"
mkdir -p $HOMEDIR/cosmovisor/genesis/bin && mkdir -p $HOMEDIR/cosmovisor/upgrades
cp $HOME/go/bin/swisstronikd $HOMEDIR/cosmovisor/genesis/bin
cosmovisor init $HOME/go/bin/swisstronikd
## Please follow daemon_service.md before executing the followings
sudo -S systemctl daemon-reload
sudo -S systemctl enable swisstronikd
# check config one last time before starting!
sudo systemctl start swisstronikd
sudo systemctl status swisstronikd
journalctl -fu swisstronikd