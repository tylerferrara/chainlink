export function AuthenticationError(response: Response) {
  this.errors = [
    {
      status: response.status,
      detail: response.statusText
    }
  ]
}

interface IJsonApiError {
  errors: object
}

export function BadRequestError({ errors }: IJsonApiError) {
  this.errors = errors
}

export function ServerError(response: Response) {
  this.errors = [
    {
      status: response.status,
      detail: response.statusText
    }
  ]
}

export function UnknownResponseError(response: Response) {
  this.errors = [
    {
      status: response.status,
      detail: response.statusText
    }
  ]
}
