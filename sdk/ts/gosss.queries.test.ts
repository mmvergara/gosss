import { describe, expect, test } from "bun:test";
import fs from "node:fs";
import {
  GosssS3Client,
  GosssSDKS3,
  GetObjectCommand,
  PutObjectCommand,
  ListObjectsCommand,
  getSignedUrl,
  DeleteObjectCommand,
} from "./gosss";

const CONFIG = {
  endpoint: "http://localhost:8191",
  credentials: {
    accessKeyId: "test_id",
    secretAccessKey: "test_key",
  },
};

// Test bucket names
const TEST_BUCKET = "test-bucket";
const TEST_KEY = "test-file.txt";
const TEST_IMAGE_KEY = "test-image.png";

describe("GOSSS S3 Client Tests", () => {
  const client = new GosssS3Client(CONFIG);
  const sdkClient = new GosssSDKS3(CONFIG);

  // Helper function to create test bucket
  const createTestBucket = () => {
    return new Promise((resolve, reject) => {
      sdkClient.headBucket({ bucket: TEST_BUCKET }, (err, data) => {
        if (err) {
          sdkClient.createBucket({ bucket: TEST_BUCKET }, (error, data) => {
            if (error) reject(error);
            else resolve(data);
          });
        } else {
          resolve(data);
        }
      });
    });
  };

  // Helper function to delete test bucket, this expects the bucket to be present
  const deleteTestBucket = async () => {
    // Get all objects in the bucket
    const listObjectsCommand = new ListObjectsCommand({
      Bucket: TEST_BUCKET,
    });

    const listObjectsResult = await client.send(listObjectsCommand);
    if (listObjectsResult.contents != null) {
      const objectKeys = listObjectsResult.contents.map((obj) => obj.key);

      // Delete all objects in the bucket
      for (const key of objectKeys) {
        const deleteObjectCommand = new DeleteObjectCommand({
          Bucket: TEST_BUCKET,
          Key: key,
        });

        await client.send(deleteObjectCommand);
      }
    }

    return new Promise((resolve, reject) => {
      // Delete the bucket
      sdkClient.deleteBucket({ bucket: TEST_BUCKET }, (error, data) => {
        if (error) reject(error);
        else resolve(data);
      });
    });
  };

  test("Bucket Operations", async () => {
    // Create bucket
    await createTestBucket();

    // Check if bucket exists
    await new Promise((resolve, reject) => {
      sdkClient.headBucket({ bucket: TEST_BUCKET }, (error, data) => {
        expect(error).toBeNull();
        resolve(data);
      });
    });

    // Clean up
    await deleteTestBucket();
  });

  test("Object Operations", async () => {
    // Create bucket first
    await createTestBucket();

    // Test file upload
    const textContent = "Hello, GOSSS!";
    const textBlob = new Blob([textContent], { type: "text/plain" });

    const putTextCommand = new PutObjectCommand({
      Bucket: TEST_BUCKET,
      Key: TEST_KEY,
      Body: textBlob,
      ContentType: "text/plain",
    });

    const putResult = await client.send(putTextCommand);
    expect(putResult.key).toBe(TEST_KEY);

    // Test image upload
    const imageBlob = Bun.file("local.jpg");
    const putImageCommand = new PutObjectCommand({
      Bucket: TEST_BUCKET,
      Key: TEST_IMAGE_KEY,
      Body: imageBlob,
      ContentType: "image/png",
    });

    const putImageResult = await client.send(putImageCommand);
    expect(putImageResult.key).toBe(TEST_IMAGE_KEY);

    // Test list objects
    const listCommand = new ListObjectsCommand({
      Bucket: TEST_BUCKET,
    });

    const listResult = await client.send(listCommand);
    expect(listResult.contents.length).toBe(2);

    // Test get object (text file)
    const getTextCommand = new GetObjectCommand({
      Bucket: TEST_BUCKET,
      Key: TEST_KEY,
    });

    const getTextResult = await client.send(getTextCommand);
    const textReader = getTextResult.body.getReader();
    const { value } = await textReader.read();
    const downloadedText = new TextDecoder().decode(value);
    expect(downloadedText).toBe(textContent);

    // Test signed URL generation
    const signedUrl = await getSignedUrl(client, getTextCommand, {
      expiresIn: 3600,
    });
    expect(signedUrl).not.toBeNull();

    // Clean up
    await deleteTestBucket();
  });

  test("Error Handling", async () => {
    // Test non-existent bucket
    await expect(
      client.send(
        new GetObjectCommand({
          Bucket: "non-existent-bucket",
          Key: "non-existent-file",
        })
      )
    ).rejects.toThrow();

    // Test invalid credentials
    const invalidClient = new GosssS3Client({
      ...CONFIG,
      credentials: {
        accessKeyId: "invalid",
        secretAccessKey: "invalid",
      },
    });

    await expect(
      invalidClient.send(
        new GetObjectCommand({
          Bucket: TEST_BUCKET,
          Key: TEST_KEY,
        })
      )
    ).rejects.toThrow();
  });
});
