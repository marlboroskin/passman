# PassMan — Secure Password Manager

Your secrets, protected. No cloud. No tracking. Just Go.

A minimal, secure, and self-hosted password manager built with Go.  
Encrypt your credentials locally with AES-256-GCM and PBKDF2.  
No external servers. No dependencies. Just a single binary.

---

## Features

- **End-to-End Encryption**  
  Your data is encrypted with AES-256-GCM using your master password. Only you can decrypt it.

- **Strong Key Derivation**  
  PBKDF2-HMAC-SHA256 with 100,000 iterations — resistant to brute-force attacks.

- **Smart Search**  
  Find accounts by name, login, or URL — case-insensitive and fast.

- **Copy to Clipboard**  
  Instantly copy any password with one command. Auto-clears after 10 seconds (coming soon).

- **Local-Only Storage**  
  Everything is saved to `data.enc` — no cloud, no servers, no logs.

- **Backup & Restore**  
  Create encrypted backups. Restore from any backup file with your password.

- **Password Generation**  
  Generate strong passwords (8–128 characters) with letters, digits, and symbols.

- **Clean CLI Interface**  
  Simple menu-driven interface — no learning curve.

---

## Quick Start

1. **Download or build the binary**
   ```bash
   go build -o passman.exe main.go

Run it

./passman.exe   

Enter a strong master password
(e.g. MyPass123!) — you’ll need it every time.

Use the menu:

__Password Manager__
1. Create account
2. Find account
3. Delete account
4. Exit
5. Generate password
6. Copy password to clipboard
7. Create encrypted backup
8. Restore from backup


🔐 Security

Encryption: AES-256-GCM with random salt and nonce
Key Derivation: PBKDF2-HMAC-SHA256, 100,000 iterations
Storage: All data encrypted in data.enc
No Internet: Zero network calls, no analytics, no tracking

⚠️ Warning: If you lose your master password — recovery is impossible. Keep it safe.

💾 Backup & Recovery
Create Backup (Menu 7):
Saves an encrypted backup to backup/vault_2025-04-05_12-30-45.enc

Restore from Backup (Menu 8):
Overwrites current vault. Requires the same master password

Backups are encrypted — safe to store on USB, cloud, or email.

🧱 Project Structure
passman/ ├── main.go # CLI menu & flow ├── account/ │ ├── account.go # Account model │ ├── vault.go # In-memory storage & search │ └── crypto/ │ └── encrypt.go # AES + PBKDF2 encryption ├── files/ │ └── files.go # Safe file I/O ├── data.enc # Your encrypted vault (never share!) ├── backup/ # Encrypted backup files ├── go.mod └── README.md

📦 Dependencies
github.com/fatih/color — Colored terminal output
github.com/atotto/clipboard — Copy to clipboard

Install with:
go get github.com/fatih/color
go get github.com/atotto/clipboard

🛡️ Best Practices
Use a strong master password (12+ chars, mix of upper, lower, digits, symbols)
Store data.enc in a safe place (encrypted drive, USB, etc.)
Back up regularly
Never share your master password

🙌 Made with Go
This project proves that security, simplicity, and usability can coexist —
in under 500 lines of clean Go code.

No frameworks. No bloat. Just trust.

🚀 Use it. Secure it. Own it.# passman
