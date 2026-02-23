import os
import sys
import argparse
import requests
import json
from dotenv import load_dotenv

# Carica variabili d'ambiente
load_dotenv()

ACCESS_TOKEN = os.getenv('LINKEDIN_ACCESS_TOKEN') or os.getenv('LINKEDIN_TOKEN')
USER_URN = os.getenv('LINKEDIN_USER_URN') # Opzionale, se non c'è lo recuperiamo

def get_user_urn():
    """Recupera l'URN dell'utente se non presente in .env"""
    if USER_URN:
        return USER_URN
        
    url = "https://api.linkedin.com/v2/userinfo"
    headers = {"Authorization": f"Bearer {ACCESS_TOKEN}"}
    
    try:
        response = requests.get(url, headers=headers)
        response.raise_for_status()
        data = response.json()
        return data.get("sub") # 'sub' è l'ID utente in OpenID Connect
    except Exception as e:
        # print(f"Errore recupero URN utente: {e}")
        return None

def get_companies():
    """Recupera la lista delle aziende gestite dall'utente"""
    if not ACCESS_TOKEN:
        return {"error": "Missing Access Token"}

    companies = []
    
    # 1. Aggiungi sempre il profilo personale come opzione di default
    try:
        user_urn = get_user_urn()
        if user_urn:
            companies.append({
                "id": f"urn:li:person:{user_urn}",
                "name": "👤 Profilo Personale",
                "type": "person"
            })
    except Exception as e:
        # Se fallisce il recupero del profilo, è un problema serio, ma proviamo a continuare
        pass

    # 2. Tenta di recuperare le aziende
    url = "https://api.linkedin.com/v2/organizationalEntityAcls?q=roleAssignee&role=ADMINISTRATOR&state=APPROVED&projection=(elements*(organizationalTarget~(localizedName)))"
    headers = {
        "Authorization": f"Bearer {ACCESS_TOKEN}",
        "X-Restli-Protocol-Version": "2.0.0"
    }

    try:
        response = requests.get(url, headers=headers)
        response.raise_for_status()
        data = response.json()
        
        for element in data.get("elements", []):
            org_target = element.get("organizationalTarget~", {})
            org_urn = element.get("organizationalTarget", "")
            name = org_target.get("localizedName", "Unknown Company")
            
            companies.append({
                "id": org_urn,
                "name": f"🏢 {name}",
                "type": "company"
            })
            
    except Exception as e:
        # Se fallisce il recupero delle aziende (es. permessi mancanti), restituisci comunque il profilo personale se presente
        # Potremmo aggiungere un'entry di errore fittizia per avvisare l'utente nell'UI
        if not companies:
            return {"error": f"Impossibile recuperare profilo o aziende: {str(e)}"}
        
        # Opzionale: Aggiungi un avviso come "azienda" speciale
        companies.append({
            "id": "",
            "name": f"⚠️ Impossibile caricare aziende (Permessi insufficienti?)",
            "type": "error",
            "disabled": True
        })
        
    return companies

def register_upload(author_urn, file_path):
    url = "https://api.linkedin.com/v2/assets?action=registerUpload"
    headers = {
        "Authorization": f"Bearer {ACCESS_TOKEN}",
        "Content-Type": "application/json",
    }
    
    # Determine type
    ext = os.path.splitext(file_path)[1].lower()
    if ext in ['.jpg', '.jpeg', '.png', '.gif']:
        recipe = "urn:li:digitalmediaRecipe:feedshare-image"
        media_category = "IMAGE"
    elif ext in ['.mp4', '.mov', '.avi']:
        recipe = "urn:li:digitalmediaRecipe:feedshare-video"
        media_category = "VIDEO"
    else:
        return None, None, None, "Unsupported file type"

    payload = {
        "registerUploadRequest": {
            "recipes": [recipe],
            "owner": author_urn,
            "serviceRelationships": [
                {
                    "relationshipType": "OWNER",
                    "identifier": "urn:li:userGeneratedContent"
                }
            ]
        }
    }
    
    try:
        resp = requests.post(url, headers=headers, json=payload)
        resp.raise_for_status()
        data = resp.json()
        upload_url = data['value']['uploadMechanism']['com.linkedin.digitalmedia.uploading.MediaUploadHttpRequest']['uploadUrl']
        asset = data['value']['asset']
        return upload_url, asset, media_category, None
    except Exception as e:
        return None, None, None, str(e)

