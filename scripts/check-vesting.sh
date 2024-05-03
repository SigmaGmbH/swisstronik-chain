#!/bin/bash

HOMEDIR=~/.swisstronik
KEYRING=test

set -e

function wait_for_tx () {
    echo ""
    echo "Waiting for sync tx"
    sleep 3 # wait 3 seconds
}

# Step1
# Add funder account into keyring
echo "cable flee torch mimic roof humble phone harsh wrist blade prevent cook weasel head south task toe artwork thunder gap siren disease scrap easily" | swisstronikd keys add funder --keyring-backend $KEYRING --home $HOMEDIR --recover
FUNDER=$(swisstronikd keys show funder -a --keyring-backend $KEYRING --home $HOMEDIR)
# Add vesting account into keyring(not necessary to add, just only for setup address variable)
echo "grass rely robot nasty trade hidden car total pride often area dolphin hand sad spider pudding burst shallow across brisk exhibit salute myself interest" | swisstronikd keys add vesting_account --recover --keyring-backend $KEYRING --home $HOMEDIR
VA=$(swisstronikd keys show vesting_account -a --keyring-backend $KEYRING --home $HOMEDIR) # Vesting account as a new account

echo "funder=$FUNDER"
echo "vesting account(va)=$VA"

# Step2
# Funds tokens to `FUNDER`` for gas consuming & vesting
echo "funding tokens to funder 1000swtr for gas consuming & vesting..."
swisstronikd tx bank send bob $FUNDER 1000swtr -y --gas-prices 1000000000aswtr

wait_for_tx

echo "initial funder balance"
swisstronikd query bank balances $FUNDER --output json | jq '.balances[0].amount'

echo ""
# Step3
# Create monthly vesting account of cliff days + 3 months
# As an example for demo, `swisstronikd` was built with 1 day as 1 seconds, 1 month as 30 seconds
ONE_DAY=3
ONE_MONTH=$((ONE_DAY*30))
CLIFF=30 # 30 seconds
MONTHS=3 # 90 seconds
OV=3swtr # original vesting coin amount, 10^18aswtr
echo "CLIFF=$CLIFF, MONTHS=$MONTHS, OV=$OV"
echo "CLIFF_IN_SECONDS=$(($CLIFF*$ONE_DAY))sec, MONTHS_IN_SECONDS=$(($MONTHS*$ONE_MONTH))sec"
echo "creating monthly vesting account of cliff days as 30 seconds, 3 months as 90 seconds..."
swisstronikd tx vesting create-monthly-vesting-account $VA $CLIFF $MONTHS $OV --from $FUNDER --keyring-backend $KEYRING --home $HOMEDIR -y --gas-prices 1000000000aswtr

# Send some funds for delegation tx cost
swisstronikd tx bank send bob $VA 1swtr -y --gas-prices 1000000000aswtr

wait_for_tx

echo ""
# Check if vesting account was created
echo "querying account $VA"
swisstronikd query account $VA

echo ""
# Check balances of accounts
# It should immediately move `OV` from `FUNDER` to `VA`.
# Balance of `FUNDER` should be reduced by `OV`.
echo "checking balances of accounts..."
echo "balances of funder should be reduced by $OV and gas fees"
swisstronikd query bank balances $FUNDER --output json | jq '.balances[0].amount'
# Balance of `VA` should be `OV`.
echo "balances of va should be $OV"
swisstronikd query bank balances $VA --output json | jq '.balances[0].amount'

echo ""
# Check vesting balances
# All the vesting coins should be locked and unvested, there's no vested amount before cliff
echo "checking vesting balances of va before cliff days..."
echo "all the initial vesting should be locked and unvested, nothing vested"
swisstronikd query vesting balances $VA --output json | jq '{locked:.locked[0].amount, unvested:.unvested[0].amount, vested:.vested[0].amount}'

echo ""
# Check spendable coins of `VA`, should be 0 until vested
echo "checking spendable coins of va before cliff days..."
echo "should be no spendable coins"
swisstronikd query bank spendable-balances $VA --output json | jq '.balances[0].amount'

