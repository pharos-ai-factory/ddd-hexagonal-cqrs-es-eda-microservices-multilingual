// Code generated from the API OpenAPI specification. DO NOT EDIT.
// Source: contracts/services/api/http_api/api.openapi.json
export type Schemas = {
  "collection.http_api.schemas.CollectCommand": {"code": string};
  "collection.http_api.schemas.Pickup": {"id": string; "orderId": string; "customerId": string; "code": string; "status": "ready" | "collected"};
  "communication.http_api.schemas.Notification": {"id": string; "recipient": string; "subject": string; "body": string; "status": "requested" | "sent"; "providerReceipt"?: string};
  "loyalty.http_api.schemas.Account": {"id": string; "stampBalance": number; "collections": number; "grantsEarned": number; "lastGrant"?: Schemas["loyalty.http_api.schemas.Grant"]};
  "loyalty.http_api.schemas.Grant": {"id": string; "accountId": string; "benefit": string; "validDays": number};
  "loyalty.http_api.schemas.RedeemCommand": {"orderId": string};
  "loyalty.http_api.schemas.Reward": {"id": string; "grantId": string; "customerId": string; "benefit": string; "status": "issued" | "redeemed" | "expired"; "expiresAt": string; "redeemedFor"?: string};
  "menu.http_api.schemas.Drink": {"id": string; "name": string; "revision": number; "published": boolean};
  "menu.http_api.schemas.DrinkCommand": {"name": string};
  "menu.http_api.schemas.Edition": {"id": string; "currency": string; "status": "draft" | "published"; "offers": (Array<Schemas["menu.http_api.schemas.Offer"]>) | null};
  "menu.http_api.schemas.EditionCommand": {"currency": string};
  "menu.http_api.schemas.Offer": {"code": string; "drinkId": string; "drinkRevision": number; "name": string; "minor": number; "currency": string};
  "menu.http_api.schemas.OfferCommand": {"code": string; "drinkId": string; "drinkRevision": number; "minor": number};
  "menu.http_api.schemas.PriceCommand": {"code": string; "minor": number};
  "ordering.http_api.schemas.CreateCommand": {"customerId": string; "editionId": string};
  "ordering.http_api.schemas.Line": {"id": string; "selection": Schemas["ordering.http_api.schemas.Selection"]; "quantity": number};
  "ordering.http_api.schemas.LineCommand": {"lineId": string; "editionId": string; "offerCode": string; "quantity": number};
  "ordering.http_api.schemas.Order": {"id": string; "customerId": string; "editionId": string; "currency": string; "status": "draft" | "placed"; "lines": Array<Schemas["ordering.http_api.schemas.Line"]>};
  "ordering.http_api.schemas.QuantityCommand": {"lineId": string; "quantity": number};
  "ordering.http_api.schemas.Selection": {"offerCode": string; "name": string; "minor": number};
  "preparation.http_api.schemas.Ticket": {"id": string; "orderId": string; "customerId": string; "instructions": string; "status": "queued" | "preparing" | "ready"};
  "services.api.http_api.operational.schemas.SessionDiagnostics": {"pendingRevocations": number; "expiredSessionsAwaitingRevocation": number; "processFailures": number; "lastFailure"?: {"at": string; "operation": string; "class": string}};
  "shared.http_api.schemas.EmptyCommand": Record<string, never>;
  "shared.http_api.schemas.Error": {"code": string; "message"?: string};
  "shared.http_api.schemas.Outcome": {"aggregateId": string; "version": number; "status": string};
  "shared.http_api.schemas.RejectedOutcome": {"aggregateId": string; "version": number; "status": string; "rejection": {"code": string; "message": string}};
};
export type Operations = {
  "realtimeConnect": {input: {path?: never; query?: never; headers?: never; body?: {[key: string]: unknown}}; responses: {200: ({"result": {"user": string; "subs": {[key: string]: Record<string, never>}; "expire_at": number}}) | ({"disconnect": {"code": number; "reason": string}}); 403: Schemas["shared.http_api.schemas.Error"]}};
  "realtimeRefresh": {input: {path?: never; query?: never; headers?: never; body?: {[key: string]: unknown}}; responses: {200: ({"result": {"expire_at": number}}) | ({"result": {"expired": true}}); 403: Schemas["shared.http_api.schemas.Error"]}};
  "listPickups": {input: {path?: never; query?: {"limit"?: number; "cursor"?: string}; headers?: never; body?: never}; responses: {200: (Array<{"exists": true; "version": number; "state": Schemas["collection.http_api.schemas.Pickup"]}>) | ({"items": Array<{"exists": true; "version": number; "state": Schemas["collection.http_api.schemas.Pickup"]}>; "nextCursor": (string) | null}); 400: Schemas["shared.http_api.schemas.Error"]; 401: Schemas["shared.http_api.schemas.Error"]; 403: Schemas["shared.http_api.schemas.Error"]; 503: Schemas["shared.http_api.schemas.Error"]}};
  "getPickup": {input: {path: {"id": string}; query?: never; headers?: never; body?: never}; responses: {200: {"exists": true; "version": number; "state": Schemas["collection.http_api.schemas.Pickup"]}; 400: Schemas["shared.http_api.schemas.Error"]; 401: Schemas["shared.http_api.schemas.Error"]; 403: Schemas["shared.http_api.schemas.Error"]; 404: Schemas["shared.http_api.schemas.Error"]; 503: Schemas["shared.http_api.schemas.Error"]}};
  "collectOrder": {input: {path: {"id": string}; query?: never; headers: {"Idempotency-Key": string; "If-Match": string; "X-Correlation-ID"?: string}; body: Schemas["collection.http_api.schemas.CollectCommand"]}; responses: {200: Schemas["shared.http_api.schemas.Outcome"]; 400: Schemas["shared.http_api.schemas.Error"]; 401: Schemas["shared.http_api.schemas.Error"]; 403: Schemas["shared.http_api.schemas.Error"]; 404: Schemas["shared.http_api.schemas.RejectedOutcome"]; 409: Schemas["shared.http_api.schemas.RejectedOutcome"]; 422: Schemas["shared.http_api.schemas.RejectedOutcome"]; 428: Schemas["shared.http_api.schemas.Error"]; 503: Schemas["shared.http_api.schemas.Error"]}};
  "listNotifications": {input: {path?: never; query?: {"limit"?: number; "cursor"?: string}; headers?: never; body?: never}; responses: {200: (Array<{"exists": true; "version": number; "state": Schemas["communication.http_api.schemas.Notification"]}>) | ({"items": Array<{"exists": true; "version": number; "state": Schemas["communication.http_api.schemas.Notification"]}>; "nextCursor": (string) | null}); 400: Schemas["shared.http_api.schemas.Error"]; 401: Schemas["shared.http_api.schemas.Error"]; 403: Schemas["shared.http_api.schemas.Error"]; 503: Schemas["shared.http_api.schemas.Error"]}};
  "getNotification": {input: {path: {"id": string}; query?: never; headers?: never; body?: never}; responses: {200: {"exists": true; "version": number; "state": Schemas["communication.http_api.schemas.Notification"]}; 400: Schemas["shared.http_api.schemas.Error"]; 401: Schemas["shared.http_api.schemas.Error"]; 403: Schemas["shared.http_api.schemas.Error"]; 404: Schemas["shared.http_api.schemas.Error"]; 503: Schemas["shared.http_api.schemas.Error"]}};
  "listAccounts": {input: {path?: never; query?: {"limit"?: number; "cursor"?: string}; headers?: never; body?: never}; responses: {200: (Array<{"exists": true; "version": number; "state": Schemas["loyalty.http_api.schemas.Account"]}>) | ({"items": Array<{"exists": true; "version": number; "state": Schemas["loyalty.http_api.schemas.Account"]}>; "nextCursor": (string) | null}); 400: Schemas["shared.http_api.schemas.Error"]; 401: Schemas["shared.http_api.schemas.Error"]; 403: Schemas["shared.http_api.schemas.Error"]; 503: Schemas["shared.http_api.schemas.Error"]}};
  "getAccount": {input: {path: {"id": string}; query?: never; headers?: never; body?: never}; responses: {200: {"exists": true; "version": number; "state": Schemas["loyalty.http_api.schemas.Account"]}; 400: Schemas["shared.http_api.schemas.Error"]; 401: Schemas["shared.http_api.schemas.Error"]; 403: Schemas["shared.http_api.schemas.Error"]; 404: Schemas["shared.http_api.schemas.Error"]; 503: Schemas["shared.http_api.schemas.Error"]}};
  "listRewards": {input: {path?: never; query?: {"limit"?: number; "cursor"?: string}; headers?: never; body?: never}; responses: {200: (Array<{"exists": true; "version": number; "state": Schemas["loyalty.http_api.schemas.Reward"]}>) | ({"items": Array<{"exists": true; "version": number; "state": Schemas["loyalty.http_api.schemas.Reward"]}>; "nextCursor": (string) | null}); 400: Schemas["shared.http_api.schemas.Error"]; 401: Schemas["shared.http_api.schemas.Error"]; 403: Schemas["shared.http_api.schemas.Error"]; 503: Schemas["shared.http_api.schemas.Error"]}};
  "getReward": {input: {path: {"id": string}; query?: never; headers?: never; body?: never}; responses: {200: {"exists": true; "version": number; "state": Schemas["loyalty.http_api.schemas.Reward"]}; 400: Schemas["shared.http_api.schemas.Error"]; 401: Schemas["shared.http_api.schemas.Error"]; 403: Schemas["shared.http_api.schemas.Error"]; 404: Schemas["shared.http_api.schemas.Error"]; 503: Schemas["shared.http_api.schemas.Error"]}};
  "redeemReward": {input: {path: {"id": string}; query?: never; headers: {"Idempotency-Key": string; "If-Match": string; "X-Correlation-ID"?: string}; body: Schemas["loyalty.http_api.schemas.RedeemCommand"]}; responses: {200: Schemas["shared.http_api.schemas.Outcome"]; 400: Schemas["shared.http_api.schemas.Error"]; 401: Schemas["shared.http_api.schemas.Error"]; 403: Schemas["shared.http_api.schemas.Error"]; 404: Schemas["shared.http_api.schemas.RejectedOutcome"]; 409: Schemas["shared.http_api.schemas.RejectedOutcome"]; 422: Schemas["shared.http_api.schemas.RejectedOutcome"]; 428: Schemas["shared.http_api.schemas.Error"]; 503: Schemas["shared.http_api.schemas.Error"]}};
  "listDrinks": {input: {path?: never; query?: {"limit"?: number; "cursor"?: string}; headers?: never; body?: never}; responses: {200: (Array<{"exists": true; "version": number; "state": Schemas["menu.http_api.schemas.Drink"]}>) | ({"items": Array<{"exists": true; "version": number; "state": Schemas["menu.http_api.schemas.Drink"]}>; "nextCursor": (string) | null}); 400: Schemas["shared.http_api.schemas.Error"]; 401: Schemas["shared.http_api.schemas.Error"]; 403: Schemas["shared.http_api.schemas.Error"]; 503: Schemas["shared.http_api.schemas.Error"]}};
  "getDrink": {input: {path: {"id": string}; query?: never; headers?: never; body?: never}; responses: {200: {"exists": true; "version": number; "state": Schemas["menu.http_api.schemas.Drink"]}; 400: Schemas["shared.http_api.schemas.Error"]; 401: Schemas["shared.http_api.schemas.Error"]; 403: Schemas["shared.http_api.schemas.Error"]; 404: Schemas["shared.http_api.schemas.Error"]; 503: Schemas["shared.http_api.schemas.Error"]}};
  "createDrink": {input: {path: {"id": string}; query?: never; headers: {"Idempotency-Key": string; "If-Match": string; "X-Correlation-ID"?: string}; body: Schemas["menu.http_api.schemas.DrinkCommand"]}; responses: {200: Schemas["shared.http_api.schemas.Outcome"]; 400: Schemas["shared.http_api.schemas.Error"]; 401: Schemas["shared.http_api.schemas.Error"]; 403: Schemas["shared.http_api.schemas.Error"]; 404: Schemas["shared.http_api.schemas.RejectedOutcome"]; 409: Schemas["shared.http_api.schemas.RejectedOutcome"]; 422: Schemas["shared.http_api.schemas.RejectedOutcome"]; 428: Schemas["shared.http_api.schemas.Error"]; 503: Schemas["shared.http_api.schemas.Error"]}};
  "publishDrink": {input: {path: {"id": string}; query?: never; headers: {"Idempotency-Key": string; "If-Match": string; "X-Correlation-ID"?: string}; body: Schemas["shared.http_api.schemas.EmptyCommand"]}; responses: {200: Schemas["shared.http_api.schemas.Outcome"]; 400: Schemas["shared.http_api.schemas.Error"]; 401: Schemas["shared.http_api.schemas.Error"]; 403: Schemas["shared.http_api.schemas.Error"]; 404: Schemas["shared.http_api.schemas.RejectedOutcome"]; 409: Schemas["shared.http_api.schemas.RejectedOutcome"]; 422: Schemas["shared.http_api.schemas.RejectedOutcome"]; 428: Schemas["shared.http_api.schemas.Error"]; 503: Schemas["shared.http_api.schemas.Error"]}};
  "reviseDrink": {input: {path: {"id": string}; query?: never; headers: {"Idempotency-Key": string; "If-Match": string; "X-Correlation-ID"?: string}; body: Schemas["menu.http_api.schemas.DrinkCommand"]}; responses: {200: Schemas["shared.http_api.schemas.Outcome"]; 400: Schemas["shared.http_api.schemas.Error"]; 401: Schemas["shared.http_api.schemas.Error"]; 403: Schemas["shared.http_api.schemas.Error"]; 404: Schemas["shared.http_api.schemas.RejectedOutcome"]; 409: Schemas["shared.http_api.schemas.RejectedOutcome"]; 422: Schemas["shared.http_api.schemas.RejectedOutcome"]; 428: Schemas["shared.http_api.schemas.Error"]; 503: Schemas["shared.http_api.schemas.Error"]}};
  "listEditions": {input: {path?: never; query?: {"limit"?: number; "cursor"?: string}; headers?: never; body?: never}; responses: {200: (Array<{"exists": true; "version": number; "state": Schemas["menu.http_api.schemas.Edition"]}>) | ({"items": Array<{"exists": true; "version": number; "state": Schemas["menu.http_api.schemas.Edition"]}>; "nextCursor": (string) | null}); 400: Schemas["shared.http_api.schemas.Error"]; 401: Schemas["shared.http_api.schemas.Error"]; 403: Schemas["shared.http_api.schemas.Error"]; 503: Schemas["shared.http_api.schemas.Error"]}};
  "getEdition": {input: {path: {"id": string}; query?: never; headers?: never; body?: never}; responses: {200: {"exists": true; "version": number; "state": Schemas["menu.http_api.schemas.Edition"]}; 400: Schemas["shared.http_api.schemas.Error"]; 401: Schemas["shared.http_api.schemas.Error"]; 403: Schemas["shared.http_api.schemas.Error"]; 404: Schemas["shared.http_api.schemas.Error"]; 503: Schemas["shared.http_api.schemas.Error"]}};
  "createEdition": {input: {path: {"id": string}; query?: never; headers: {"Idempotency-Key": string; "If-Match": string; "X-Correlation-ID"?: string}; body: Schemas["menu.http_api.schemas.EditionCommand"]}; responses: {200: Schemas["shared.http_api.schemas.Outcome"]; 400: Schemas["shared.http_api.schemas.Error"]; 401: Schemas["shared.http_api.schemas.Error"]; 403: Schemas["shared.http_api.schemas.Error"]; 404: Schemas["shared.http_api.schemas.RejectedOutcome"]; 409: Schemas["shared.http_api.schemas.RejectedOutcome"]; 422: Schemas["shared.http_api.schemas.RejectedOutcome"]; 428: Schemas["shared.http_api.schemas.Error"]; 503: Schemas["shared.http_api.schemas.Error"]}};
  "addOffer": {input: {path: {"id": string}; query?: never; headers: {"Idempotency-Key": string; "If-Match": string; "X-Correlation-ID"?: string}; body: Schemas["menu.http_api.schemas.OfferCommand"]}; responses: {200: Schemas["shared.http_api.schemas.Outcome"]; 400: Schemas["shared.http_api.schemas.Error"]; 401: Schemas["shared.http_api.schemas.Error"]; 403: Schemas["shared.http_api.schemas.Error"]; 404: Schemas["shared.http_api.schemas.RejectedOutcome"]; 409: Schemas["shared.http_api.schemas.RejectedOutcome"]; 422: Schemas["shared.http_api.schemas.RejectedOutcome"]; 428: Schemas["shared.http_api.schemas.Error"]; 503: Schemas["shared.http_api.schemas.Error"]}};
  "changePrice": {input: {path: {"id": string}; query?: never; headers: {"Idempotency-Key": string; "If-Match": string; "X-Correlation-ID"?: string}; body: Schemas["menu.http_api.schemas.PriceCommand"]}; responses: {200: Schemas["shared.http_api.schemas.Outcome"]; 400: Schemas["shared.http_api.schemas.Error"]; 401: Schemas["shared.http_api.schemas.Error"]; 403: Schemas["shared.http_api.schemas.Error"]; 404: Schemas["shared.http_api.schemas.RejectedOutcome"]; 409: Schemas["shared.http_api.schemas.RejectedOutcome"]; 422: Schemas["shared.http_api.schemas.RejectedOutcome"]; 428: Schemas["shared.http_api.schemas.Error"]; 503: Schemas["shared.http_api.schemas.Error"]}};
  "publishEdition": {input: {path: {"id": string}; query?: never; headers: {"Idempotency-Key": string; "If-Match": string; "X-Correlation-ID"?: string}; body: Schemas["shared.http_api.schemas.EmptyCommand"]}; responses: {200: Schemas["shared.http_api.schemas.Outcome"]; 400: Schemas["shared.http_api.schemas.Error"]; 401: Schemas["shared.http_api.schemas.Error"]; 403: Schemas["shared.http_api.schemas.Error"]; 404: Schemas["shared.http_api.schemas.RejectedOutcome"]; 409: Schemas["shared.http_api.schemas.RejectedOutcome"]; 422: Schemas["shared.http_api.schemas.RejectedOutcome"]; 428: Schemas["shared.http_api.schemas.Error"]; 503: Schemas["shared.http_api.schemas.Error"]}};
  "listOrders": {input: {path?: never; query?: {"limit"?: number; "cursor"?: string}; headers?: never; body?: never}; responses: {200: (Array<{"exists": true; "version": number; "state": Schemas["ordering.http_api.schemas.Order"]}>) | ({"items": Array<{"exists": true; "version": number; "state": Schemas["ordering.http_api.schemas.Order"]}>; "nextCursor": (string) | null}); 400: Schemas["shared.http_api.schemas.Error"]; 401: Schemas["shared.http_api.schemas.Error"]; 403: Schemas["shared.http_api.schemas.Error"]; 503: Schemas["shared.http_api.schemas.Error"]}};
  "getOrder": {input: {path: {"id": string}; query?: never; headers?: never; body?: never}; responses: {200: {"exists": true; "version": number; "state": Schemas["ordering.http_api.schemas.Order"]}; 400: Schemas["shared.http_api.schemas.Error"]; 401: Schemas["shared.http_api.schemas.Error"]; 403: Schemas["shared.http_api.schemas.Error"]; 404: Schemas["shared.http_api.schemas.Error"]; 503: Schemas["shared.http_api.schemas.Error"]}};
  "createOrder": {input: {path: {"id": string}; query?: never; headers: {"Idempotency-Key": string; "If-Match": string; "X-Correlation-ID"?: string}; body: Schemas["ordering.http_api.schemas.CreateCommand"]}; responses: {200: Schemas["shared.http_api.schemas.Outcome"]; 400: Schemas["shared.http_api.schemas.Error"]; 401: Schemas["shared.http_api.schemas.Error"]; 403: Schemas["shared.http_api.schemas.Error"]; 404: Schemas["shared.http_api.schemas.RejectedOutcome"]; 409: Schemas["shared.http_api.schemas.RejectedOutcome"]; 422: Schemas["shared.http_api.schemas.RejectedOutcome"]; 428: Schemas["shared.http_api.schemas.Error"]; 503: Schemas["shared.http_api.schemas.Error"]}};
  "addLine": {input: {path: {"id": string}; query?: never; headers: {"Idempotency-Key": string; "If-Match": string; "X-Correlation-ID"?: string}; body: Schemas["ordering.http_api.schemas.LineCommand"]}; responses: {200: Schemas["shared.http_api.schemas.Outcome"]; 400: Schemas["shared.http_api.schemas.Error"]; 401: Schemas["shared.http_api.schemas.Error"]; 403: Schemas["shared.http_api.schemas.Error"]; 404: Schemas["shared.http_api.schemas.RejectedOutcome"]; 409: Schemas["shared.http_api.schemas.RejectedOutcome"]; 422: Schemas["shared.http_api.schemas.RejectedOutcome"]; 428: Schemas["shared.http_api.schemas.Error"]; 503: Schemas["shared.http_api.schemas.Error"]}};
  "placeOrder": {input: {path: {"id": string}; query?: never; headers: {"Idempotency-Key": string; "If-Match": string; "X-Correlation-ID"?: string}; body: Schemas["shared.http_api.schemas.EmptyCommand"]}; responses: {200: Schemas["shared.http_api.schemas.Outcome"]; 400: Schemas["shared.http_api.schemas.Error"]; 401: Schemas["shared.http_api.schemas.Error"]; 403: Schemas["shared.http_api.schemas.Error"]; 404: Schemas["shared.http_api.schemas.RejectedOutcome"]; 409: Schemas["shared.http_api.schemas.RejectedOutcome"]; 422: Schemas["shared.http_api.schemas.RejectedOutcome"]; 428: Schemas["shared.http_api.schemas.Error"]; 503: Schemas["shared.http_api.schemas.Error"]}};
  "changeQuantity": {input: {path: {"id": string}; query?: never; headers: {"Idempotency-Key": string; "If-Match": string; "X-Correlation-ID"?: string}; body: Schemas["ordering.http_api.schemas.QuantityCommand"]}; responses: {200: Schemas["shared.http_api.schemas.Outcome"]; 400: Schemas["shared.http_api.schemas.Error"]; 401: Schemas["shared.http_api.schemas.Error"]; 403: Schemas["shared.http_api.schemas.Error"]; 404: Schemas["shared.http_api.schemas.RejectedOutcome"]; 409: Schemas["shared.http_api.schemas.RejectedOutcome"]; 422: Schemas["shared.http_api.schemas.RejectedOutcome"]; 428: Schemas["shared.http_api.schemas.Error"]; 503: Schemas["shared.http_api.schemas.Error"]}};
  "listTickets": {input: {path?: never; query?: {"limit"?: number; "cursor"?: string}; headers?: never; body?: never}; responses: {200: (Array<{"exists": true; "version": number; "state": Schemas["preparation.http_api.schemas.Ticket"]}>) | ({"items": Array<{"exists": true; "version": number; "state": Schemas["preparation.http_api.schemas.Ticket"]}>; "nextCursor": (string) | null}); 400: Schemas["shared.http_api.schemas.Error"]; 401: Schemas["shared.http_api.schemas.Error"]; 403: Schemas["shared.http_api.schemas.Error"]; 503: Schemas["shared.http_api.schemas.Error"]}};
  "getTicket": {input: {path: {"id": string}; query?: never; headers?: never; body?: never}; responses: {200: {"exists": true; "version": number; "state": Schemas["preparation.http_api.schemas.Ticket"]}; 400: Schemas["shared.http_api.schemas.Error"]; 401: Schemas["shared.http_api.schemas.Error"]; 403: Schemas["shared.http_api.schemas.Error"]; 404: Schemas["shared.http_api.schemas.Error"]; 503: Schemas["shared.http_api.schemas.Error"]}};
  "completePreparation": {input: {path: {"id": string}; query?: never; headers: {"Idempotency-Key": string; "If-Match": string; "X-Correlation-ID"?: string}; body: Schemas["shared.http_api.schemas.EmptyCommand"]}; responses: {200: Schemas["shared.http_api.schemas.Outcome"]; 400: Schemas["shared.http_api.schemas.Error"]; 401: Schemas["shared.http_api.schemas.Error"]; 403: Schemas["shared.http_api.schemas.Error"]; 404: Schemas["shared.http_api.schemas.RejectedOutcome"]; 409: Schemas["shared.http_api.schemas.RejectedOutcome"]; 422: Schemas["shared.http_api.schemas.RejectedOutcome"]; 428: Schemas["shared.http_api.schemas.Error"]; 503: Schemas["shared.http_api.schemas.Error"]}};
  "startPreparation": {input: {path: {"id": string}; query?: never; headers: {"Idempotency-Key": string; "If-Match": string; "X-Correlation-ID"?: string}; body: Schemas["shared.http_api.schemas.EmptyCommand"]}; responses: {200: Schemas["shared.http_api.schemas.Outcome"]; 400: Schemas["shared.http_api.schemas.Error"]; 401: Schemas["shared.http_api.schemas.Error"]; 403: Schemas["shared.http_api.schemas.Error"]; 404: Schemas["shared.http_api.schemas.RejectedOutcome"]; 409: Schemas["shared.http_api.schemas.RejectedOutcome"]; 422: Schemas["shared.http_api.schemas.RejectedOutcome"]; 428: Schemas["shared.http_api.schemas.Error"]; 503: Schemas["shared.http_api.schemas.Error"]}};
  "login": {input: {path?: never; query?: never; headers?: never; body: {"password": string}}; responses: {200: {"authenticated": true}; 400: Schemas["shared.http_api.schemas.Error"]; 401: Schemas["shared.http_api.schemas.Error"]; 403: Schemas["shared.http_api.schemas.Error"]; 503: Schemas["shared.http_api.schemas.Error"]}};
  "logout": {input: {path?: never; query?: never; headers?: never; body?: never}; responses: {200: {"authenticated": false}; 403: Schemas["shared.http_api.schemas.Error"]; 503: Schemas["shared.http_api.schemas.Error"]}};
  "session": {input: {path?: never; query?: never; headers?: never; body?: never}; responses: {200: {"subject": string; "name": string; "role": "operator"}; 401: {"authenticated": false}; 503: Schemas["shared.http_api.schemas.Error"]}};
  "apiDiagnostics": {input: {path?: never; query?: never; headers?: never; body?: never}; responses: {200: Schemas["services.api.http_api.operational.schemas.SessionDiagnostics"]; 401: Schemas["shared.http_api.schemas.Error"]; 503: Schemas["shared.http_api.schemas.Error"]}};
  "apiHealth": {input: {path?: never; query?: never; headers?: never; body?: never}; responses: {200: {"status": "ok"; "role": "client-api"}}};
};
export const routes = {
  "realtimeConnect": {
    "method": "POST",
    "path": "/api/realtime/connect"
  },
  "realtimeRefresh": {
    "method": "POST",
    "path": "/api/realtime/refresh"
  },
  "listPickups": {
    "method": "GET",
    "path": "/api/v1/collection/pickups"
  },
  "getPickup": {
    "method": "GET",
    "path": "/api/v1/collection/pickups/{id}"
  },
  "collectOrder": {
    "method": "POST",
    "path": "/api/v1/collection/pickups/{id}/collect"
  },
  "listNotifications": {
    "method": "GET",
    "path": "/api/v1/communication/notifications"
  },
  "getNotification": {
    "method": "GET",
    "path": "/api/v1/communication/notifications/{id}"
  },
  "listAccounts": {
    "method": "GET",
    "path": "/api/v1/loyalty/accounts"
  },
  "getAccount": {
    "method": "GET",
    "path": "/api/v1/loyalty/accounts/{id}"
  },
  "listRewards": {
    "method": "GET",
    "path": "/api/v1/loyalty/rewards"
  },
  "getReward": {
    "method": "GET",
    "path": "/api/v1/loyalty/rewards/{id}"
  },
  "redeemReward": {
    "method": "POST",
    "path": "/api/v1/loyalty/rewards/{id}/redeem"
  },
  "listDrinks": {
    "method": "GET",
    "path": "/api/v1/menu/drinks"
  },
  "getDrink": {
    "method": "GET",
    "path": "/api/v1/menu/drinks/{id}"
  },
  "createDrink": {
    "method": "POST",
    "path": "/api/v1/menu/drinks/{id}"
  },
  "publishDrink": {
    "method": "POST",
    "path": "/api/v1/menu/drinks/{id}/publish"
  },
  "reviseDrink": {
    "method": "POST",
    "path": "/api/v1/menu/drinks/{id}/revise"
  },
  "listEditions": {
    "method": "GET",
    "path": "/api/v1/menu/editions"
  },
  "getEdition": {
    "method": "GET",
    "path": "/api/v1/menu/editions/{id}"
  },
  "createEdition": {
    "method": "POST",
    "path": "/api/v1/menu/editions/{id}"
  },
  "addOffer": {
    "method": "POST",
    "path": "/api/v1/menu/editions/{id}/offers"
  },
  "changePrice": {
    "method": "POST",
    "path": "/api/v1/menu/editions/{id}/prices"
  },
  "publishEdition": {
    "method": "POST",
    "path": "/api/v1/menu/editions/{id}/publish"
  },
  "listOrders": {
    "method": "GET",
    "path": "/api/v1/ordering/orders"
  },
  "getOrder": {
    "method": "GET",
    "path": "/api/v1/ordering/orders/{id}"
  },
  "createOrder": {
    "method": "POST",
    "path": "/api/v1/ordering/orders/{id}"
  },
  "addLine": {
    "method": "POST",
    "path": "/api/v1/ordering/orders/{id}/lines"
  },
  "placeOrder": {
    "method": "POST",
    "path": "/api/v1/ordering/orders/{id}/place"
  },
  "changeQuantity": {
    "method": "POST",
    "path": "/api/v1/ordering/orders/{id}/quantities"
  },
  "listTickets": {
    "method": "GET",
    "path": "/api/v1/preparation/tickets"
  },
  "getTicket": {
    "method": "GET",
    "path": "/api/v1/preparation/tickets/{id}"
  },
  "completePreparation": {
    "method": "POST",
    "path": "/api/v1/preparation/tickets/{id}/complete"
  },
  "startPreparation": {
    "method": "POST",
    "path": "/api/v1/preparation/tickets/{id}/start"
  },
  "login": {
    "method": "POST",
    "path": "/auth/login"
  },
  "logout": {
    "method": "POST",
    "path": "/auth/logout"
  },
  "session": {
    "method": "GET",
    "path": "/auth/session"
  },
  "apiDiagnostics": {
    "method": "GET",
    "path": "/diagnostics"
  },
  "apiHealth": {
    "method": "GET",
    "path": "/healthz"
  }
} as const;
export type Operation = keyof Operations;
export type Input<K extends Operation> = Operations[K]["input"];
export type Success<K extends Operation> = Operations[K]["responses"][200];
export type CommandOperation = "collectOrder" | "redeemReward" | "createDrink" | "publishDrink" | "reviseDrink" | "createEdition" | "addOffer" | "changePrice" | "publishEdition" | "createOrder" | "addLine" | "placeOrder" | "changeQuantity" | "completePreparation" | "startPreparation";
export type ListOperation = "listPickups" | "listNotifications" | "listAccounts" | "listRewards" | "listDrinks" | "listEditions" | "listOrders" | "listTickets";
export type CommandBody<K extends CommandOperation> = Input<K>["body"];
export type Intersection<U> = (U extends unknown ? (value: U) => void : never) extends (value: infer I) => void ? I : never;
export type CommandParameters = Intersection<{[K in CommandOperation]: Omit<Input<K>, "body">}[CommandOperation]>;
export type SendCommand = <K extends CommandOperation>(operation: K, id: string, body: CommandBody<NoInfer<K>>, version: number) => Promise<boolean>;
