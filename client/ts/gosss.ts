/**
 * Standard error response structure from GOSSS
 */
export interface GosssErrorResponse {
  code: string;
  message: string;
  resource: string;
  timestamp: string;
}

/**
 * Custom error class for GOSSS operations
 */
export class GosssError extends Error {
  readonly code: string;
  readonly timestamp: Date;
  readonly resource: string;

  constructor({ message, code, resource, timestamp }: GosssErrorResponse) {
    super(message);
    this.name = "GosssError";
    this.code = code;
    this.timestamp = new Date(timestamp);
    this.resource = resource;
  }
}

/**
 * Configuration options for the GOSSS S3 Client
 */
interface GOSSS3ClientOptions {
  /**
   * GOSSS endpoint URL where the GOSSS server is hosted
   * @example "http://localhost:8080"
   */
  readonly endpoint: string;

  /**
   * Authentication credentials for GOSSS
   */
  readonly credentials: {
    /**
     * GOSSS Access Key ID for authentication
     */
    readonly accessKeyId: string;
    /**
     * GOSSS Secret Access Key for authentication
     */
    readonly secretAccessKey: string;
  };
}

// Command Interfaces
interface GetObjectCommandInput {
  Bucket: string;
  Key: string;
  ResponseContentType?: string;
}

interface GetObjectCommandOutput {
  body: ReadableStream;
  contentType: string;
  contentLength: number;
  etag: string;
  lastModified?: Date;
}

interface PutObjectCommandInput {
  Bucket: string;
  Key: string;
  Body: Blob | ReadableStream | string;
  ContentType?: string;
}

interface PutObjectCommandOutput {
  key: string;
  size: number;
  lastModified: Date;
  etag: string;
  contentType: string;
}

type ObjectMetadata = {
  key: string;
  size: number;
  lastModified: Date;
  etag: string;
  contentType: string;
};

interface ListObjectsCommandInput {
  Bucket: string;
  Prefix?: string;
}

interface ListObjectsCommandOutput {
  name: string;
  prefix: string;
  contents: ObjectMetadata[];
}

export interface GosssCreateBucketOutput {
  Location?: string;
}

export type HeadBucketOutput = Record<string, never>;
export type DeleteBucketOutput = Record<string, never>;

// Command Classes
export class GetObjectCommand {
  readonly input: GetObjectCommandInput;

  constructor(input: GetObjectCommandInput) {
    this.input = input;
  }

  buildRequest(endpoint: string): Request {
    const url = new URL(`${endpoint}/${this.input.Bucket}/${this.input.Key}`);
    return new Request(url.toString(), { method: "GET" });
  }
}

export class PutObjectCommand {
  readonly input: PutObjectCommandInput;

  constructor(input: PutObjectCommandInput) {
    this.input = input;
  }

  buildRequest(endpoint: string): Request {
    const url = new URL(`${endpoint}/${this.input.Bucket}/${this.input.Key}`);
    const headers = new Headers();

    headers.set("Content-Type", "application/octet-stream");
    if (this.input.ContentType) {
      headers.set("Content-Type", this.input.ContentType);
    }

    return new Request(url.toString(), {
      method: "PUT",
      headers,
      body: this.input.Body,
    });
  }
}

export class ListObjectsCommand {
  readonly input: ListObjectsCommandInput;

  constructor(input: ListObjectsCommandInput) {
    this.input = input;
  }

  buildRequest(endpoint: string): Request {
    const url = new URL(`${endpoint}/${this.input.Bucket}`);
    const params = new URLSearchParams();
    url.search = params.toString();
    return new Request(url.toString(), { method: "GET" });
  }
}

// Main Client Classes
export class GosssS3Client {
  readonly options: GOSSS3ClientOptions;
  private readonly headers: Headers;

  constructor(options: GOSSS3ClientOptions) {
    this.options = options;
    this.headers = new Headers({
      Authorization: `${options.credentials.accessKeyId}=${options.credentials.secretAccessKey}`,
    });
  }

  async send(command: GetObjectCommand): Promise<GetObjectCommandOutput>;
  async send(command: PutObjectCommand): Promise<PutObjectCommandOutput>;
  async send(command: ListObjectsCommand): Promise<ListObjectsCommandOutput>;

