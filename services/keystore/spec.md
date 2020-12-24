## Keystore Spec

### Problem

We need a keystore service that supports non-custodial applications.
It will make the process of diamnetizing any applications easier as
they don't have to implement the logic to create a diamnet account
and handle the encrypted private key themselves.

It is also intended to be the service that wallet SDK talks to.

### Authentication

For simplicity we will have each application spin up their own keystore
server, so there won’t be any routing logic in the keystore server that
directs requests to the correct client server to authenticate. Since we
don’t anticipate a lot of requests to the keystore from each user, we
should be able to tolerate having another round trip for relaying the
auth token to the client server.

<img src=attachments/2019-07-10-keystore-auth.png>

Keystore will forward two header fields, *Authorization* and *Cookie*, to the
designated endpoint on the client server with an extra header field
*X-Forwarded-For* specifying the request's origin. At this moment, keystore
forwards incoming requests by using HTTP GET method. We plan on adding the
support for clients who use GraphQL to authenticate in the future.

Clients are expected to put their auth tokens in one of the request header
fields. For example, those who use a bearer token to authenticate should have an
*Authorization* header in the following format:

```
Authorization: Bearer <token>
```

As mentioned above, clients will have to configure a API endpoint on their
servers used for authentication when booting up the keystore. For those who
choose to autheticate via a REST endpoint, keystore expects to receive a
response in the following json format:

```json
{
	"userID": "some-user-id"
}
```

Requests that the keystore is not able to derive a userID from will
receive the following error:

*not_authorized:*
```json
{
	"type": "not_authorized",
	"title": "Not Authorized",
	"status": 401,
	"detail": "The request is not authorized."
}
```

### Raw Key Data

*RawKeyData Object:*

```typescript
interface RawKeyData {
	keyType: string;
	publicKey: string;
	privateKey: string;
	path?: string;
	extra?: any;
}
```

```typescript
type RawKeys = RawKeyData[]
```

The clients will encrypt RawKeys with a salt based on the encrypter they use.
The clients will transmit the encrypted keys blob as a base64 URL encoded string.

### Encrypted Key Data

*EncryptedKeysData Object:*

```typescript
interface EncryptedKeysData {
	encrypterName: string;
	salt: string;
	keysBlob: string;
	creationTime: number;
	modifiedTime: number;	
}
```

We support three different kinds of HTTP methods to manipulate keys:

### PUT /keys

Put Keys Request:

```typescript
interface PutKeysRequest {
	encrypterName: string;
	salt: string;
	keysBlob: string;
}
```

Put Keys Response:

```typescript
type PutKeysResponse = EncryptedKeysData;
```

<details><summary>Errors</summary>

*bad_request:*
```json
{
	"keysBlob": "",
	"salt": "some-salt",
	"encrypterName": "identity"
}
```
```json
{
	"type": "bad_request",
	"title": "Bad Request",
	"status": 400,
	"detail": "The request you sent was invalid in some way.",
	"extras": {
		"invalid_field": "keysBlob",
		"reason": "field value cannot be empty"
	}
}
```
<hr />

*bad_request:*
```json
{
	"keysBlob": "some-base64-encoded-blob",
	"salt": "",
	"encrypterName": "identity"
}
```
```json
{
	"type": "bad_request",
	"title": "Bad Request",
	"status": 400,
	"detail": "The request you sent was invalid in some way.",
	"extras": {
		"invalid_field": "salt",
		"reason": "field value cannot be empty"
	}
}
```
<hr />

*bad_request:*
```json
{
	"keysBlob": "some-base64-encoded-blob",
	"salt": "some-salt",
	"encrypterName": ""
}
```
```json
{
	"type": "bad_request",
	"title": "Bad Request",
	"status": 400,
	"detail": "The request you sent was invalid in some way.",
	"extras": {
		"invalid_field": "encrypterName",
		"reason": "field value cannot be empty"
	}
}
```
<hr />

*invalid_keys_blob:*
```json
{
	"keysBlob": "some-badly-encoded-blob",
	"salt": "some-salt",
	"encrypterName": "identity"
}
```
```json
{
	"type": "invalid_keys_blob",
	"title": "Invalid Keys Blob",
	"status": 400,
	"detail": "The keysBlob in your request body is not a valid base64
		string. Please encode the keysBlob in your request body as a base64
		string properly and try again."
}
```
</details>

### GET /keys

Get Keys Request:

This endpoint will return the keys blob corresponding to the auth token
in the request header, if the token is valid. This endpoint does not take
any parameter.

Get Keys Response:

```typescript
type GetKeysResponse = EncryptedKeysData;
```
<details><summary>Errors</summary>

*not_found:*

The keystore cannot find any keys assocaited with the derived userID.
```json
{
	"type": "not_found",
	"title": "Resourse Missing",
	"status": 404,
	"detail": "The resource at the url requested was not found. This
		usually occurs for one of two reasons:  The url requested is not valid,
		or no data in our database could be found with the parameters
		provided."
}
```
</details>

### DELETE /keys

Delete Keys Request:

This endpoint will delete the keys blob corresponding to the auth token
in the request header and return the deleted keys blob to the client, if
the token is valid. This endpoint does not take any parameter.

Delete Keys Response:

*Success:*

```typescript
interface Success {
	message: "ok";
}
```

<details><summary>Errors</summary>
</details>