def upload_file(upload_url, file_path):
    try:
        with open(file_path, 'rb') as f:
            headers = {"Authorization": f"Bearer {ACCESS_TOKEN}"}
            resp = requests.put(upload_url, data=f, headers=headers)
            if resp.status_code not in [200, 201]:
                return False, resp.text
        return True, None
    except Exception as e:
        return False, str(e)

def post_to_linkedin(text, media_path=None, author_urn=None, json_output=False):
    if not ACCESS_TOKEN:
        msg = "Errore: LINKEDIN_ACCESS_TOKEN (o LINKEDIN_TOKEN) non trovato nel file .env"
        if json_output:
            print(json.dumps({"error": msg}))
        else:
            print(msg)
        return

    if not author_urn:
        user_id = get_user_urn()
        if not user_id:
            msg = "Impossibile recuperare l'ID autore."
            if json_output:
                print(json.dumps({"error": msg}))
            else:
                print(msg)
            return
        author_urn = f"urn:li:person:{user_id}"

    share_media_category = "NONE"
    media_content = []

    if media_path:
        upload_url, asset, category, err = register_upload(author_urn, media_path)
        if err:
            msg = f"Errore registrazione upload: {err}"
            if json_output:
                print(json.dumps({"error": msg}))
            else:
                print(msg)
            return
        
        success, err = upload_file(upload_url, media_path)
        if not success:
            msg = f"Errore caricamento file: {err}"
            if json_output:
                print(json.dumps({"error": msg}))
            else:
                print(msg)
            return
            
        share_media_category = category
        media_content.append({
            "status": "READY",
            "description": {
                "text": text[:200] # Optional description
            },
            "media": asset,
            "title": {
                "text": "Media Content" # Optional title
            }
        })

    url = "https://api.linkedin.com/v2/ugcPosts"
    headers = {
        "Authorization": f"Bearer {ACCESS_TOKEN}",
        "Content-Type": "application/json",
        "X-Restli-Protocol-Version": "2.0.0"
    }

    specific_content = {
        "com.linkedin.ugc.ShareContent": {
            "shareCommentary": {
                "text": text
            },
            "shareMediaCategory": share_media_category
        }
    }
    
    if share_media_category != "NONE":
        specific_content["com.linkedin.ugc.ShareContent"]["media"] = media_content

    payload = {
        "author": author_urn,
        "lifecycleState": "PUBLISHED",
        "specificContent": specific_content,
        "visibility": {
            "com.linkedin.ugc.MemberNetworkVisibility": "PUBLIC"
        }
    }

    try:
        response = requests.post(url, headers=headers, json=payload)
        response.raise_for_status()
        res_json = response.json()
        if json_output:
            print(json.dumps(res_json))
        else:
            print(f"✅ Post pubblicato con successo! ID: {res_json.get('id')}")
    except requests.exceptions.RequestException as e:
        error_msg = f"Errore pubblicazione: {e}"
        details = ""
        if hasattr(e.response, 'text'):
            details = e.response.text
            error_msg += f" - {details}"
        
        if json_output:
            print(json.dumps({"error": str(e), "details": details}))
        else:
            print(f"❌ {error_msg}")

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='LinkedIn Post Helper')
    parser.add_argument('text', nargs='?', help='Text content of the post')
    parser.add_argument('--media', help='Path to image or video file')
    parser.add_argument('--author', help='Author URN (urn:li:person:ID or urn:li:organization:ID)')
    parser.add_argument('--list-companies', action='store_true', help='List managed companies')
    parser.add_argument('--json', action='store_true', help='Output in JSON format')
    args = parser.parse_args()

    if args.list_companies:
        companies = get_companies()
        if args.json:
            print(json.dumps(companies))
        else:
            if isinstance(companies, list):
                print("Available targets:")
                for c in companies:
                    print(f"- {c['name']} ({c['id']})")
            else:
                print(companies)
    elif args.text:
        post_to_linkedin(args.text, media_path=args.media, author_urn=args.author, json_output=args.json)
    else:
        print("Usage: python post_to_linkedin.py <text> [--media <path>] [--author <urn>] [--list-companies] [--json]")