  async send(
    command: GetObjectCommand | PutObjectCommand | ListObjectsCommand
  ): Promise<
    GetObjectCommandOutput | PutObjectCommandOutput | ListObjectsCommandOutput
  > {
    const request = command.buildRequest(this.options.endpoint);
    this.headers.forEach((value, key) => {
      request.headers.set(key, value);
    });

    const response = await fetch(request);

    if (!response.ok) {
      const err = (await response.json()) as GosssErrorResponse;
      throw new GosssError(err);
    }

    if (command instanceof PutObjectCommand) {
      return (await response.json()) as PutObjectCommandOutput;
    }

    if (command instanceof GetObjectCommand) {
      return {
        body: response.body!,
        contentType: response.headers.get("Content-Type") || "",
        contentLength: parseInt(
          response.headers.get("Content-Length") || "0",
          10
        ),
        etag: response.headers.get("Etag") || "",
        lastModified: new Date(response.headers.get("Last-Modified") || ""),
      };
    }

    if (command instanceof ListObjectsCommand) {
      return (await response.json()) as ListObjectsCommandOutput;
    }

    throw new GosssError({
      message: "Unsupported command type",
      code: "500",
      resource: "",
      timestamp: new Date().toISOString(),
    });
  }
}

export class GosssSDKS3 {
  private readonly options: GOSSS3ClientOptions;
  private readonly headers: Headers;

  constructor(options: GOSSS3ClientOptions) {
    this.options = options;
    this.headers = new Headers({
      Authorization: `${options.credentials.accessKeyId}=${options.credentials.secretAccessKey}`,
    });
  }

  async createBucket(
    params: { bucket: string },
    callback: (
      error: GosssError | null,
      data: GosssCreateBucketOutput | undefined
    ) => void
  ): Promise<void> {
    try {
      const response = await fetch(
        `${this.options.endpoint}/${params.bucket}`,
        {
          method: "PUT",
          headers: this.headers,
        }
      );

      if (!response.ok) {
        const errorData = (await response.json()) as GosssErrorResponse;
        callback(new GosssError(errorData), undefined);
        return;
      }

      callback(null, {
        Location: `${this.options.endpoint}/${params.bucket}`,
      });
    } catch (error: any) {
      callback(
        new GosssError({
          message: error.message,
          code: "500",
          resource: params.bucket,
          timestamp: new Date().toISOString(),
        }),
        undefined
      );
    }
  }

  async headBucket(
    params: { bucket: string },
    callback: (
      error: GosssError | null,
      data: HeadBucketOutput | undefined
    ) => void
  ): Promise<void> {
    try {
      const response = await fetch(
        `${this.options.endpoint}/${params.bucket}`,
        {
          method: "HEAD",
          headers: this.headers,
        }
      );

      if (!response.ok) {
        callback(
          new GosssError({
            code: response.status.toString(),
            message: `Bucket '${params.bucket}' not found`,
            resource: params.bucket,
            timestamp: new Date().toISOString(),
          }),
          undefined
        );
        return;
      }

      callback(null, {});
    } catch (error: any) {
      callback(
        new GosssError({
          message: error.message,
          code: "500",
          resource: params.bucket,
          timestamp: new Date().toISOString(),
        }),
        undefined
      );
    }
  }

  async deleteBucket(
    params: { bucket: string },
    callback: (
      error: GosssError | null,
      data: DeleteBucketOutput | undefined
    ) => void
  ): Promise<void> {
    try {
      const response = await fetch(
        `${this.options.endpoint}/${params.bucket}`,
        {
          method: "DELETE",
          headers: this.headers,
        }
      );

      if (!response.ok) {
        const errorData = (await response.json()) as GosssErrorResponse;
        callback(new GosssError(errorData), undefined);
        return;
      }

      callback(null, {});
    } catch (error: any) {
      callback(
        new GosssError({
          message: error.message,
          code: "500",
          resource: params.bucket,
          timestamp: new Date().toISOString(),
        }),
        undefined
      );
    }
  }
}
