#!/bin/bash

HOMEDIR=~/.swisstronik
KEYRING=test

set -e

function wait_for_tx () {
    echo -e "\nWaiting for sync tx...\n"
    sleep 3 # wait 3 seconds
}

function check_vesting_distribution () {
    # Check vesting balances
    echo "checking vesting balances..."
    swisstronikd query vesting balances $1 --output json | jq '{locked:.locked[0].amount, unvested:.unvested[0].amount, vested:.vested[0].amount}'

    # Check spendable coins
    echo -e "\nchecking spendable coins..."
    swisstronikd query bank spendable-balances $1 --output json | jq '.balances[0].amount'

    echo "regular bank balances output"
    swisstronikd query bank balances $1 --output json
}

######### STEP 1 #########
# Create keys for funder and vesting_account
echo -e "\nStep 1"
echo "cable flee torch mimic roof humble phone harsh wrist blade prevent cook weasel head south task toe artwork thunder gap siren disease scrap easily" | swisstronikd keys add funder --keyring-backend $KEYRING --home $HOMEDIR --recover > /dev/null 2>&1
echo "grass rely robot nasty trade hidden car total pride often area dolphin hand sad spider pudding burst shallow across brisk exhibit salute myself interest" | swisstronikd keys add vesting_account --recover --keyring-backend $KEYRING --home $HOMEDIR > /dev/null 2>&1
FUNDER_ADDRESS=$(swisstronikd keys show funder -a --keyring-backend $KEYRING --home $HOMEDIR)
VESTING_ACC_ADDRESS=$(swisstronikd keys show vesting_account -a --keyring-backend $KEYRING --home $HOMEDIR)
echo "Funder address: " $FUNDER_ADDRESS
echo "Vesting account address: " $VESTING_ACC_ADDRESS
echo -e "##########################\n"
##########################


######### STEP 2 #########
# Funds tokens to `FUNDER`` for gas consuming & vesting
echo -e "\nStep 2"
echo -e "\nFunding tokens to funder 4swtr for gas consuming & vesting..."
swisstronikd tx bank send bob $FUNDER_ADDRESS 4swtr -y --gas-prices 1000000000aswtr --output json | jq '.txhash'
wait_for_tx
echo "initial funder balance:" $(swisstronikd query bank balances $FUNDER_ADDRESS --output json | jq -r '.balances[0].amount')
echo -e "##########################\n"
##########################


######### STEP 3 #########
# Create monthly vesting account of cliff days + 3 months
# As an example for demo, `swisstronikd` was built with 1 day as 3 seconds, 1 month as 90 seconds
echo -e "\nStep 3"
ONE_DAY=3
ONE_MONTH=$((ONE_DAY*30))
CLIFF=30 # 90 seconds
MONTHS=3 # 270 seconds
ORIGINAL_VESTING_AMOUNT=3swtr # original vesting coin amount, 3 * 10^18aswtr
echo "Cliff days: $CLIFF, Months: $MONTHS, Vesting amount=$ORIGINAL_VESTING_AMOUNT"
echo "creating monthly vesting account of cliff days as 30 seconds, 3 months as 90 seconds..."
swisstronikd tx vesting create-monthly-vesting-account $VESTING_ACC_ADDRESS $CLIFF $MONTHS $ORIGINAL_VESTING_AMOUNT --from $FUNDER_ADDRESS --keyring-backend $KEYRING --home $HOMEDIR -y --gas-prices 1000000000aswtr --output json | jq '.txhash'
echo "send spendable 1swtr to vesting account..."
swisstronikd tx bank send bob $VESTING_ACC_ADDRESS 1swtr -y --gas-prices 1000000000aswtr --output json | jq '.txhash'

wait_for_tx
echo "querying account $VESTING_ACC_ADDRESS"
swisstronikd query account $VESTING_ACC_ADDRESS
# Check balances of accounts
# It should immediately move `ORIGINAL_VESTING_AMOUNT` from `FUNDER_ADDRESS` to `VESTING_ACC_ADDRESS`.
# Balance of `FUNDER_ADDRESS` should be reduced by `ORIGINAL_VESTING_AMOUNT`.
echo -e "\nchecking balances of accounts..."
echo "balances of funder should be reduced by $ORIGINAL_VESTING_AMOUNT and gas fees"
swisstronikd query bank balances $FUNDER_ADDRESS --output json | jq -r '.balances[0].amount'
# Balance of `VA` should be `OV`.
echo "balances of vesting account should be $ORIGINAL_VESTING_AMOUNT + 1 swtr"
swisstronikd query bank balances $VESTING_ACC_ADDRESS --output json | jq -r '.balances[0].amount'

# Check vesting balances
# All the vesting coins should be locked and unvested, there's no vested amount before cliff
echo -e "\nChecking vesting balances of vesting account before cliff days..."
echo "all the initial vesting should be locked and unvested, nothing vested"
swisstronikd query vesting balances $VESTING_ACC_ADDRESS --output json | jq '{locked:.locked[0].amount, unvested:.unvested[0].amount, vested:.vested[0].amount}'

# Check spendable coins of `VA`, should be 0 until vested
echo -e "\nChecking spendable coins of vesting account before cliff days..."
echo "Should be 1 swtr as spendable balance"
swisstronikd query bank spendable-balances $VESTING_ACC_ADDRESS --output json | jq '.balances[0].amount'
echo -e "##########################\n"
##########################

######### STEP 4 #########
# Wait for cliff days (90 seconds)
echo -e "\nStep 4"
echo "waiting for cliff days..."
sleep $((CLIFF*ONE_DAY))
echo "Checking vesting balances after cliff. All initial vesting coins should be locked and unvested. Should be only 1 swtr spendable"
check_vesting_distribution $VESTING_ACC_ADDRESS
echo -e "##########################\n"
##########################

######### STEP 5.1 #########
# Wait for first month (90 seconds)
echo -e "\nStep 5.1"
echo "waiting for first month..."
sleep $((ONE_MONTH))
echo "Checking vesting balances after first month. 1swtr + 1/3 of initial vesting should be spendable, the rest are unvested and locked"
check_vesting_distribution $VESTING_ACC_ADDRESS
echo -e "##########################\n"
##########################

######### STEP 5.2 #########
# Wait for second month (90 seconds)
echo -e "\nStep 5.2"
echo "waiting for second month..."
sleep $((ONE_MONTH))
echo "Checking vesting balances after second month. 1swtr + 2/3 of initial vesting should be spendable, the rest are unvested and locked"
check_vesting_distribution $VESTING_ACC_ADDRESS
echo -e "##########################\n"
##########################

######### STEP 5.3 #########
# Wait for third month (90 seconds)
echo -e "\nStep 5.3"
echo "waiting for third month..."
sleep $((ONE_MONTH))
echo "Checking vesting balances after second month. All funds should be accessible: 4 swtr"
check_vesting_distribution $VESTING_ACC_ADDRESS
echo -e "##########################\n"
##########################

# # Should be able to delegate locked coins
# echo ""
# echo "Trying to delegate locked coins"
# VALIDATOR=$(swisstronikd q staking validators --output json | jq -r '.validators[0].operator_address')
# echo $VALIDATOR
# swisstronikd tx staking delegate $VALIDATOR 3swtr --gas-prices 1000000000aswtr --from vesting_account -y --gas 250000
# wait_for_tx
# echo "Check delegations"
# swisstronikd q staking delegations $VA

# # Try to unbond locked coins
# swisstronikd tx staking unbond $VALIDATOR 3swtr --gas-prices 1000000000aswtr --from vesting_account -y --gas 250000