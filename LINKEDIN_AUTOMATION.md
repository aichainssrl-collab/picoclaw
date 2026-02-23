# Guida all'Automazione LinkedIn per PicoClaw

Questa guida spiega come configurare un'automazione per connettersi a LinkedIn e pubblicare post automaticamente utilizzando PicoClaw.

## 1. Prerequisiti: Configurazione LinkedIn Developer

Prima di iniziare, è necessario registrare un'applicazione sulla piattaforma sviluppatori di LinkedIn.

1.  Accedi a [LinkedIn Developers](https://www.linkedin.com/developers/).
2.  Clicca su **"Create app"**.
    *   Compila i dettagli (Nome, Pagina LinkedIn associata, Logo).
3.  Nella scheda **Products**, richiedi l'accesso a:
    *   **Share on LinkedIn** (per postare).
    *   **Sign In with LinkedIn** (per l'autenticazione).
4.  Nella scheda **Auth**:
    *   Copia il **Client ID** e il **Client Secret**.
    *   Aggiungi un **Authorized redirect URL** per il flusso OAuth (es. `http://localhost:4001/linkedin/callback` se stai testando localmente, o un URL valido).

## 2. Ottenere l'Access Token (OAuth 2.0)

Per pubblicare a nome di un utente, serve un **Access Token**.

### Passaggio A: Ottenere il Codice di Autorizzazione
Genera questo URL nel browser (sostituisci i valori):
```
https://www.linkedin.com/oauth/v2/authorization?response_type=code&client_id=77ouk2bw4z1mrt&redirect_uri=http://localhost:4001/linkedin/callback&state=random_string&scope=w_member_social%20openid%20profile%20email
```
Autorizza l'app e copia il parametro `code` dall'URL di reindirizzamento.

### Passaggio B: Scambiare il Codice per il Token
Esegui una richiesta POST (puoi usare `curl` o Postman):
```bash
curl -X POST https://www.linkedin.com/oauth/v2/accessToken \
 -d grant_type=authorization_code \
 -d code={YOUR_AUTHORIZATION_CODE} \
 -d redirect_uri={YOUR_REDIRECT_URI} \
 -d client_id={YOUR_CLIENT_ID} \
 -d client_secret={YOUR_CLIENT_SECRET}
```
La risposta conterrà l'`access_token`.

## 3. Utilizzo dell'Interfaccia Web (Consigliato)

PicoClaw ora include una pagina web dedicata per gestire l'autenticazione e la pubblicazione su LinkedIn.

1.  Assicurati che il gateway sia in esecuzione:
    ```bash
    go run cmd/picoclaw/main.go gateway --config config/config.json
    ```
2.  Apri il browser all'indirizzo: [http://localhost:4001/linkedin](http://localhost:4001/linkedin)
3.  Clicca su **"Recupera Code LinkedIn"**.
4.  Autorizza l'applicazione su LinkedIn.
5.  Verrai reindirizzato automaticamente alla pagina di gestione, e il codice verrà inserito nel campo apposito.
6.  Clicca su **"Recupera Token"** per ottenere e salvare automaticamente il token nel file `.env`.
7.  Usa il modulo **"Crea e Pubblica Post"** per inviare aggiornamenti direttamente dal browser.

## 4. Script di Automazione (Manuale)

Se preferisci usare la riga di comando:

### A. Recupero Token (`get_linkedin_token.py`)
Questo script ti guida passo passo nell'ottenere il tuo Access Token.

1. Esegui lo script:
   ```bash
   python3 get_linkedin_token.py
   ```
2. Segui le istruzioni a schermo:
   - Visita l'URL generato.
   - Autorizza l'applicazione.
   - Copia il parametro `code` dall'URL di reindirizzamento.
   - Incolla il codice nello script.
3. Lo script ti restituirà il tuo **Access Token**.
4. Copia il token nel file `.env` alla voce `LINKEDIN_ACCESS_TOKEN`.

### B. Pubblicazione Post (`post_to_linkedin.py`)
Una volta configurato il token, puoi pubblicare post direttamente da terminale.

**Utilizzo:**
```bash
python3 post_to_linkedin.py "Il testo del tuo post qui"
```

**Esempio:**
```bash
python3 post_to_linkedin.py "Ciao LinkedIn! Sto testando la mia automazione con Python 🚀"
```

## 5. Configurazione Environment (.env)
Assicurati che il tuo file `.env` contenga:

```env
LINKEDIN_CLIENT_ID=tuo_client_id
LINKEDIN_CLIENT_SECRET=tuo_client_secret
LINKEDIN_REDIRECT_URI=http://localhost:4001/linkedin
LINKEDIN_ACCESS_TOKEN=tuo_access_token_recuperato
```

## 6. Automazione Avanzata (Cron Job)

Puoi pianificare post automatici aggiungendo un job al file `workspace/cron/jobs.json` di PicoClaw (se supportato) o usando `crontab` di sistema:

```json
{
  "name": "Daily LinkedIn Update",
  "schedule": "0 9 * * *",
  "command": "python3 /path/to/workspace/scripts/linkedin_post.py 'Buongiorno rete! Ecco il mio aggiornamento quotidiano.'"
}
```