# Should be able to delegate locked coins
echo ""
echo "Trying to delegate locked coins"
VALIDATOR=$(swisstronikd q staking validators --output json | jq -r '.validators[0].operator_address')
echo $VALIDATOR
swisstronikd tx staking delegate $VALIDATOR 3swtr --gas-prices 1000000000aswtr --from vesting_account -y --gas 250000
wait_for_tx
echo "Check delegations"
swisstronikd q staking delegations $VA

# Try to unbond locked coins
swisstronikd tx staking unbond $VALIDATOR 3swtr --gas-prices 1000000000aswtr --from vesting_account -y --gas 250000

echo ""
# Step4
# Wait for cliff days (30 seconds)
echo "waiting for cliff days..."
sleep $((CLIFF*ONE_DAY))

echo ""
# Check vesting balances
echo "checking vesting balances of va after cliff days..."
echo "all the initial vesting should be locked and unvested, nothing vested"
swisstronikd query vesting balances $VA --output json | jq '{locked:.locked[0].amount, unvested:.unvested[0].amount, vested:.vested[0].amount}'

echo ""
# Check spendable coinds of `VA`
echo "checking spendable coins of va after cliff days..."
echo "should be no spendable coins after cliff days"
swisstronikd query bank spendable-balances $VA --output json | jq '.balances[0].amount'

echo "regular q bank balances output"
swisstronikd query bank balances $VA --output json

echo ""
# Step5.1
# Wait for first month
echo "waiting for first vesting period..."
sleep $((ONE_MONTH))

echo ""
# Check vesting balances
echo "checking vesting balances of va after first month..."
echo "1/3 of initial vesting should be vested, the rest are unvested and locked"
swisstronikd query vesting balances $VA --output json | jq '{locked:.locked[0].amount, unvested:.unvested[0].amount, vested:.vested[0].amount}'

echo ""
# Check spendable coinds of `VA`
echo "checking spendable coins of va after first vesting period..."
echo "vested coins should be spendable"
swisstronikd query bank spendable-balances $VA --output json | jq '.balances[0].amount'

echo ""
# Step5.2
# Wait for second month
echo "waiting for second vesting period..."
sleep $((ONE_MONTH))

echo ""
# Check vesting balances
echo "checking vesting balances of va after second month..."
echo "2/3 of initial vesting should be vested, the rest are unvested and locked"
swisstronikd query vesting balances $VA --output json | jq '{locked:.locked[0].amount, unvested:.unvested[0].amount, vested:.vested[0].amount}'

echo ""
# Check spendable coinds of `VA`
echo "checking spendable coins of va after second vesting period..."
echo "vested coins should be spendable"
swisstronikd query bank spendable-balances $VA --output json | jq '.balances[0].amount'


echo ""
# Step5.3
# Wait for third month
echo "waiting for third vesting period..."
sleep $((ONE_MONTH))

echo ""
# Check vesting balances
echo "checking vesting balances of va after third month..."
echo "full initial vesting should be vested"
swisstronikd query vesting balances $VA --output json | jq '{locked:.locked[0].amount, unvested:.unvested[0].amount, vested:.vested[0].amount}'

echo ""
# Check spendable coinds of `VA`
echo "checking spendable coins of va after first vesting period..."
echo "vested coins should be spendable"
swisstronikd query bank spendable-balances $VA --output json | jq '.balances[0].amount'


echo ""
# Check vesting balances
echo "checking vesting balances of va at the end of vesting period..."
echo "all the initial vesting should be vested, nothing is unvested and locked"
swisstronikd query vesting balances $VA --output json | jq '{locked:.locked[0].amount, unvested:.unvested[0].amount, vested:.vested[0].amount}'

echo ""
# Check spendable coinds of `VA`
echo "checking spendable coins of va at the end of vesting period..."
echo "all the inital vesting coins should be spendable"
swisstronikd query bank spendable-balances $VA --output json | jq '.balances[0].amount'

