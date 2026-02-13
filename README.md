## Prerequisite

1. Go to https://portal.azure.com/ and search "Microsoft Entra ID"
2. Click "Add" -> "App registration"
3. Set a Redirect URI as "http://localhost:9091/callback"
4. Copy "Application (client) ID" and "Directory (tenant) ID"
5. Create "Client secrets" and copy it
6. Create `.env` with these values

```
# Sample value for Mac
BROWSER_PATH="/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
OUTPUT_DIR=[Your directory to save generated HTML file]

CLIENT_ID=[Application (client) ID]
TENANT_ID=[Directory (tenant) ID]
CLIENT_SECRET=[Client secrets]
```

## Start App

```
go run main.go
```
