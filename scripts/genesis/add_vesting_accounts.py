import json
import csv
import sys
import os
from decimal import Decimal

DENOM = "aswtr"
SECONDS_IN_DAY = 86400
SECONDS_IN_MONTH = 2592000  # 30 days
DECIMALS = 10**18

def load_csv_accounts(csv_path):
    with open(csv_path, 'r') as f:
        reader = csv.DictReader(f)
        return list(reader)

def load_genesis(path):
    with open(path, 'r') as f:
        return json.load(f)

def save_genesis(path, data):
    with open(path, 'w') as f:
        json.dump(data, f, indent=2)

def convert_amount(amount_str):
    raw = Decimal(amount_str.strip().split(DENOM)[0])
    return str(int(raw * DECIMALS))

def build_vesting_periods(amount_per_month, months):
    return [
        {
            "amount": [{"amount": amount_per_month, "denom": DENOM}],
            "length": str(SECONDS_IN_MONTH)
        } for _ in range(int(months))
    ]

def remove_existing_account(genesis, address):
    accounts = genesis['app_state']['auth']['accounts']
    updated = [
        acc for acc in accounts
        if not (acc.get('@type') == "/swisstronik.vesting.MonthlyVestingAccount" and
                acc.get('base_vesting_account', {}).get('base_account', {}).get('address') == address)
    ]
    if len(accounts) != len(updated):
        print(f"\033[31m[DEL]\033[0m Removed existing vesting account: {address}")
    genesis['app_state']['auth']['accounts'] = updated

def add_bank_balance(genesis, address, amount):
    balances = genesis['app_state']['bank']['balances']
    balances.append({
        "address": address,
        "coins": [{"denom": DENOM, "amount": amount}]
    })

def add_vesting_account(genesis, address, start_time, cliff_time, end_time, total_vesting, spendable, months):
    vesting_amount_per_month = str(int(Decimal(total_vesting) // int(months)))
    corrected_total_vesting = str(int(vesting_amount_per_month) * int(months))
    vesting_periods = build_vesting_periods(vesting_amount_per_month, months)

    # if total_vesting % month != 0, add the remainder to the last month
    remainder = int(total_vesting) - int(corrected_total_vesting)
    if remainder > 0:
        print("DEBUG: Adding remainder to the last month. Original amount: ", total_vesting, "Corrected:", corrected_total_vesting, "Remainder:", remainder)
        vesting_periods[-1]['amount'][0]['amount'] = str(int(vesting_periods[-1]['amount'][0]['amount']) + remainder)

    account = {
        "@type": "/swisstronik.vesting.MonthlyVestingAccount",
        "base_vesting_account": {
            "base_account": {
                "address": address,
                "pub_key": None,
                "sequence": "0"
            },
            "delegated_free": [],
            "delegated_vesting": [],
            "end_time": str(end_time),
            "original_vesting": [{
                "amount": total_vesting,
                "denom": DENOM
            }]
        },
        "cliff_time": str(cliff_time),
        "start_time": str(start_time),
        "vesting_periods": vesting_periods
    }

    genesis['app_state']['auth']['accounts'].append(account)
    print(f"\033[32m[ADD]\033[0m Added vesting account for address: {address}")
    return str(int(corrected_total_vesting) + int(spendable) + remainder)  # total for bank

def main():
    if len(sys.argv) != 4:
        print("Usage: python add_vesting_accounts.py <start_timestamp> <csv_file_path> <genesis_file_path>")
        sys.exit(1)

    start_time = int(sys.argv[1])
    csv_path = sys.argv[2]
    genesis_path = sys.argv[3]

    genesis = load_genesis(genesis_path)
    entries = load_csv_accounts(csv_path)

    for row in entries:
        if row['address'] == 'address':
            continue

        address = row['address'].strip()
        vesting_amount = convert_amount(row['original_vesting'])
        cliff_days = int(row['cliff_days'])
        months = int(row['months'])
        spendable = convert_amount(row['spendable'])

        cliff_time = start_time + cliff_days * SECONDS_IN_DAY
        end_time = cliff_time + months * SECONDS_IN_MONTH

        remove_existing_account(genesis, address)

        total_amount = add_vesting_account(
            genesis, address, start_time, cliff_time, end_time,
            vesting_amount, spendable, months
        )
        add_bank_balance(genesis, address, total_amount)

    save_genesis(genesis_path, genesis)
    print(f"\n✅ Vesting accounts added to {genesis_path}")

if __name__ == "__main__":
    main()
