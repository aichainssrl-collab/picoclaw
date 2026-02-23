import unittest
import json
import os
import sys

# Add workspace/news to path to import update_crypto if needed, 
# but here we just test the output file
sys.path.append(os.path.join(os.path.dirname(__file__), '../workspace/news'))

class TestCryptoBTCChange(unittest.TestCase):
    def test_crypto_json_btc_change(self):
        crypto_file = os.path.join(os.path.dirname(__file__), '../workspace/news/crypto.json')
        
        # Ensure file exists
        self.assertTrue(os.path.exists(crypto_file), "crypto.json does not exist")
        
        with open(crypto_file, 'r') as f:
            data = json.load(f)
            
        self.assertIn('currencies', data, "crypto.json missing 'currencies' key")
        self.assertTrue(len(data['currencies']) > 0, "No currencies found in crypto.json")
        
        found_valid_btc_change = False
        for coin in data['currencies']:
            self.assertIn('change_24h_btc', coin, f"coin {coin.get('name')} missing 'change_24h_btc'")
            val = coin['change_24h_btc']
            print(f"Coin: {coin['name']}, BTC Change: {val}")
            
            # BTC itself might be 0.00%
            if coin['symbol'] == 'BTC':
                self.assertEqual(val, '+0.00%', "BTC change should be +0.00%")
            
            # Check format
            if val != "N/A":
                found_valid_btc_change = True
                self.assertTrue(val.endswith('%'), f"Invalid format for {coin['name']}: {val}")
                
        self.assertTrue(found_valid_btc_change, "All coins have N/A for change_24h_btc (except maybe BTC)")

if __name__ == '__main__':
    unittest.main()
