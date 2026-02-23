# Confronto: PicoClaw vs OpenClaw

Questo documento illustra le principali differenze tra **PicoClaw** e **OpenClaw**, evidenziando i vantaggi in termini di prestazioni, efficienza e costi che rendono PicoClaw una soluzione ideale per l'edge computing e i dispositivi a basso consumo.

## 📊 Tabella Riassuntiva

| Caratteristica | OpenClaw | PicoClaw |
| :--- | :--- | :--- |
| **Linguaggio** | TypeScript (Node.js) | **Go (Golang)** |
| **Utilizzo RAM** | > 1 GB | **< 10 MB** (99% in meno) |
| **Tempo di Avvio** (0.8GHz) | > 500 secondi | **< 1 secondo** (400x più veloce) |
| **Hardware Richiesto** | PC/Mac (es. Mac Mini) | **Qualsiasi Linux SBC** (es. LicheeRV) |
| **Costo Hardware** | ~$600+ | **A partire da $10** |
| **Architettura** | Monolitica / Interprete Script | **Binario Singolo / Nativo** |

## 🚀 Dettagli delle Differenze

### 1. Linguaggio e Architettura
*   **OpenClaw**: Basato su TypeScript e l'ecosistema Node.js. Sebbene flessibile, porta con sé l'overhead del runtime JavaScript e delle dipendenze npm.
*   **PicoClaw**: Riscritto completamente in **Go**. Questo permette di compilare un singolo binario statico, privo di dipendenze esterne, ottimizzato per l'esecuzione nativa su diverse architetture (x86, ARM, RISC-V).

### 2. Prestazioni e Risorse
*   **Memoria**: PicoClaw è progettato per essere **estremamente leggero**. Mentre OpenClaw può saturare facilmente 1GB di RAM, PicoClaw opera confortevolmente in meno di 10MB, rendendolo perfetto per ambienti con risorse limitate.
*   **Velocità**: Grazie alla natura compilata di Go e all'architettura ottimizzata, PicoClaw si avvia istantaneamente (<1s), contro i minuti necessari a OpenClaw su hardware di fascia bassa.

### 3. Hardware e Costi
*   **OpenClaw**: Generalmente richiede hardware desktop o server potenti per girare fluidamente.
*   **PicoClaw**: Abilita l'intelligenza artificiale su dispositivi **IoT ed Edge** ultra-economici (fino a $10). È compatibile con board come LicheeRV Nano, Milk-V Duo, e Raspberry Pi Zero, democratizzando l'accesso agli assistenti AI.

## 🔄 Migrazione da OpenClaw

PicoClaw include uno strumento integrato per facilitare la transizione da OpenClaw.

È possibile migrare configurazioni e workspace esistenti utilizzando il comando:

```bash
picoclaw migrate
```

Opzioni disponibili:
*   `--dry-run`: Simula la migrazione senza applicare modifiche.
*   `--refresh`: Risincronizza i file del workspace.
*   `--config-only`: Migra solo la configurazione.

## 💡 Conclusione

PicoClaw non è solo un'alternativa a OpenClaw, ma un'evoluzione verso l'efficienza estrema. Mantiene le capacità di un agente AI avanzato rimuovendo le barriere hardware, permettendo deploy "ovunque", dal cloud ai microcontrollori Linux embedded.
