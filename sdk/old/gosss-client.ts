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

export class GetObjectCommand {
  readonly input: GetObjectCommandInput;

  constructor(input: GetObjectCommandInput) {
    this.input = input;
  }

  buildRequest(endpoint: string): Request {
    const url = new URL(`${endpoint}/${this.input.Bucket}/${this.input.Key}`);

    return new Request(url.toString(), {
      method: "GET",
    });
  }
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
export class ListObjectsCommand {
  readonly input: ListObjectsCommandInput;

  constructor(input: ListObjectsCommandInput) {
    this.input = input;
  }

  buildRequest(endpoint: string): Request {
    const url = new URL(`${endpoint}/${this.input.Bucket}`);
    const params = new URLSearchParams();

    url.search = params.toString();

    return new Request(url.toString(), {
      method: "GET",
    });
  }
}

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

  /**
   * General implementation of send
   * @param command - Command to execute
   * @returns Promise resolving to command output
   */
  async send(
    command: GetObjectCommand | PutObjectCommand | ListObjectsCommand
  ): Promise<
    GetObjectCommandOutput | PutObjectCommandOutput | ListObjectsCommandOutput
  > {
    const request = command.buildRequest(this.options.endpoint);

    // Add authorization headers
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
        body: response.body,
        contentType: response.headers.get("Content-Type") || undefined,
        contentLength: parseInt(
          response.headers.get("Content-Length") || "0",
          10
        ),
        etag: response.headers.get("Etag") || undefined,
        lastModified: new Date(response.headers.get("Last-Modified") || ""),
      } as GetObjectCommandOutput;
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
