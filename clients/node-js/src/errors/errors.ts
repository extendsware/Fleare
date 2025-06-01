export class FleareError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "FleareError";
  }
}

export class AuthenticationError extends FleareError {
  constructor(message: string) {
    super(message);
    this.name = "AuthenticationError";
  }
}
