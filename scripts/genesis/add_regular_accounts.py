import json
import csv
import sys
from decimal import Decimal

DENOM = "aswtr"
CHAIN_ID = "swisstronik_1848-1"
DECIMALS = 10**18

def load_csv_accounts(csv_path):
    with open(csv_path, 'r') as f:
        reader = csv.DictReader(f)
        return [
            {
                "address": row['address'].strip(),
                "balance": int(row['balance'].strip())
            }
            for row in reader
        ]

def get_next_account_number(genesis):
    account_numbers = []
    for acc in genesis['app_state']['auth']['accounts']:
        if acc.get('@type') == "/cosmos.auth.v1beta1.BaseAccount":
            try:
                account_numbers.append(int(acc.get('account_number', '0')))
            except (ValueError, TypeError):
                continue
    return max(account_numbers, default=0) + 1

def convert_amount(amount_str):
    raw = Decimal(amount_str.strip().split(DENOM)[0])
    return str(int(raw * DECIMALS))

def add_accounts_to_genesis(genesis_path, csv_path):
    with open(genesis_path, 'r') as f:
        genesis = json.load(f)

    # Safety check: chain-id
    if genesis.get("chain_id") != CHAIN_ID:
        print(f"⚠️  Warning: Expected chain-id '{CHAIN_ID}', found '{genesis.get('chain_id')}'")

    existing_addresses = {
        acc.get('address') for acc in genesis['app_state']['auth']['accounts']
        if acc.get('@type') == "/cosmos.auth.v1beta1.BaseAccount" and 'address' in acc
    }
    new_accounts = load_csv_accounts(csv_path)

    next_account_number = get_next_account_number(genesis)
    total_new_balance = 0
    added = 0

    for acc in new_accounts:
        addr = acc['address']
        bal = convert_amount(str(acc['balance']))

        if addr in existing_addresses:
            print(f"Skipping existing account: {addr}")
            continue

        # Add to auth.accounts
        genesis['app_state']['auth']['accounts'].append({
            "@type": "/cosmos.auth.v1beta1.BaseAccount",
            "address": addr,
            "pub_key": None,
            "account_number": str(next_account_number),
            "sequence": "0"
        })
        next_account_number += 1

        # Add to bank.balances
        genesis['app_state']['bank']['balances'].append({
            "address": addr,
            "coins": [
                {
                    "denom": DENOM,
                    "amount": bal
                }
            ]
        })

        total_new_balance += int(bal)
        added += 1

    # Update total supply
    supply = genesis['app_state']['bank'].get('supply', [])
    for coin in supply:
        if coin['denom'] == DENOM:
            coin['amount'] = str(int(coin['amount']) + total_new_balance)
            break
    else:
        # Denom not found, add new
        supply.append({
            "denom": DENOM,
            "amount": str(total_new_balance)
        })
    genesis['app_state']['bank']['supply'] = supply

    with open(genesis_path, 'w') as f:
        json.dump(genesis, f, indent=2)

    print(f"✅ Added {added} new accounts. New total supply: +{total_new_balance} {DENOM}")

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python add_genesis_accounts.py <path_to_genesis.json> <path_to_accounts.csv>")
    else:
        add_accounts_to_genesis(sys.argv[1], sys.argv[2])
