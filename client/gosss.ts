/**
 * Configuration options for the GOSSS S3 Client
 */
export interface GOSSS3ClientOptions {
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

export interface GosssCreateBucketOutput {
  Location?: string;
}
export type HeadBucketOutput = Record<string, never>;
export type DeleteBucketOutput = Record<string, never>;

export class GosssSDKS3 {
  private readonly options: GOSSS3ClientOptions;
  private readonly headers: Headers;

  constructor(options: GOSSS3ClientOptions) {
    this.options = options;
    this.headers = new Headers({
      Authorization: `${options.credentials.accessKeyId}=${options.credentials.secretAccessKey}`,
    });
  }

  /**
   * Creates a new bucket
   * @param params - Parameters for creating a bucket
   * @param callback - Callback function to handle the response
   */
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

  /**
   * Checks if a bucket exists
   * @param params - Parameters for checking a bucket
   * @param callback - Callback function to handle the response
   */
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

  /**
   * Deletes a bucket
   * @param params - Parameters for deleting a bucket
   * @param callback - Callback function to handle the response
   */
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
