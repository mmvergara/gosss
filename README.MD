# Golang Simple Storage System (GOSSS)

S3 Like Storage System written in vanilla Golang designed for local development and testing.

> Thought it is fully tested, This is a hobby project is not intended for production use, use at your own risk.

## Features

### S3 Like Operations

- Create Bucket
- Delete Bucket
- Head Bucket
- Put Object
- Get Object
- Delete Object
- List Objects

## Typescript SDK Usage

- Copy the `./sdk/ts/gosss.ts` file to your project

### Bucket operations

```typescript
import { GosssSDKS3 } from "./gosss-sdk";

const s3 = new GosssSDKS3({
  endpoint: "http://localhost:8080", // Replace with your GOSSS endpoint
  credentials: {
    accessKeyId: "your-access-key-id",
    secretAccessKey: "your-secret-access-key",
  },
});

// Create a bucket
s3.createBucket({ bucket: "new-bucket" }, (error, data) => {
  if (error) {
    console.error("Error creating bucket:", error);
  } else {
    console.log("Bucket created:", data);
  }
});

// Head a bucket
s3.headBucket({ bucket: "new-bucket" }, (error, data) => {
  if (error) {
    console.error("Error checking bucket:", error);
  } else {
    console.log("Bucket exists:", data);
  }
});

// Delete a bucket
s3.deleteBucket({ bucket: "new-bucket" }, (error, data) => {
  if (error) {
    console.error("Error deleting bucket:", error);
  } else {
    console.log("Bucket deleted:", data);
  }
});
```

### Object operations

```typescript
import {
  GosssS3Client,
  GetObjectCommand,
  PutObjectCommand,
  ListObjectsCommand,
} from "./gosss-sdk";

// Initialize the client
const client = new GosssS3Client({
  endpoint: "http://localhost:8080", // Replace with your GOSSS endpoint
  credentials: {
    accessKeyId: "your-access-key-id",
    secretAccessKey: "your-secret-access-key",
  },
});

// Example of getting an object
const getObjectCommand = new GetObjectCommand({
  Bucket: "example-bucket",
  Key: "example-object-key",
});

client
  .send(getObjectCommand)
  .then((output) => {
    console.log(output);
  })
  .catch((error) => {
    console.error(error);
  });

// Example of putting an object
const putObjectCommand = new PutObjectCommand({
  Bucket: "example-bucket",
  Key: "example-object-key",
  Body: "Hello, GOSSS!",
});

client
  .send(putObjectCommand)
  .then((output) => {
    console.log(output);
  })
  .catch((error) => {
    console.error(error);
  });

// Example of listing objects
const listObjectsCommand = new ListObjectsCommand({
  Bucket: "example-bucket",
});

client
  .send(listObjectsCommand)
  .then((output) => {
    console.log(output);
  })
  .catch((error) => {
    console.error(error);
  });
```

### Signed URLs Implementation

On the client-side, the function getSignedUrl is in charge of generating the signed URL. It combines the expiration, bucket, and object key into a string, signs it with HMAC-SHA256, and appends everything into a URL. This URL is then used to securely access the object.

On the server-side, we have two important functions. The first, generateSignature, takes an expiration time, the bucket name, and the object key, and creates a secure signature using HMAC-SHA256. This signature acts like a unique "stamp" that ensures no one can tamper with the URL.

The second function, GetSignedObject, is the one that actually handles requests for signed URLs. It checks the expiration time and verifies the signature. If everything looks good, it retrieves the object from the storage and streams it back to the client.
