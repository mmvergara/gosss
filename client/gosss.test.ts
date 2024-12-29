import { describe, expect, test, mock, beforeEach } from "bun:test";
import { GosssSDKS3, GosssError } from "./gosss";

describe("GOSSS SDK", () => {
  const mockOptions = {
    endpoint: "http://localhost:8080",
    credentials: {
      accessKeyId: "test-key",
      secretAccessKey: "test-secret",
    },
  };

  let s3: GosssSDKS3;

  // Mock global fetch
  const originalFetch = global.fetch;

  beforeEach(() => {
    s3 = new GosssSDKS3(mockOptions);
    // Reset fetch mock
    global.fetch = originalFetch;
  });

  describe("createBucket", () => {
    test("successfully creates a bucket", async () => {
      global.fetch = mock(async () => {
        return new Response(null, { status: 200 });
      });

      await s3.createBucket({ bucket: "test-bucket" }, (error, data) => {
        expect(error).toBeNull();
        expect(data).toEqual({
          Location: "http://localhost:8080/test-bucket",
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
