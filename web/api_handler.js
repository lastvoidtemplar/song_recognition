export class ApiHandler {
  #endpoint;
  #method;
  #body;
  #timeoutInterval;
  #countRetries;

  #isFetching;
  #statusCode;

  #loadingCallback;
  #successCallback;
  #errorCallback;
  #failCallback;

  constructor(endpoint, method, body = undefined, timeoutInterval = 5000) {
    this.#endpoint = endpoint;
    this.#method = method;
    this.#body = body;
    this.#timeoutInterval = timeoutInterval;
    this.#countRetries = 0;
    this.#isFetching = false;
  }

  onLoading(callback) {
    this.#loadingCallback = callback;
  }

  onSuccess(callback) {
    this.#successCallback = callback;
  }

  onError(callback) {
    this.#errorCallback = callback;
  }

  onFail(callback) {
    this.#failCallback = callback;
  }

  initiateFetch() {
    if (this.#isFetching === false) {
      this.#isFetching = true;
      typeof this.#loadingCallback === "function" && this.#loadingCallback();
      this.#fetch();
    }
  }

  #fetch() {
    fetch(this.#endpoint, {
      method: this.#method,
      body: this.#body,
      signal: AbortSignal.timeout(this.#timeoutInterval),
    })
      .then((resp) => {
        this.#statusCode = resp.status;
        return resp.json();
      })
      .then((data) => {
        if (200 <= this.#statusCode && this.#statusCode <= 299) {
          typeof this.#successCallback === "function" &&
            this.#successCallback(data);
          this.#isFetching = false;
        } else {
          typeof this.#errorCallback === "function" &&
            this.#errorCallback(this.#statusCode, data);
          this.#isFetching = false;
        }
      })
      .catch((err) => {
        if (err instanceof TypeError || err instanceof DOMException) {
          if (this.#countRetries < 2) {
            setTimeout(() => {
              this.#fetch();
            }, Math.pow(2, this.#countRetries) * 1000);
            this.#countRetries++;
            return;
          } else {
            typeof this.#failCallback === "function" && this.#failCallback(err);
            this.#isFetching = false;
            return;
          }
        }
        if (err instanceof SyntaxError) {
          typeof this.#failCallback === "function" && this.#failCallback(err);
          this.#isFetching = false;
          return;
        }
        this.#isFetching = false;
      });
  }
}
