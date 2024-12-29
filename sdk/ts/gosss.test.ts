import { describe, expect, test, mock, beforeEach } from "bun:test";
import {
  GosssS3Client,
  GosssSDKS3,
  GetObjectCommand,
  PutObjectCommand,
  ListObjectsCommand,
  GosssError,
} from "./gosss";

describe("GOSSS", () => {
  const mockOptions = {
    endpoint: "http://localhost:8191",
    credentials: {
      accessKeyId: "test-key",
      secretAccessKey: "test-secret",
    },
  };

  const originalFetch = global.fetch;

  beforeEach(() => {
    global.fetch = originalFetch;
  });

  describe("GosssS3Client", () => {
    let client: GosssS3Client;

    beforeEach(() => {
      client = new GosssS3Client(mockOptions);
    });

    describe("PutObjectCommand", () => {
      test("successfully puts an object", async () => {
        const mockResponse = {
          key: "test/file.txt",
          size: 11,
          lastModified: new Date().toISOString(),
          etag: "123456",
          contentType: "text/plain",
        };

        global.fetch = mock(async () => {
          return new Response(JSON.stringify(mockResponse), {
            status: 200,
            headers: { "Content-Type": "application/json" },
          });
        });

        const command = new PutObjectCommand({
          Bucket: "testBucket",
          Key: "test/file.txt",
          Body: "Hello World",
          ContentType: "text/plain",
        });

        const result = await client.send(command);
        expect(result).toEqual(mockResponse);
      });

      test("handles put object errors", async () => {
        const errorResponse = {
          code: "NoSuchBucket",
          message: "The specified bucket does not exist",
          resource: "testBucket",
          timestamp: new Date().toISOString(),
        };

        global.fetch = mock(async () => {
          return new Response(JSON.stringify(errorResponse), {
            status: 404,
            headers: { "Content-Type": "application/json" },
          });
        });

        const command = new PutObjectCommand({
          Bucket: "nonexistentBucket",
          Key: "test.txt",
          Body: "Hello World",
        });

        await expect(client.send(command)).rejects.toThrow(GosssError);
      });
    });

    describe("GetObjectCommand", () => {
      test("successfully gets an object", async () => {
        const mockBody = "Hello World";
        const mockStream = new ReadableStream({
          start(controller) {
            controller.enqueue(new TextEncoder().encode(mockBody));
            controller.close();
          },
        });

        global.fetch = mock(async () => {
          return new Response(mockStream, {
            status: 200,
            headers: {
              "Content-Type": "text/plain",
              "Content-Length": "11",
              ETag: "123456",
              "Last-Modified": new Date().toISOString(),
            },
          });
        });

        const command = new GetObjectCommand({
          Bucket: "testBucket",
          Key: "test/file.txt",
        });

        const result = await client.send(command);
        expect(result.contentType).toBe("text/plain");
        expect(result.contentLength).toBe(11);
        expect(result.etag).toBe("123456");
        expect(result.body).toBeDefined();
      });

      test("handles get object errors", async () => {
        const errorResponse = {
          code: "NoSuchKey",
          message: "The specified key does not exist",
          resource: "testBucket/test.txt",
          timestamp: new Date().toISOString(),
        };

        global.fetch = mock(async () => {
          return new Response(JSON.stringify(errorResponse), {
            status: 404,
            headers: { "Content-Type": "application/json" },
          });
        });

        const command = new GetObjectCommand({
          Bucket: "testBucket",
          Key: "nonexistent.txt",
        });

        await expect(client.send(command)).rejects.toThrow(GosssError);
      });
    });

    describe("ListObjectsCommand", () => {
      test("successfully lists objects", async () => {
        const mockResponse = {
          name: "testBucket",
          prefix: "",
          contents: [
            {
              key: "test/file1.txt",
              size: 11,
              lastModified: new Date(),
              etag: "123456",
              contentType: "text/plain",
            },
          ],
        };

        global.fetch = mock(async () => {
          return new Response(JSON.stringify(mockResponse), {
            status: 200,
            headers: { "Content-Type": "application/json" },
          });
        });

        const command = new ListObjectsCommand({
          Bucket: "testBucket",
        });

        const result = await client.send(command);
        expect(result.name).toBe("testBucket");
        expect(result.contents).toHaveLength(1);
        expect(result.contents[0].key).toBe("test/file1.txt");
      });
    });

    describe("Authorization", () => {
      test("includes correct authorization header", async () => {
        let capturedHeaders: Headers;

        global.fetch = mock(async (request: Request) => {
          capturedHeaders = request.headers;
          return new Response(JSON.stringify({}), {
            status: 200,
            headers: { "Content-Type": "application/json" },
          });
        });

        const command = new ListObjectsCommand({
          Bucket: "testBucket",
        });

        await client.send(command);

        const expectedAuth = `${mockOptions.credentials.accessKeyId}=${mockOptions.credentials.secretAccessKey}`;
        expect(capturedHeaders!.get("Authorization")).toBe(expectedAuth);
      });
    });
  });

  describe("GosssSDKS3", () => {
    let s3: GosssSDKS3;

    beforeEach(() => {
      s3 = new GosssSDKS3(mockOptions);
    });

    describe("createBucket", () => {
      test("successfully creates a bucket", async () => {
        global.fetch = mock(async () => {
          return new Response(null, { status: 200 });
        });

        await s3.createBucket({ bucket: "test-bucket" }, (error, data) => {
          expect(error).toBeNull();
          expect(data).toEqual({
            Location: "http://localhost:8191/test-bucket",
          });
        });
      });

      test("handles server error", async () => {
        const errorResponse = {
          code: "500",
          message: "Internal Server Error",
          resource: "test-bucket",
          timestamp: new Date().toISOString(),
        };

        global.fetch = mock(async () => {
          return new Response(JSON.stringify(errorResponse), {
            status: 500,
            headers: { "Content-Type": "application/json" },
          });
        });

        await s3.createBucket({ bucket: "test-bucket" }, (error, data) => {
          expect(error).toBeInstanceOf(GosssError);
          expect(error?.code).toBe("500");
          expect(data).toBeUndefined();
        });
      });
    });

    describe("headBucket", () => {
      test("successfully checks bucket existence", async () => {
        global.fetch = mock(async () => {
          return new Response(null, { status: 200 });
        });

        await s3.headBucket({ bucket: "test-bucket" }, (error, data) => {
          expect(error).toBeNull();
          expect(data).toEqual({});
        });
      });

      test("handles non-existent bucket", async () => {
        global.fetch = mock(async () => {
          return new Response(null, { status: 404 });
        });

        await s3.headBucket({ bucket: "test-bucket" }, (error, data) => {
          expect(error).toBeInstanceOf(GosssError);
          expect(error?.code).toBe("404");
          expect(data).toBeUndefined();
        });
      });
    });

    describe("deleteBucket", () => {
      test("successfully deletes a bucket", async () => {
        global.fetch = mock(async () => {
          return new Response(null, { status: 200 });
        });

        await s3.deleteBucket({ bucket: "test-bucket" }, (error, data) => {
          expect(error).toBeNull();
          expect(data).toEqual({});
        });
      });

      test("handles deletion error", async () => {
        const errorResponse = {
          code: "409",
          message: "Bucket not empty",
          resource: "test-bucket",
          timestamp: new Date().toISOString(),
        };

        global.fetch = mock(async () => {
          return new Response(JSON.stringify(errorResponse), {
            status: 409,
            headers: { "Content-Type": "application/json" },
          });
        });

        await s3.deleteBucket({ bucket: "test-bucket" }, (error, data) => {
          expect(error).toBeInstanceOf(GosssError);
          expect(error?.code).toBe("409");
          expect(data).toBeUndefined();
        });
      });
    });
  });
});
