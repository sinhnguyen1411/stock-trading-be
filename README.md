# stock-trading-be
Contain backend source code

# Install dependencies
## Install buf
Install buf with scoop

Open PowerShell and run the following command:

```powershell
Invoke-RestMethod -Uri https://get.scoop.sh -OutFile scoop-install.ps1
.\scoop-install.ps1
```

Install buf
```bash
scoop install buf
```

## Server Ports
The gRPC server listens on the configured port. If that port is already in use,
it now falls back to an available system-assigned port and logs the chosen
address.

