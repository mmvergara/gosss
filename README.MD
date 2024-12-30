Created: Dec 29, 2024

<h1 align="center">Golang Simple Storage System (GOSSS)</h1>

<p align="center">
<img src="logo.png" height="200">
</p>

<p align="center">
S3 Like Storage System written in vanilla Golang <br/> with a typescript SDK designed for local development and testing.
</p>

#

> Thought it is fully unit tested, This is a hobby project is not intended for production use, use at your own risk.

## Features

### S3 Like Operations

- Create Bucket
- Delete Bucket
- Head Bucket
- Put Object
- Get Object
- Delete Object
- List Objects
- Get Signed Object URL

## Build and Deploy

- Clone the repository
- Fill the `.env` file

### Run locally

```bash
# install dependencies
go mod tidy

# run the server
go run cmd/server/main.go

# using air for hot reload
air
```

### Docker

```bash
docker build -t gosss .

docker run -p 8191:8191 gosss
```

Defaults:

- PORT = `8191`
- AccessKey: `test_id`
- GOSSSSecret Key: `test_key`

storage path is ./data by default, you can change this in the `./internal/config/config.go` file. and make sure to update dockerfile accordingly.

---

# Clients

- [Typescript SDK](#Typescript-SDK) just copy the `./sdk/ts/gosss.ts` file to your project.

- You can open `./client.html` to interact with the API via the browser.

# Typescript SDK

This SDK allows you to interact with the GOSSS S3-like API to perform operations like creating buckets, uploading, downloading, listing, and deleting objects. It also includes the functionality to generate signed URLs for temporary access to objects.

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
  - [Creating an Instance](#creating-an-instance)
  - [Uploading an Object](#uploading-an-object)
  - [Downloading an Object](#downloading-an-object)
  - [Listing Objects](#listing-objects)
  - [Deleting an Object](#deleting-an-object)
  - [Creating a Bucket](#creating-a-bucket)
  - [Getting a Signed URL](#getting-a-signed-url)
  - [Handling Errors](#handling-errors)
- [API Reference](#api-reference)

---

## Installation

Just copy the `./sdk/ts/gosss.ts` file to your project.

---

## Usage

### Creating an Instance

To interact with the GOSSS S3 API, first, create an instance of the `GosssS3Client` class.

```ts
import { GosssS3Client } from "./GosssS3Client";

const client = new GosssS3Client({
  endpoint: "http://localhost:8080", // GOSSS API endpoint
  credentials: {
    accessKeyId: "your-access-key-id",
    secretAccessKey: "your-secret-access-key",
  },
});
```

### Uploading an Object

To upload an object, use the `PutObjectCommand`:

```ts
import { PutObjectCommand } from "./GosssS3Client";

const command = new PutObjectCommand({
  Bucket: "your-bucket",
  Key: "file-name.txt",
  Body: "Hello, GOSSS!", // Blob, ReadableStream, or string
  ContentType: "text/plain",
});

const response = await client.send(command);
console.log(response);
```

### Downloading an Object

To download an object, use the `GetObjectCommand`:

```ts
import { GetObjectCommand } from "./GosssS3Client";

const command = new GetObjectCommand({
  Bucket: "your-bucket",
  Key: "file-name.txt",
});

const response = await client.send(command);
const content = await response.body.text();
console.log(content);
```

### Listing Objects

To list objects in a bucket, use the `ListObjectsCommand`:

```ts
import { ListObjectsCommand } from "./GosssS3Client";

const command = new ListObjectsCommand({
  Bucket: "your-bucket",
});

const response = await client.send(command);
console.log(response.contents);
```

### Deleting an Object

To delete an object from a bucket, use the `DeleteObjectCommand`:

```ts
import { DeleteObjectCommand } from "./GosssS3Client";

const command = new DeleteObjectCommand({
  Bucket: "your-bucket",
  Key: "file-name.txt",
});

const response = await client.send(command);
console.log("Deleted:", response);
```

### Creating a Bucket

To create a new bucket, use the `GosssSDKS3` class:

```ts
import { GosssSDKS3 } from "./GosssS3Client";

const sdk = new GosssSDKS3({
  endpoint: "http://localhost:8080",
  credentials: {
    accessKeyId: "your-access-key-id",
    secretAccessKey: "your-secret-access-key",
  },
});

sdk.createBucket({ bucket: "new-bucket" }, (error, data) => {
  if (error) {
    console.error("Error creating bucket:", error);
  } else {
    console.log("Bucket created:", data);
  }
});
```

### Getting a Signed URL

To generate a signed URL for temporary access to an object, use the `getSignedUrl` function:

```ts
import { getSignedUrl } from "./GosssS3Client";
import { GetObjectCommand } from "./GosssS3Client";

const command = new GetObjectCommand({
  Bucket: "your-bucket",
  Key: "file-name.txt",
});

const url = await getSignedUrl(client, command, { expiresIn: 3600 });
console.log("Signed URL:", url);
```

---

## Handling Errors

The SDK uses a custom `GosssError` class to handle errors. Errors will be thrown for failed requests, and you can catch them as follows:

```ts
try {
  const response = await client.send(command);
} catch (error) {
  if (error instanceof GosssError) {
    console.error("Error code:", error.code);
    console.error("Error message:", error.message);
    console.error("Resource:", error.resource);
    console.error("Timestamp:", error.timestamp);
  }
}
```

---

## API Reference

### `GosssErrorResponse`

Defines the standard error structure returned by the GOSSS API.

- `code`: The error code.
- `message`: A descriptive error message.
- `resource`: The resource related to the error.
- `timestamp`: The time the error occurred.

### `GosssError`

Custom error class that extends `Error`.

- `code`: The error code.
- `message`: The error message.
- `resource`: The resource associated with the error.
- `timestamp`: The timestamp of when the error occurred.

### `GOSSS3ClientOptions`

Configuration options for the `GosssS3Client`.

- `endpoint`: The GOSSS server endpoint URL.
- `credentials`: The credentials used for authentication (access key and secret key).

### Commands

- `PutObjectCommand`: Upload an object to a bucket.
- `GetObjectCommand`: Download an object from a bucket.
- `ListObjectsCommand`: List objects in a bucket.
- `DeleteObjectCommand`: Delete an object from a bucket.

### `GosssSDKS3`

Provides callback-based operations for managing buckets:

- `createBucket()`
- `headBucket()`
- `deleteBucket()`

### `getSignedUrl()`

Generates a signed URL for temporary access to an object in a bucket.

- `expiresIn`: The number of seconds for which the URL is valid.

---

### Signed URLs Implementation

On the client-side, the function getSignedUrl is in charge of generating the signed URL. It combines the expiration, bucket, and object key into a string, signs it with HMAC-SHA256, and appends everything into a URL. This URL is then used to securely access the object.

On the server-side, we have two important functions. The first, generateSignature, takes an expiration time, the bucket name, and the object key, and creates a secure signature using HMAC-SHA256. This signature acts like a unique "stamp" that ensures no one can tamper with the URL.

The second function, GetSignedObject, is the one that actually handles requests for signed URLs. It checks the expiration time and verifies the signature. If everything looks good, it retrieves the object from the storage and streams it back to the client.

# License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
