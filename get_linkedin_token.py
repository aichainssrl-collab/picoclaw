import os
import requests
import urllib.parse
import argparse
import json
from dotenv import load_dotenv

# Carica le variabili d'ambiente dal file .env
load_dotenv()

CLIENT_ID = os.getenv('LINKEDIN_CLIENT_ID')
CLIENT_SECRET = os.getenv('LINKEDIN_CLIENT_SECRET')
REDIRECT_URI = os.getenv('LINKEDIN_REDIRECT_URI')
SCOPE = 'w_member_social w_organization_social rw_organization_admin openid profile email'

def get_authorization_url():
    """Genera l'URL per l'autorizzazione dell'utente."""
    base_url = "https://www.linkedin.com/oauth/v2/authorization"
    params = {
        'response_type': 'code',
        'client_id': CLIENT_ID,
        'redirect_uri': REDIRECT_URI,
        'state': 'random_string_xyz', # In produzione, usa un valore casuale sicuro
        'scope': SCOPE
    }
    url = f"{base_url}?{urllib.parse.urlencode(params)}"
    return url

def exchange_code_for_token(authorization_code):
    """Scambia il codice di autorizzazione con un access token."""
    token_url = "https://www.linkedin.com/oauth/v2/accessToken"
    data = {
        'grant_type': 'authorization_code',
        'code': authorization_code,
        'redirect_uri': REDIRECT_URI,
        'client_id': CLIENT_ID,
        'client_secret': CLIENT_SECRET
    }
    
    headers = {
        'Content-Type': 'application/x-www-form-urlencoded'
    }

    try:
        response = requests.post(token_url, data=data, headers=headers)
        response.raise_for_status()
        return response.json()
    except requests.exceptions.RequestException as e:
        if 'args' not in globals() or (args.get_url or args.code):
            # Silent error for JSON output mode, let main handle it
            pass
        else:
            print(f"Errore durante la richiesta del token: {e}")
            if hasattr(response, 'text') and response.text:
                print(f"Dettagli errore: {response.text}")
        return None

def main():
    parser = argparse.ArgumentParser(description='LinkedIn Token Helper')
    parser.add_argument('--get-url', action='store_true', help='Print authorization URL only')
    parser.add_argument('--code', help='Exchange code for token')
    global args
    args = parser.parse_args()

    if not CLIENT_ID or not CLIENT_SECRET or not REDIRECT_URI:
        if args.get_url or args.code:
            print(json.dumps({"error": "Missing env vars"}))
            return
        print("Errore: Assicurati che LINKEDIN_CLIENT_ID, LINKEDIN_CLIENT_SECRET e LINKEDIN_REDIRECT_URI siano impostati nel file .env")
        return

    if args.get_url:
        print(json.dumps({"url": get_authorization_url()}))
        return

    if args.code:
        token_data = exchange_code_for_token(args.code)
        if token_data:
            print(json.dumps(token_data))
        else:
            print(json.dumps({"error": "Failed to retrieve token"}))
        return

    # Interactive mode
    print("--- Recupero Automatico Token LinkedIn ---")
    print("\n1. Visita questo URL nel tuo browser per autorizzare l'app:")
    print(f"\n{get_authorization_url()}\n")
    
    print("2. Dopo l'autorizzazione, verrai reindirizzato a un URL che appare come:")
    print(f"   {REDIRECT_URI}?code=AUTHORIZATION_CODE&state=...")
    
    auth_code = input("\n3. Incolla qui il 'code' ricevuto (tutto ciò che c'è dopo 'code=' e prima di '&state='): ").strip()
    
    if not auth_code:
        print("Codice non fornito. Uscita.")
        return

    print("\nTentativo di recupero del token...")
    token_data = exchange_code_for_token(auth_code)
    
    if token_data and 'access_token' in token_data:
        access_token = token_data['access_token']
        expires_in = token_data.get('expires_in', 'N/A')
        print("\n✅ Access Token recuperato con successo!")
        print(f"Token: {access_token}")
        print(f"Scade tra: {expires_in} secondi")
        
        # Opzionale: Salva o aggiorna il .env
        print("\nCopia questo token nel tuo file .env come LINKEDIN_ACCESS_TOKEN")
    else:
        print("\n❌ Impossibile recuperare il token.")

if __name__ == "__main__":
    main()
