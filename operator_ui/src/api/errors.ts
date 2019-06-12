export interface ApiError {
  status: string
  detail: string
}

export function AuthenticationError(errors: ApiError[]) {
  this.errors = errors
}

export function BadRequestError(errors: ApiError[]) {
  this.errors = errors
}

export function ServerError(errors: ApiError[]) {
  this.errors = errors
}

export function UnknownResponseError(errors: ApiError[]) {
  this.errors = errors
}

export function JsonParseError(errors: ApiError[]) {
  this.errors = errors
}
