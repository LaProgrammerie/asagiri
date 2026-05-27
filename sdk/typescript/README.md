# @laprogrammerie/asagiri

Client TypeScript pour l’API REST locale du runtime Asagiri (`asa runtime serve`).

## Installation

```bash
npm install @laprogrammerie/asagiri
# ou depuis le dépôt
cd sdk/typescript && npm install && npm run build
```

## Usage

Démarrer l’API locale :

```bash
asa runtime serve --port 8765
```

Client :

```ts
import { AsagiriClient } from "@laprogrammerie/asagiri";

const runtime = new AsagiriClient({
  baseUrl: "http://127.0.0.1:8765",
  token: process.env.ASA_RUNTIME_TOKEN, // optionnel si .asagiri/runtime/api.token
});

const status = await runtime.status();
const session = await runtime.startSession("workspace-redesign", "workspace-saas", "onboarding");
await runtime.runFlow(session.id, "onboarding");
```

## Sécurité

- L’API écoute uniquement sur `127.0.0.1`.
- Token optionnel : fichier `.asagiri/runtime/api.token` ou header `Authorization: Bearer …`.

## Tests

```bash
npm test
```
