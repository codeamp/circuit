const async = require("async");
const discovery = require("clever-discovery");
const kayvee = require("kayvee");
const request = require("request");
const opentracing = require("opentracing");
const {commandFactory} = require("hystrixjs");
const RollingNumberEvent = require("hystrixjs/lib/metrics/RollingNumberEvent");

/**
 * @external Span
 * @see {@link https://doc.esdoc.org/github.com/opentracing/opentracing-javascript/class/src/span.js~Span.html}
 */

const { Errors } = require("./types");

/**
 * The exponential retry policy will retry five times with an exponential backoff.
 * @alias module:atlas-api-client.RetryPolicies.Exponential
 */
const exponentialRetryPolicy = {
  backoffs() {
    const ret = [];
    let next = 100.0; // milliseconds
    const e = 0.05; // +/- 5% jitter
    while (ret.length < 5) {
      const jitter = ((Math.random() * 2) - 1) * e * next;
      ret.push(next + jitter);
      next *= 2;
    }
    return ret;
  },
  retry(requestOptions, err, res) {
    if (err || requestOptions.method === "POST" ||
        requestOptions.method === "PATCH" ||
        res.statusCode < 500) {
      return false;
    }
    return true;
  },
};

/**
 * Use this retry policy to retry a request once.
 * @alias module:atlas-api-client.RetryPolicies.Single
 */
const singleRetryPolicy = {
  backoffs() {
    return [1000];
  },
  retry(requestOptions, err, res) {
    if (err || requestOptions.method === "POST" ||
        requestOptions.method === "PATCH" ||
        res.statusCode < 500) {
      return false;
    }
    return true;
  },
};

/**
 * Use this retry policy to turn off retries.
 * @alias module:atlas-api-client.RetryPolicies.None
 */
const noRetryPolicy = {
  backoffs() {
    return [];
  },
  retry() {
    return false;
  },
};

/**
 * Request status log is used to
 * to output the status of a request returned
 * by the client.
 */
function responseLog(logger, req, res, err) {
  var res = res || { };
  var req = req || { };
  var logData = {
	"backend": "atlas-api-client",
	"method": req.method || "",
	"uri": req.uri || "",
    "message": err || (res.statusMessage || ""),
    "status_code": res.statusCode || 0,
  };

  if (err) {
    logger.errorD("client-request-finished", logData);
  } else {
    logger.infoD("client-request-finished", logData);
  }
}

/**
 * Default circuit breaker options.
 * @alias module:atlas-api-client.DefaultCircuitOptions
 */
const defaultCircuitOptions = {
  forceClosed:            true,
  requestVolumeThreshold: 20,
  maxConcurrentRequests:  100,
  requestVolumeThreshold: 20,
  sleepWindow:            5000,
  errorPercentThreshold:  90,
  logIntervalMs:          30000
};

/**
 * atlas-api-client client library.
 * @module atlas-api-client
 * @typicalname AtlasAPIClient
 */

/**
 * atlas-api-client client
 * @alias module:atlas-api-client
 */
class AtlasAPIClient {

  /**
   * Create a new client object.
   * @param {Object} options - Options for constructing a client object.
   * @param {string} [options.address] - URL where the server is located. Must provide
   * this or the discovery argument
   * @param {bool} [options.discovery] - Use clever-discovery to locate the server. Must provide
   * this or the address argument
   * @param {number} [options.timeout] - The timeout to use for all client requests,
   * in milliseconds. This can be overridden on a per-request basis. Default is 5000ms.
   * @param {bool} [options.keepalive] - Set keepalive to true for client requests. This sets the
   * forever: true attribute in request. Defaults to false
   * @param {module:atlas-api-client.RetryPolicies} [options.retryPolicy=RetryPolicies.Single] - The logic to
   * determine which requests to retry, as well as how many times to retry.
   * @param {module:kayvee.Logger} [options.logger=logger.New("atlas-api-client-wagclient")] - The Kayvee
   * logger to use in the client.
   * @param {Object} [options.circuit] - Options for constructing the client's circuit breaker.
   * @param {bool} [options.circuit.forceClosed] - When set to true the circuit will always be closed. Default: true.
   * @param {number} [options.circuit.maxConcurrentRequests] - the maximum number of concurrent requests
   * the client can make at the same time. Default: 100.
   * @param {number} [options.circuit.requestVolumeThreshold] - The minimum number of requests needed
   * before a circuit can be tripped due to health. Default: 20.
   * @param {number} [options.circuit.sleepWindow] - how long, in milliseconds, to wait after a circuit opens
   * before testing for recovery. Default: 5000.
   * @param {number} [options.circuit.errorPercentThreshold] - the threshold to place on the rolling error
   * rate. Once the error rate exceeds this percentage, the circuit opens.
   * Default: 90.
   */
  constructor(options) {
    options = options || {};

    if (options.discovery) {
      try {
        this.address = discovery("atlas-api-client", "http").url();
      } catch (e) {
        this.address = discovery("atlas-api-client", "default").url();
      }
    } else if (options.address) {
      this.address = options.address;
    } else {
      throw new Error("Cannot initialize atlas-api-client without discovery or address");
    }
    if (options.keepalive) {
      this.keepalive = options.keepalive;
    } else {
      this.keepalive = false;
    }
    if (options.timeout) {
      this.timeout = options.timeout;
    } else {
      this.timeout = 5000;
    }
    if (options.retryPolicy) {
      this.retryPolicy = options.retryPolicy;
    }
    if (options.logger) {
      this.logger = options.logger;
    } else {
      this.logger =  new kayvee.logger("atlas-api-client-wagclient");
    }
    if (options.tracer) {
      this.tracer = options.tracer;
    } else {
      this.tracer = opentracing.globalTracer();
    }

    const circuitOptions = Object.assign({}, defaultCircuitOptions, options.circuit);
    this._hystrixCommand = commandFactory.getOrCreate("atlas-api-client").
      errorHandler(this._hystrixCommandErrorHandler).
      circuitBreakerForceClosed(circuitOptions.forceClosed).
      requestVolumeRejectionThreshold(circuitOptions.maxConcurrentRequests).
      circuitBreakerRequestVolumeThreshold(circuitOptions.requestVolumeThreshold).
      circuitBreakerSleepWindowInMilliseconds(circuitOptions.sleepWindow).
      circuitBreakerErrorThresholdPercentage(circuitOptions.errorPercentThreshold).
      timeout(0).
      statisticalWindowLength(10000).
      statisticalWindowNumberOfBuckets(10).
      run(this._hystrixCommandRun).
      context(this).
      build();

    setInterval(() => this._logCircuitState(), circuitOptions.logIntervalMs);
  }

  _hystrixCommandErrorHandler(err) {
    // to avoid counting 4XXs as errors, only count an error if it comes from the request library
    if (err._fromRequest === true) {
      return err;
    }
    return false;
  }

  _hystrixCommandRun(method, args) {
    return method.apply(this, args);
  }

  _logCircuitState(logger) {
    // code below heavily borrows from hystrix's internal HystrixSSEStream.js logic
    const metrics = this._hystrixCommand.metrics;
    const healthCounts = metrics.getHealthCounts()
    const circuitBreaker = this._hystrixCommand.circuitBreaker;
    this.logger.infoD("atlas-api-client", {
      "requestCount":                    healthCounts.totalCount,
      "errorCount":                      healthCounts.errorCount,
      "errorPercentage":                 healthCounts.errorPercentage,
      "isCircuitBreakerOpen":            circuitBreaker.isOpen(),
      "rollingCountFailure":             metrics.getRollingCount(RollingNumberEvent.FAILURE),
      "rollingCountShortCircuited":      metrics.getRollingCount(RollingNumberEvent.SHORT_CIRCUITED),
      "rollingCountSuccess":             metrics.getRollingCount(RollingNumberEvent.SUCCESS),
      "rollingCountTimeout":             metrics.getRollingCount(RollingNumberEvent.TIMEOUT),
      "currentConcurrentExecutionCount": metrics.getCurrentExecutionCount(),
      "latencyTotalMean":                metrics.getExecutionTime("mean") || 0,
    });
  }

  /**
   * Get All Clusters
   * @param {string} groupID
   * @param {object} [options]
   * @param {number} [options.timeout] - A request specific timeout
   * @param {external:Span} [options.span] - An OpenTracing span - For example from the parent request
   * @param {module:atlas-api-client.RetryPolicies} [options.retryPolicy] - A request specific retryPolicy
   * @param {function} [cb]
   * @returns {Promise}
   * @fulfill {Object}
   * @reject {module:atlas-api-client.Errors.BadRequest}
   * @reject {module:atlas-api-client.Errors.Unauthorized}
   * @reject {module:atlas-api-client.Errors.NotFound}
   * @reject {module:atlas-api-client.Errors.InternalError}
   * @reject {Error}
   */
  getClusters(groupID, options, cb) {
    return this._hystrixCommand.execute(this._getClusters, arguments);
  }
  _getClusters(groupID, options, cb) {
    const params = {};
    params["groupID"] = groupID;

    if (!cb && typeof options === "function") {
      cb = options;
      options = undefined;
    }

    return new Promise((resolve, reject) => {
      const rejecter = (err) => {
        reject(err);
        if (cb) {
          cb(err);
        }
      };
      const resolver = (data) => {
        resolve(data);
        if (cb) {
          cb(null, data);
        }
      };


      if (!options) {
        options = {};
      }

      const timeout = options.timeout || this.timeout;
      const tracer = options.tracer || this.tracer;
      const span = options.span;

      const headers = {};
      if (!params.groupID) {
        rejecter(new Error("groupID must be non-empty because it's a path parameter"));
        return;
      }

      const query = {};

      if (span) {
        // Need to get tracer to inject. Use HTTP headers format so we can properly escape special characters
        tracer.inject(span, opentracing.FORMAT_HTTP_HEADERS, headers);
        span.logEvent("GET /api/atlas/v1.0/groups/{groupID}/clusters");
        span.setTag("span.kind", "client");
      }

	  const requestOptions = {
        method: "GET",
        uri: this.address + "/api/atlas/v1.0/groups/" + params.groupID + "/clusters",
        json: true,
        timeout,
        headers,
        qs: query,
        useQuerystring: true,
      };
      if (this.keepalive) {
        requestOptions.forever = true;
      }
  

      const retryPolicy = options.retryPolicy || this.retryPolicy || singleRetryPolicy;
      const backoffs = retryPolicy.backoffs();
      const logger = this.logger;
  
      let retries = 0;
      (function requestOnce() {
        request(requestOptions, (err, response, body) => {
          if (retries < backoffs.length && retryPolicy.retry(requestOptions, err, response, body)) {
            const backoff = backoffs[retries];
            retries += 1;
            setTimeout(requestOnce, backoff);
            return;
          }
          if (err) {
            err._fromRequest = true;
            responseLog(logger, requestOptions, response, err)
            rejecter(err);
            return;
          }

          switch (response.statusCode) {
            case 200:
              resolver(body);
              break;
            
            case 400:
              var err = new Errors.BadRequest(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 401:
              var err = new Errors.Unauthorized(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 404:
              var err = new Errors.NotFound(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 500:
              var err = new Errors.InternalError(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            default:
              var err = new Error("Received unexpected statusCode " + response.statusCode);
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
          }
        });
      }());
    });
  }

  /**
   * Create a Cluster
   * @param {Object} params
   * @param {string} params.groupID
   * @param params.createOrUpdateClusterRequest
   * @param {object} [options]
   * @param {number} [options.timeout] - A request specific timeout
   * @param {external:Span} [options.span] - An OpenTracing span - For example from the parent request
   * @param {module:atlas-api-client.RetryPolicies} [options.retryPolicy] - A request specific retryPolicy
   * @param {function} [cb]
   * @returns {Promise}
   * @fulfill {Object}
   * @reject {module:atlas-api-client.Errors.BadRequest}
   * @reject {module:atlas-api-client.Errors.Unauthorized}
   * @reject {module:atlas-api-client.Errors.NotFound}
   * @reject {module:atlas-api-client.Errors.InternalError}
   * @reject {Error}
   */
  createCluster(params, options, cb) {
    return this._hystrixCommand.execute(this._createCluster, arguments);
  }
  _createCluster(params, options, cb) {
    if (!cb && typeof options === "function") {
      cb = options;
      options = undefined;
    }

    return new Promise((resolve, reject) => {
      const rejecter = (err) => {
        reject(err);
        if (cb) {
          cb(err);
        }
      };
      const resolver = (data) => {
        resolve(data);
        if (cb) {
          cb(null, data);
        }
      };


      if (!options) {
        options = {};
      }

      const timeout = options.timeout || this.timeout;
      const tracer = options.tracer || this.tracer;
      const span = options.span;

      const headers = {};
      if (!params.groupID) {
        rejecter(new Error("groupID must be non-empty because it's a path parameter"));
        return;
      }

      const query = {};

      if (span) {
        // Need to get tracer to inject. Use HTTP headers format so we can properly escape special characters
        tracer.inject(span, opentracing.FORMAT_HTTP_HEADERS, headers);
        span.logEvent("POST /api/atlas/v1.0/groups/{groupID}/clusters");
        span.setTag("span.kind", "client");
      }

	  const requestOptions = {
        method: "POST",
        uri: this.address + "/api/atlas/v1.0/groups/" + params.groupID + "/clusters",
        json: true,
        timeout,
        headers,
        qs: query,
        useQuerystring: true,
      };
      if (this.keepalive) {
        requestOptions.forever = true;
      }
  
      requestOptions.body = params.createOrUpdateClusterRequest;
  

      const retryPolicy = options.retryPolicy || this.retryPolicy || singleRetryPolicy;
      const backoffs = retryPolicy.backoffs();
      const logger = this.logger;
  
      let retries = 0;
      (function requestOnce() {
        request(requestOptions, (err, response, body) => {
          if (retries < backoffs.length && retryPolicy.retry(requestOptions, err, response, body)) {
            const backoff = backoffs[retries];
            retries += 1;
            setTimeout(requestOnce, backoff);
            return;
          }
          if (err) {
            err._fromRequest = true;
            responseLog(logger, requestOptions, response, err)
            rejecter(err);
            return;
          }

          switch (response.statusCode) {
            case 201:
              resolver(body);
              break;
            
            case 400:
              var err = new Errors.BadRequest(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 401:
              var err = new Errors.Unauthorized(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 404:
              var err = new Errors.NotFound(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 500:
              var err = new Errors.InternalError(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            default:
              var err = new Error("Received unexpected statusCode " + response.statusCode);
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
          }
        });
      }());
    });
  }

  /**
   * Deletes a cluster
   * @param {Object} params
   * @param {string} params.groupID
   * @param {string} params.clusterName
   * @param {object} [options]
   * @param {number} [options.timeout] - A request specific timeout
   * @param {external:Span} [options.span] - An OpenTracing span - For example from the parent request
   * @param {module:atlas-api-client.RetryPolicies} [options.retryPolicy] - A request specific retryPolicy
   * @param {function} [cb]
   * @returns {Promise}
   * @fulfill {undefined}
   * @reject {module:atlas-api-client.Errors.BadRequest}
   * @reject {module:atlas-api-client.Errors.Unauthorized}
   * @reject {module:atlas-api-client.Errors.NotFound}
   * @reject {module:atlas-api-client.Errors.InternalError}
   * @reject {Error}
   */
  deleteCluster(params, options, cb) {
    return this._hystrixCommand.execute(this._deleteCluster, arguments);
  }
  _deleteCluster(params, options, cb) {
    if (!cb && typeof options === "function") {
      cb = options;
      options = undefined;
    }

    return new Promise((resolve, reject) => {
      const rejecter = (err) => {
        reject(err);
        if (cb) {
          cb(err);
        }
      };
      const resolver = (data) => {
        resolve(data);
        if (cb) {
          cb(null, data);
        }
      };


      if (!options) {
        options = {};
      }

      const timeout = options.timeout || this.timeout;
      const tracer = options.tracer || this.tracer;
      const span = options.span;

      const headers = {};
      if (!params.groupID) {
        rejecter(new Error("groupID must be non-empty because it's a path parameter"));
        return;
      }
      if (!params.clusterName) {
        rejecter(new Error("clusterName must be non-empty because it's a path parameter"));
        return;
      }

      const query = {};

      if (span) {
        // Need to get tracer to inject. Use HTTP headers format so we can properly escape special characters
        tracer.inject(span, opentracing.FORMAT_HTTP_HEADERS, headers);
        span.logEvent("DELETE /api/atlas/v1.0/groups/{groupID}/clusters/{clusterName}");
        span.setTag("span.kind", "client");
      }

	  const requestOptions = {
        method: "DELETE",
        uri: this.address + "/api/atlas/v1.0/groups/" + params.groupID + "/clusters/" + params.clusterName + "",
        json: true,
        timeout,
        headers,
        qs: query,
        useQuerystring: true,
      };
      if (this.keepalive) {
        requestOptions.forever = true;
      }
  

      const retryPolicy = options.retryPolicy || this.retryPolicy || singleRetryPolicy;
      const backoffs = retryPolicy.backoffs();
      const logger = this.logger;
  
      let retries = 0;
      (function requestOnce() {
        request(requestOptions, (err, response, body) => {
          if (retries < backoffs.length && retryPolicy.retry(requestOptions, err, response, body)) {
            const backoff = backoffs[retries];
            retries += 1;
            setTimeout(requestOnce, backoff);
            return;
          }
          if (err) {
            err._fromRequest = true;
            responseLog(logger, requestOptions, response, err)
            rejecter(err);
            return;
          }

          switch (response.statusCode) {
            case 202:
              resolver();
              break;
            
            case 400:
              var err = new Errors.BadRequest(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 401:
              var err = new Errors.Unauthorized(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 404:
              var err = new Errors.NotFound(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 500:
              var err = new Errors.InternalError(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            default:
              var err = new Error("Received unexpected statusCode " + response.statusCode);
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
          }
        });
      }());
    });
  }

  /**
   * Gets a cluster
   * @param {Object} params
   * @param {string} params.groupID
   * @param {string} params.clusterName
   * @param {object} [options]
   * @param {number} [options.timeout] - A request specific timeout
   * @param {external:Span} [options.span] - An OpenTracing span - For example from the parent request
   * @param {module:atlas-api-client.RetryPolicies} [options.retryPolicy] - A request specific retryPolicy
   * @param {function} [cb]
   * @returns {Promise}
   * @fulfill {Object}
   * @reject {module:atlas-api-client.Errors.BadRequest}
   * @reject {module:atlas-api-client.Errors.NotFound}
   * @reject {module:atlas-api-client.Errors.InternalError}
   * @reject {Error}
   */
  getCluster(params, options, cb) {
    return this._hystrixCommand.execute(this._getCluster, arguments);
  }
  _getCluster(params, options, cb) {
    if (!cb && typeof options === "function") {
      cb = options;
      options = undefined;
    }

    return new Promise((resolve, reject) => {
      const rejecter = (err) => {
        reject(err);
        if (cb) {
          cb(err);
        }
      };
      const resolver = (data) => {
        resolve(data);
        if (cb) {
          cb(null, data);
        }
      };


      if (!options) {
        options = {};
      }

      const timeout = options.timeout || this.timeout;
      const tracer = options.tracer || this.tracer;
      const span = options.span;

      const headers = {};
      if (!params.groupID) {
        rejecter(new Error("groupID must be non-empty because it's a path parameter"));
        return;
      }
      if (!params.clusterName) {
        rejecter(new Error("clusterName must be non-empty because it's a path parameter"));
        return;
      }

      const query = {};

      if (span) {
        // Need to get tracer to inject. Use HTTP headers format so we can properly escape special characters
        tracer.inject(span, opentracing.FORMAT_HTTP_HEADERS, headers);
        span.logEvent("GET /api/atlas/v1.0/groups/{groupID}/clusters/{clusterName}");
        span.setTag("span.kind", "client");
      }

	  const requestOptions = {
        method: "GET",
        uri: this.address + "/api/atlas/v1.0/groups/" + params.groupID + "/clusters/" + params.clusterName + "",
        json: true,
        timeout,
        headers,
        qs: query,
        useQuerystring: true,
      };
      if (this.keepalive) {
        requestOptions.forever = true;
      }
  

      const retryPolicy = options.retryPolicy || this.retryPolicy || singleRetryPolicy;
      const backoffs = retryPolicy.backoffs();
      const logger = this.logger;
  
      let retries = 0;
      (function requestOnce() {
        request(requestOptions, (err, response, body) => {
          if (retries < backoffs.length && retryPolicy.retry(requestOptions, err, response, body)) {
            const backoff = backoffs[retries];
            retries += 1;
            setTimeout(requestOnce, backoff);
            return;
          }
          if (err) {
            err._fromRequest = true;
            responseLog(logger, requestOptions, response, err)
            rejecter(err);
            return;
          }

          switch (response.statusCode) {
            case 200:
              resolver(body);
              break;
            
            case 400:
              var err = new Errors.BadRequest(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 404:
              var err = new Errors.NotFound(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 500:
              var err = new Errors.InternalError(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            default:
              var err = new Error("Received unexpected statusCode " + response.statusCode);
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
          }
        });
      }());
    });
  }

  /**
   * Update a Cluster
   * @param {Object} params
   * @param {string} params.groupID
   * @param {string} params.clusterName
   * @param params.createOrUpdateClusterRequest
   * @param {object} [options]
   * @param {number} [options.timeout] - A request specific timeout
   * @param {external:Span} [options.span] - An OpenTracing span - For example from the parent request
   * @param {module:atlas-api-client.RetryPolicies} [options.retryPolicy] - A request specific retryPolicy
   * @param {function} [cb]
   * @returns {Promise}
   * @fulfill {Object}
   * @reject {module:atlas-api-client.Errors.BadRequest}
   * @reject {module:atlas-api-client.Errors.Unauthorized}
   * @reject {module:atlas-api-client.Errors.NotFound}
   * @reject {module:atlas-api-client.Errors.InternalError}
   * @reject {Error}
   */
  updateCluster(params, options, cb) {
    return this._hystrixCommand.execute(this._updateCluster, arguments);
  }
  _updateCluster(params, options, cb) {
    if (!cb && typeof options === "function") {
      cb = options;
      options = undefined;
    }

    return new Promise((resolve, reject) => {
      const rejecter = (err) => {
        reject(err);
        if (cb) {
          cb(err);
        }
      };
      const resolver = (data) => {
        resolve(data);
        if (cb) {
          cb(null, data);
        }
      };


      if (!options) {
        options = {};
      }

      const timeout = options.timeout || this.timeout;
      const tracer = options.tracer || this.tracer;
      const span = options.span;

      const headers = {};
      if (!params.groupID) {
        rejecter(new Error("groupID must be non-empty because it's a path parameter"));
        return;
      }
      if (!params.clusterName) {
        rejecter(new Error("clusterName must be non-empty because it's a path parameter"));
        return;
      }

      const query = {};

      if (span) {
        // Need to get tracer to inject. Use HTTP headers format so we can properly escape special characters
        tracer.inject(span, opentracing.FORMAT_HTTP_HEADERS, headers);
        span.logEvent("PATCH /api/atlas/v1.0/groups/{groupID}/clusters/{clusterName}");
        span.setTag("span.kind", "client");
      }

	  const requestOptions = {
        method: "PATCH",
        uri: this.address + "/api/atlas/v1.0/groups/" + params.groupID + "/clusters/" + params.clusterName + "",
        json: true,
        timeout,
        headers,
        qs: query,
        useQuerystring: true,
      };
      if (this.keepalive) {
        requestOptions.forever = true;
      }
  
      requestOptions.body = params.createOrUpdateClusterRequest;
  

      const retryPolicy = options.retryPolicy || this.retryPolicy || singleRetryPolicy;
      const backoffs = retryPolicy.backoffs();
      const logger = this.logger;
  
      let retries = 0;
      (function requestOnce() {
        request(requestOptions, (err, response, body) => {
          if (retries < backoffs.length && retryPolicy.retry(requestOptions, err, response, body)) {
            const backoff = backoffs[retries];
            retries += 1;
            setTimeout(requestOnce, backoff);
            return;
          }
          if (err) {
            err._fromRequest = true;
            responseLog(logger, requestOptions, response, err)
            rejecter(err);
            return;
          }

          switch (response.statusCode) {
            case 200:
              resolver(body);
              break;
            
            case 400:
              var err = new Errors.BadRequest(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 401:
              var err = new Errors.Unauthorized(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 404:
              var err = new Errors.NotFound(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 500:
              var err = new Errors.InternalError(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            default:
              var err = new Error("Received unexpected statusCode " + response.statusCode);
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
          }
        });
      }());
    });
  }

  /**
   * Get All Containers
   * @param {string} groupID
   * @param {object} [options]
   * @param {number} [options.timeout] - A request specific timeout
   * @param {external:Span} [options.span] - An OpenTracing span - For example from the parent request
   * @param {module:atlas-api-client.RetryPolicies} [options.retryPolicy] - A request specific retryPolicy
   * @param {function} [cb]
   * @returns {Promise}
   * @fulfill {Object}
   * @reject {module:atlas-api-client.Errors.BadRequest}
   * @reject {module:atlas-api-client.Errors.Unauthorized}
   * @reject {module:atlas-api-client.Errors.NotFound}
   * @reject {module:atlas-api-client.Errors.InternalError}
   * @reject {Error}
   */
  getContainers(groupID, options, cb) {
    return this._hystrixCommand.execute(this._getContainers, arguments);
  }
  _getContainers(groupID, options, cb) {
    const params = {};
    params["groupID"] = groupID;

    if (!cb && typeof options === "function") {
      cb = options;
      options = undefined;
    }

    return new Promise((resolve, reject) => {
      const rejecter = (err) => {
        reject(err);
        if (cb) {
          cb(err);
        }
      };
      const resolver = (data) => {
        resolve(data);
        if (cb) {
          cb(null, data);
        }
      };


      if (!options) {
        options = {};
      }

      const timeout = options.timeout || this.timeout;
      const tracer = options.tracer || this.tracer;
      const span = options.span;

      const headers = {};
      if (!params.groupID) {
        rejecter(new Error("groupID must be non-empty because it's a path parameter"));
        return;
      }

      const query = {};

      if (span) {
        // Need to get tracer to inject. Use HTTP headers format so we can properly escape special characters
        tracer.inject(span, opentracing.FORMAT_HTTP_HEADERS, headers);
        span.logEvent("GET /api/atlas/v1.0/groups/{groupID}/containers");
        span.setTag("span.kind", "client");
      }

	  const requestOptions = {
        method: "GET",
        uri: this.address + "/api/atlas/v1.0/groups/" + params.groupID + "/containers",
        json: true,
        timeout,
        headers,
        qs: query,
        useQuerystring: true,
      };
      if (this.keepalive) {
        requestOptions.forever = true;
      }
  

      const retryPolicy = options.retryPolicy || this.retryPolicy || singleRetryPolicy;
      const backoffs = retryPolicy.backoffs();
      const logger = this.logger;
  
      let retries = 0;
      (function requestOnce() {
        request(requestOptions, (err, response, body) => {
          if (retries < backoffs.length && retryPolicy.retry(requestOptions, err, response, body)) {
            const backoff = backoffs[retries];
            retries += 1;
            setTimeout(requestOnce, backoff);
            return;
          }
          if (err) {
            err._fromRequest = true;
            responseLog(logger, requestOptions, response, err)
            rejecter(err);
            return;
          }

          switch (response.statusCode) {
            case 200:
              resolver(body);
              break;
            
            case 400:
              var err = new Errors.BadRequest(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 401:
              var err = new Errors.Unauthorized(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 404:
              var err = new Errors.NotFound(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 500:
              var err = new Errors.InternalError(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            default:
              var err = new Error("Received unexpected statusCode " + response.statusCode);
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
          }
        });
      }());
    });
  }

  /**
   * Create a Container
   * @param {Object} params
   * @param {string} params.groupID
   * @param params.createOrUpdateContainerRequest
   * @param {object} [options]
   * @param {number} [options.timeout] - A request specific timeout
   * @param {external:Span} [options.span] - An OpenTracing span - For example from the parent request
   * @param {module:atlas-api-client.RetryPolicies} [options.retryPolicy] - A request specific retryPolicy
   * @param {function} [cb]
   * @returns {Promise}
   * @fulfill {Object}
   * @reject {module:atlas-api-client.Errors.BadRequest}
   * @reject {module:atlas-api-client.Errors.Unauthorized}
   * @reject {module:atlas-api-client.Errors.NotFound}
   * @reject {module:atlas-api-client.Errors.InternalError}
   * @reject {Error}
   */
  createContainer(params, options, cb) {
    return this._hystrixCommand.execute(this._createContainer, arguments);
  }
  _createContainer(params, options, cb) {
    if (!cb && typeof options === "function") {
      cb = options;
      options = undefined;
    }

    return new Promise((resolve, reject) => {
      const rejecter = (err) => {
        reject(err);
        if (cb) {
          cb(err);
        }
      };
      const resolver = (data) => {
        resolve(data);
        if (cb) {
          cb(null, data);
        }
      };


      if (!options) {
        options = {};
      }

      const timeout = options.timeout || this.timeout;
      const tracer = options.tracer || this.tracer;
      const span = options.span;

      const headers = {};
      if (!params.groupID) {
        rejecter(new Error("groupID must be non-empty because it's a path parameter"));
        return;
      }

      const query = {};

      if (span) {
        // Need to get tracer to inject. Use HTTP headers format so we can properly escape special characters
        tracer.inject(span, opentracing.FORMAT_HTTP_HEADERS, headers);
        span.logEvent("POST /api/atlas/v1.0/groups/{groupID}/containers");
        span.setTag("span.kind", "client");
      }

	  const requestOptions = {
        method: "POST",
        uri: this.address + "/api/atlas/v1.0/groups/" + params.groupID + "/containers",
        json: true,
        timeout,
        headers,
        qs: query,
        useQuerystring: true,
      };
      if (this.keepalive) {
        requestOptions.forever = true;
      }
  
      requestOptions.body = params.createOrUpdateContainerRequest;
  

      const retryPolicy = options.retryPolicy || this.retryPolicy || singleRetryPolicy;
      const backoffs = retryPolicy.backoffs();
      const logger = this.logger;
  
      let retries = 0;
      (function requestOnce() {
        request(requestOptions, (err, response, body) => {
          if (retries < backoffs.length && retryPolicy.retry(requestOptions, err, response, body)) {
            const backoff = backoffs[retries];
            retries += 1;
            setTimeout(requestOnce, backoff);
            return;
          }
          if (err) {
            err._fromRequest = true;
            responseLog(logger, requestOptions, response, err)
            rejecter(err);
            return;
          }

          switch (response.statusCode) {
            case 201:
              resolver(body);
              break;
            
            case 400:
              var err = new Errors.BadRequest(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 401:
              var err = new Errors.Unauthorized(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 404:
              var err = new Errors.NotFound(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 500:
              var err = new Errors.InternalError(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            default:
              var err = new Error("Received unexpected statusCode " + response.statusCode);
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
          }
        });
      }());
    });
  }

  /**
   * Gets a container
   * @param {Object} params
   * @param {string} params.groupID
   * @param {string} params.containerID
   * @param {object} [options]
   * @param {number} [options.timeout] - A request specific timeout
   * @param {external:Span} [options.span] - An OpenTracing span - For example from the parent request
   * @param {module:atlas-api-client.RetryPolicies} [options.retryPolicy] - A request specific retryPolicy
   * @param {function} [cb]
   * @returns {Promise}
   * @fulfill {Object}
   * @reject {module:atlas-api-client.Errors.BadRequest}
   * @reject {module:atlas-api-client.Errors.NotFound}
   * @reject {module:atlas-api-client.Errors.InternalError}
   * @reject {Error}
   */
  getContainer(params, options, cb) {
    return this._hystrixCommand.execute(this._getContainer, arguments);
  }
  _getContainer(params, options, cb) {
    if (!cb && typeof options === "function") {
      cb = options;
      options = undefined;
    }

    return new Promise((resolve, reject) => {
      const rejecter = (err) => {
        reject(err);
        if (cb) {
          cb(err);
        }
      };
      const resolver = (data) => {
        resolve(data);
        if (cb) {
          cb(null, data);
        }
      };


      if (!options) {
        options = {};
      }

      const timeout = options.timeout || this.timeout;
      const tracer = options.tracer || this.tracer;
      const span = options.span;

      const headers = {};
      if (!params.groupID) {
        rejecter(new Error("groupID must be non-empty because it's a path parameter"));
        return;
      }
      if (!params.containerID) {
        rejecter(new Error("containerID must be non-empty because it's a path parameter"));
        return;
      }

      const query = {};

      if (span) {
        // Need to get tracer to inject. Use HTTP headers format so we can properly escape special characters
        tracer.inject(span, opentracing.FORMAT_HTTP_HEADERS, headers);
        span.logEvent("GET /api/atlas/v1.0/groups/{groupID}/containers/{containerID}");
        span.setTag("span.kind", "client");
      }

	  const requestOptions = {
        method: "GET",
        uri: this.address + "/api/atlas/v1.0/groups/" + params.groupID + "/containers/" + params.containerID + "",
        json: true,
        timeout,
        headers,
        qs: query,
        useQuerystring: true,
      };
      if (this.keepalive) {
        requestOptions.forever = true;
      }
  

      const retryPolicy = options.retryPolicy || this.retryPolicy || singleRetryPolicy;
      const backoffs = retryPolicy.backoffs();
      const logger = this.logger;
  
      let retries = 0;
      (function requestOnce() {
        request(requestOptions, (err, response, body) => {
          if (retries < backoffs.length && retryPolicy.retry(requestOptions, err, response, body)) {
            const backoff = backoffs[retries];
            retries += 1;
            setTimeout(requestOnce, backoff);
            return;
          }
          if (err) {
            err._fromRequest = true;
            responseLog(logger, requestOptions, response, err)
            rejecter(err);
            return;
          }

          switch (response.statusCode) {
            case 200:
              resolver(body);
              break;
            
            case 400:
              var err = new Errors.BadRequest(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 404:
              var err = new Errors.NotFound(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 500:
              var err = new Errors.InternalError(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            default:
              var err = new Error("Received unexpected statusCode " + response.statusCode);
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
          }
        });
      }());
    });
  }

  /**
   * Update a Container
   * @param {Object} params
   * @param {string} params.groupID
   * @param {string} params.containerID
   * @param params.createOrUpdateContainerRequest
   * @param {object} [options]
   * @param {number} [options.timeout] - A request specific timeout
   * @param {external:Span} [options.span] - An OpenTracing span - For example from the parent request
   * @param {module:atlas-api-client.RetryPolicies} [options.retryPolicy] - A request specific retryPolicy
   * @param {function} [cb]
   * @returns {Promise}
   * @fulfill {Object}
   * @reject {module:atlas-api-client.Errors.BadRequest}
   * @reject {module:atlas-api-client.Errors.Unauthorized}
   * @reject {module:atlas-api-client.Errors.NotFound}
   * @reject {module:atlas-api-client.Errors.InternalError}
   * @reject {Error}
   */
  updateContainer(params, options, cb) {
    return this._hystrixCommand.execute(this._updateContainer, arguments);
  }
  _updateContainer(params, options, cb) {
    if (!cb && typeof options === "function") {
      cb = options;
      options = undefined;
    }

    return new Promise((resolve, reject) => {
      const rejecter = (err) => {
        reject(err);
        if (cb) {
          cb(err);
        }
      };
      const resolver = (data) => {
        resolve(data);
        if (cb) {
          cb(null, data);
        }
      };


      if (!options) {
        options = {};
      }

      const timeout = options.timeout || this.timeout;
      const tracer = options.tracer || this.tracer;
      const span = options.span;

      const headers = {};
      if (!params.groupID) {
        rejecter(new Error("groupID must be non-empty because it's a path parameter"));
        return;
      }
      if (!params.containerID) {
        rejecter(new Error("containerID must be non-empty because it's a path parameter"));
        return;
      }

      const query = {};

      if (span) {
        // Need to get tracer to inject. Use HTTP headers format so we can properly escape special characters
        tracer.inject(span, opentracing.FORMAT_HTTP_HEADERS, headers);
        span.logEvent("PATCH /api/atlas/v1.0/groups/{groupID}/containers/{containerID}");
        span.setTag("span.kind", "client");
      }

	  const requestOptions = {
        method: "PATCH",
        uri: this.address + "/api/atlas/v1.0/groups/" + params.groupID + "/containers/" + params.containerID + "",
        json: true,
        timeout,
        headers,
        qs: query,
        useQuerystring: true,
      };
      if (this.keepalive) {
        requestOptions.forever = true;
      }
  
      requestOptions.body = params.createOrUpdateContainerRequest;
  

      const retryPolicy = options.retryPolicy || this.retryPolicy || singleRetryPolicy;
      const backoffs = retryPolicy.backoffs();
      const logger = this.logger;
  
      let retries = 0;
      (function requestOnce() {
        request(requestOptions, (err, response, body) => {
          if (retries < backoffs.length && retryPolicy.retry(requestOptions, err, response, body)) {
            const backoff = backoffs[retries];
            retries += 1;
            setTimeout(requestOnce, backoff);
            return;
          }
          if (err) {
            err._fromRequest = true;
            responseLog(logger, requestOptions, response, err)
            rejecter(err);
            return;
          }

          switch (response.statusCode) {
            case 200:
              resolver(body);
              break;
            
            case 400:
              var err = new Errors.BadRequest(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 401:
              var err = new Errors.Unauthorized(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 404:
              var err = new Errors.NotFound(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 500:
              var err = new Errors.InternalError(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            default:
              var err = new Error("Received unexpected statusCode " + response.statusCode);
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
          }
        });
      }());
    });
  }

  /**
   * Get All DatabaseUsers
   * @param {string} groupID
   * @param {object} [options]
   * @param {number} [options.timeout] - A request specific timeout
   * @param {external:Span} [options.span] - An OpenTracing span - For example from the parent request
   * @param {module:atlas-api-client.RetryPolicies} [options.retryPolicy] - A request specific retryPolicy
   * @param {function} [cb]
   * @returns {Promise}
   * @fulfill {Object}
   * @reject {module:atlas-api-client.Errors.BadRequest}
   * @reject {module:atlas-api-client.Errors.Unauthorized}
   * @reject {module:atlas-api-client.Errors.NotFound}
   * @reject {module:atlas-api-client.Errors.InternalError}
   * @reject {Error}
   */
  getDatabaseUsers(groupID, options, cb) {
    return this._hystrixCommand.execute(this._getDatabaseUsers, arguments);
  }
  _getDatabaseUsers(groupID, options, cb) {
    const params = {};
    params["groupID"] = groupID;

    if (!cb && typeof options === "function") {
      cb = options;
      options = undefined;
    }

    return new Promise((resolve, reject) => {
      const rejecter = (err) => {
        reject(err);
        if (cb) {
          cb(err);
        }
      };
      const resolver = (data) => {
        resolve(data);
        if (cb) {
          cb(null, data);
        }
      };


      if (!options) {
        options = {};
      }

      const timeout = options.timeout || this.timeout;
      const tracer = options.tracer || this.tracer;
      const span = options.span;

      const headers = {};
      if (!params.groupID) {
        rejecter(new Error("groupID must be non-empty because it's a path parameter"));
        return;
      }

      const query = {};

      if (span) {
        // Need to get tracer to inject. Use HTTP headers format so we can properly escape special characters
        tracer.inject(span, opentracing.FORMAT_HTTP_HEADERS, headers);
        span.logEvent("GET /api/atlas/v1.0/groups/{groupID}/databaseUsers");
        span.setTag("span.kind", "client");
      }

	  const requestOptions = {
        method: "GET",
        uri: this.address + "/api/atlas/v1.0/groups/" + params.groupID + "/databaseUsers",
        json: true,
        timeout,
        headers,
        qs: query,
        useQuerystring: true,
      };
      if (this.keepalive) {
        requestOptions.forever = true;
      }
  

      const retryPolicy = options.retryPolicy || this.retryPolicy || singleRetryPolicy;
      const backoffs = retryPolicy.backoffs();
      const logger = this.logger;
  
      let retries = 0;
      (function requestOnce() {
        request(requestOptions, (err, response, body) => {
          if (retries < backoffs.length && retryPolicy.retry(requestOptions, err, response, body)) {
            const backoff = backoffs[retries];
            retries += 1;
            setTimeout(requestOnce, backoff);
            return;
          }
          if (err) {
            err._fromRequest = true;
            responseLog(logger, requestOptions, response, err)
            rejecter(err);
            return;
          }

          switch (response.statusCode) {
            case 200:
              resolver(body);
              break;
            
            case 400:
              var err = new Errors.BadRequest(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 401:
              var err = new Errors.Unauthorized(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 404:
              var err = new Errors.NotFound(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 500:
              var err = new Errors.InternalError(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            default:
              var err = new Error("Received unexpected statusCode " + response.statusCode);
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
          }
        });
      }());
    });
  }

  /**
   * Create a DatabaseUser
   * @param {Object} params
   * @param {string} params.groupID
   * @param params.createDatabaseUserRequest
   * @param {object} [options]
   * @param {number} [options.timeout] - A request specific timeout
   * @param {external:Span} [options.span] - An OpenTracing span - For example from the parent request
   * @param {module:atlas-api-client.RetryPolicies} [options.retryPolicy] - A request specific retryPolicy
   * @param {function} [cb]
   * @returns {Promise}
   * @fulfill {Object}
   * @reject {module:atlas-api-client.Errors.BadRequest}
   * @reject {module:atlas-api-client.Errors.Unauthorized}
   * @reject {module:atlas-api-client.Errors.NotFound}
   * @reject {module:atlas-api-client.Errors.InternalError}
   * @reject {Error}
   */
  createDatabaseUser(params, options, cb) {
    return this._hystrixCommand.execute(this._createDatabaseUser, arguments);
  }
  _createDatabaseUser(params, options, cb) {
    if (!cb && typeof options === "function") {
      cb = options;
      options = undefined;
    }

    return new Promise((resolve, reject) => {
      const rejecter = (err) => {
        reject(err);
        if (cb) {
          cb(err);
        }
      };
      const resolver = (data) => {
        resolve(data);
        if (cb) {
          cb(null, data);
        }
      };


      if (!options) {
        options = {};
      }

      const timeout = options.timeout || this.timeout;
      const tracer = options.tracer || this.tracer;
      const span = options.span;

      const headers = {};
      if (!params.groupID) {
        rejecter(new Error("groupID must be non-empty because it's a path parameter"));
        return;
      }

      const query = {};

      if (span) {
        // Need to get tracer to inject. Use HTTP headers format so we can properly escape special characters
        tracer.inject(span, opentracing.FORMAT_HTTP_HEADERS, headers);
        span.logEvent("POST /api/atlas/v1.0/groups/{groupID}/databaseUsers");
        span.setTag("span.kind", "client");
      }

	  const requestOptions = {
        method: "POST",
        uri: this.address + "/api/atlas/v1.0/groups/" + params.groupID + "/databaseUsers",
        json: true,
        timeout,
        headers,
        qs: query,
        useQuerystring: true,
      };
      if (this.keepalive) {
        requestOptions.forever = true;
      }
  
      requestOptions.body = params.createDatabaseUserRequest;
  

      const retryPolicy = options.retryPolicy || this.retryPolicy || singleRetryPolicy;
      const backoffs = retryPolicy.backoffs();
      const logger = this.logger;
  
      let retries = 0;
      (function requestOnce() {
        request(requestOptions, (err, response, body) => {
          if (retries < backoffs.length && retryPolicy.retry(requestOptions, err, response, body)) {
            const backoff = backoffs[retries];
            retries += 1;
            setTimeout(requestOnce, backoff);
            return;
          }
          if (err) {
            err._fromRequest = true;
            responseLog(logger, requestOptions, response, err)
            rejecter(err);
            return;
          }

          switch (response.statusCode) {
            case 201:
              resolver(body);
              break;
            
            case 400:
              var err = new Errors.BadRequest(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 401:
              var err = new Errors.Unauthorized(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 404:
              var err = new Errors.NotFound(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 500:
              var err = new Errors.InternalError(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            default:
              var err = new Error("Received unexpected statusCode " + response.statusCode);
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
          }
        });
      }());
    });
  }

  /**
   * Deletes a DatabaseUser
   * @param {Object} params
   * @param {string} params.groupID
   * @param {string} params.username
   * @param {object} [options]
   * @param {number} [options.timeout] - A request specific timeout
   * @param {external:Span} [options.span] - An OpenTracing span - For example from the parent request
   * @param {module:atlas-api-client.RetryPolicies} [options.retryPolicy] - A request specific retryPolicy
   * @param {function} [cb]
   * @returns {Promise}
   * @fulfill {undefined}
   * @reject {module:atlas-api-client.Errors.BadRequest}
   * @reject {module:atlas-api-client.Errors.Unauthorized}
   * @reject {module:atlas-api-client.Errors.NotFound}
   * @reject {module:atlas-api-client.Errors.InternalError}
   * @reject {Error}
   */
  deleteDatabaseUser(params, options, cb) {
    return this._hystrixCommand.execute(this._deleteDatabaseUser, arguments);
  }
  _deleteDatabaseUser(params, options, cb) {
    if (!cb && typeof options === "function") {
      cb = options;
      options = undefined;
    }

    return new Promise((resolve, reject) => {
      const rejecter = (err) => {
        reject(err);
        if (cb) {
          cb(err);
        }
      };
      const resolver = (data) => {
        resolve(data);
        if (cb) {
          cb(null, data);
        }
      };


      if (!options) {
        options = {};
      }

      const timeout = options.timeout || this.timeout;
      const tracer = options.tracer || this.tracer;
      const span = options.span;

      const headers = {};
      if (!params.groupID) {
        rejecter(new Error("groupID must be non-empty because it's a path parameter"));
        return;
      }
      if (!params.username) {
        rejecter(new Error("username must be non-empty because it's a path parameter"));
        return;
      }

      const query = {};

      if (span) {
        // Need to get tracer to inject. Use HTTP headers format so we can properly escape special characters
        tracer.inject(span, opentracing.FORMAT_HTTP_HEADERS, headers);
        span.logEvent("DELETE /api/atlas/v1.0/groups/{groupID}/databaseUsers/admin/{username}");
        span.setTag("span.kind", "client");
      }

	  const requestOptions = {
        method: "DELETE",
        uri: this.address + "/api/atlas/v1.0/groups/" + params.groupID + "/databaseUsers/admin/" + params.username + "",
        json: true,
        timeout,
        headers,
        qs: query,
        useQuerystring: true,
      };
      if (this.keepalive) {
        requestOptions.forever = true;
      }
  

      const retryPolicy = options.retryPolicy || this.retryPolicy || singleRetryPolicy;
      const backoffs = retryPolicy.backoffs();
      const logger = this.logger;
  
      let retries = 0;
      (function requestOnce() {
        request(requestOptions, (err, response, body) => {
          if (retries < backoffs.length && retryPolicy.retry(requestOptions, err, response, body)) {
            const backoff = backoffs[retries];
            retries += 1;
            setTimeout(requestOnce, backoff);
            return;
          }
          if (err) {
            err._fromRequest = true;
            responseLog(logger, requestOptions, response, err)
            rejecter(err);
            return;
          }

          switch (response.statusCode) {
            case 200:
              resolver();
              break;
            
            case 400:
              var err = new Errors.BadRequest(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 401:
              var err = new Errors.Unauthorized(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 404:
              var err = new Errors.NotFound(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 500:
              var err = new Errors.InternalError(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            default:
              var err = new Error("Received unexpected statusCode " + response.statusCode);
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
          }
        });
      }());
    });
  }

  /**
   * Gets a database user
   * @param {Object} params
   * @param {string} params.groupID
   * @param {string} params.username
   * @param {object} [options]
   * @param {number} [options.timeout] - A request specific timeout
   * @param {external:Span} [options.span] - An OpenTracing span - For example from the parent request
   * @param {module:atlas-api-client.RetryPolicies} [options.retryPolicy] - A request specific retryPolicy
   * @param {function} [cb]
   * @returns {Promise}
   * @fulfill {Object}
   * @reject {module:atlas-api-client.Errors.BadRequest}
   * @reject {module:atlas-api-client.Errors.NotFound}
   * @reject {module:atlas-api-client.Errors.InternalError}
   * @reject {Error}
   */
  getDatabaseUser(params, options, cb) {
    return this._hystrixCommand.execute(this._getDatabaseUser, arguments);
  }
  _getDatabaseUser(params, options, cb) {
    if (!cb && typeof options === "function") {
      cb = options;
      options = undefined;
    }

    return new Promise((resolve, reject) => {
      const rejecter = (err) => {
        reject(err);
        if (cb) {
          cb(err);
        }
      };
      const resolver = (data) => {
        resolve(data);
        if (cb) {
          cb(null, data);
        }
      };


      if (!options) {
        options = {};
      }

      const timeout = options.timeout || this.timeout;
      const tracer = options.tracer || this.tracer;
      const span = options.span;

      const headers = {};
      if (!params.groupID) {
        rejecter(new Error("groupID must be non-empty because it's a path parameter"));
        return;
      }
      if (!params.username) {
        rejecter(new Error("username must be non-empty because it's a path parameter"));
        return;
      }

      const query = {};

      if (span) {
        // Need to get tracer to inject. Use HTTP headers format so we can properly escape special characters
        tracer.inject(span, opentracing.FORMAT_HTTP_HEADERS, headers);
        span.logEvent("GET /api/atlas/v1.0/groups/{groupID}/databaseUsers/admin/{username}");
        span.setTag("span.kind", "client");
      }

	  const requestOptions = {
        method: "GET",
        uri: this.address + "/api/atlas/v1.0/groups/" + params.groupID + "/databaseUsers/admin/" + params.username + "",
        json: true,
        timeout,
        headers,
        qs: query,
        useQuerystring: true,
      };
      if (this.keepalive) {
        requestOptions.forever = true;
      }
  

      const retryPolicy = options.retryPolicy || this.retryPolicy || singleRetryPolicy;
      const backoffs = retryPolicy.backoffs();
      const logger = this.logger;
  
      let retries = 0;
      (function requestOnce() {
        request(requestOptions, (err, response, body) => {
          if (retries < backoffs.length && retryPolicy.retry(requestOptions, err, response, body)) {
            const backoff = backoffs[retries];
            retries += 1;
            setTimeout(requestOnce, backoff);
            return;
          }
          if (err) {
            err._fromRequest = true;
            responseLog(logger, requestOptions, response, err)
            rejecter(err);
            return;
          }

          switch (response.statusCode) {
            case 200:
              resolver(body);
              break;
            
            case 400:
              var err = new Errors.BadRequest(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 404:
              var err = new Errors.NotFound(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 500:
              var err = new Errors.InternalError(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            default:
              var err = new Error("Received unexpected statusCode " + response.statusCode);
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
          }
        });
      }());
    });
  }

  /**
   * Update a DatabaseUser
   * @param {Object} params
   * @param {string} params.groupID
   * @param {string} params.username
   * @param params.updateDatabaseUserRequest
   * @param {object} [options]
   * @param {number} [options.timeout] - A request specific timeout
   * @param {external:Span} [options.span] - An OpenTracing span - For example from the parent request
   * @param {module:atlas-api-client.RetryPolicies} [options.retryPolicy] - A request specific retryPolicy
   * @param {function} [cb]
   * @returns {Promise}
   * @fulfill {Object}
   * @reject {module:atlas-api-client.Errors.BadRequest}
   * @reject {module:atlas-api-client.Errors.Unauthorized}
   * @reject {module:atlas-api-client.Errors.NotFound}
   * @reject {module:atlas-api-client.Errors.InternalError}
   * @reject {Error}
   */
  updateDatabaseUser(params, options, cb) {
    return this._hystrixCommand.execute(this._updateDatabaseUser, arguments);
  }
  _updateDatabaseUser(params, options, cb) {
    if (!cb && typeof options === "function") {
      cb = options;
      options = undefined;
    }

    return new Promise((resolve, reject) => {
      const rejecter = (err) => {
        reject(err);
        if (cb) {
          cb(err);
        }
      };
      const resolver = (data) => {
        resolve(data);
        if (cb) {
          cb(null, data);
        }
      };


      if (!options) {
        options = {};
      }

      const timeout = options.timeout || this.timeout;
      const tracer = options.tracer || this.tracer;
      const span = options.span;

      const headers = {};
      if (!params.groupID) {
        rejecter(new Error("groupID must be non-empty because it's a path parameter"));
        return;
      }
      if (!params.username) {
        rejecter(new Error("username must be non-empty because it's a path parameter"));
        return;
      }

      const query = {};

      if (span) {
        // Need to get tracer to inject. Use HTTP headers format so we can properly escape special characters
        tracer.inject(span, opentracing.FORMAT_HTTP_HEADERS, headers);
        span.logEvent("PATCH /api/atlas/v1.0/groups/{groupID}/databaseUsers/admin/{username}");
        span.setTag("span.kind", "client");
      }

	  const requestOptions = {
        method: "PATCH",
        uri: this.address + "/api/atlas/v1.0/groups/" + params.groupID + "/databaseUsers/admin/" + params.username + "",
        json: true,
        timeout,
        headers,
        qs: query,
        useQuerystring: true,
      };
      if (this.keepalive) {
        requestOptions.forever = true;
      }
  
      requestOptions.body = params.updateDatabaseUserRequest;
  

      const retryPolicy = options.retryPolicy || this.retryPolicy || singleRetryPolicy;
      const backoffs = retryPolicy.backoffs();
      const logger = this.logger;
  
      let retries = 0;
      (function requestOnce() {
        request(requestOptions, (err, response, body) => {
          if (retries < backoffs.length && retryPolicy.retry(requestOptions, err, response, body)) {
            const backoff = backoffs[retries];
            retries += 1;
            setTimeout(requestOnce, backoff);
            return;
          }
          if (err) {
            err._fromRequest = true;
            responseLog(logger, requestOptions, response, err)
            rejecter(err);
            return;
          }

          switch (response.statusCode) {
            case 200:
              resolver(body);
              break;
            
            case 400:
              var err = new Errors.BadRequest(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 401:
              var err = new Errors.Unauthorized(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 404:
              var err = new Errors.NotFound(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 500:
              var err = new Errors.InternalError(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            default:
              var err = new Error("Received unexpected statusCode " + response.statusCode);
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
          }
        });
      }());
    });
  }

  /**
   * Get All Processes
   * @param {string} groupID
   * @param {object} [options]
   * @param {number} [options.timeout] - A request specific timeout
   * @param {external:Span} [options.span] - An OpenTracing span - For example from the parent request
   * @param {module:atlas-api-client.RetryPolicies} [options.retryPolicy] - A request specific retryPolicy
   * @param {function} [cb]
   * @returns {Promise}
   * @fulfill {Object}
   * @reject {module:atlas-api-client.Errors.BadRequest}
   * @reject {module:atlas-api-client.Errors.Unauthorized}
   * @reject {module:atlas-api-client.Errors.NotFound}
   * @reject {module:atlas-api-client.Errors.InternalError}
   * @reject {Error}
   */
  getProcesses(groupID, options, cb) {
    return this._hystrixCommand.execute(this._getProcesses, arguments);
  }
  _getProcesses(groupID, options, cb) {
    const params = {};
    params["groupID"] = groupID;

    if (!cb && typeof options === "function") {
      cb = options;
      options = undefined;
    }

    return new Promise((resolve, reject) => {
      const rejecter = (err) => {
        reject(err);
        if (cb) {
          cb(err);
        }
      };
      const resolver = (data) => {
        resolve(data);
        if (cb) {
          cb(null, data);
        }
      };


      if (!options) {
        options = {};
      }

      const timeout = options.timeout || this.timeout;
      const tracer = options.tracer || this.tracer;
      const span = options.span;

      const headers = {};
      if (!params.groupID) {
        rejecter(new Error("groupID must be non-empty because it's a path parameter"));
        return;
      }

      const query = {};

      if (span) {
        // Need to get tracer to inject. Use HTTP headers format so we can properly escape special characters
        tracer.inject(span, opentracing.FORMAT_HTTP_HEADERS, headers);
        span.logEvent("GET /api/atlas/v1.0/groups/{groupID}/processes");
        span.setTag("span.kind", "client");
      }

	  const requestOptions = {
        method: "GET",
        uri: this.address + "/api/atlas/v1.0/groups/" + params.groupID + "/processes",
        json: true,
        timeout,
        headers,
        qs: query,
        useQuerystring: true,
      };
      if (this.keepalive) {
        requestOptions.forever = true;
      }
  

      const retryPolicy = options.retryPolicy || this.retryPolicy || singleRetryPolicy;
      const backoffs = retryPolicy.backoffs();
      const logger = this.logger;
  
      let retries = 0;
      (function requestOnce() {
        request(requestOptions, (err, response, body) => {
          if (retries < backoffs.length && retryPolicy.retry(requestOptions, err, response, body)) {
            const backoff = backoffs[retries];
            retries += 1;
            setTimeout(requestOnce, backoff);
            return;
          }
          if (err) {
            err._fromRequest = true;
            responseLog(logger, requestOptions, response, err)
            rejecter(err);
            return;
          }

          switch (response.statusCode) {
            case 200:
              resolver(body);
              break;
            
            case 400:
              var err = new Errors.BadRequest(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 401:
              var err = new Errors.Unauthorized(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 404:
              var err = new Errors.NotFound(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 500:
              var err = new Errors.InternalError(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            default:
              var err = new Error("Received unexpected statusCode " + response.statusCode);
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
          }
        });
      }());
    });
  }

  /**
   * Get the available databases for a Atlas MongoDB Process
   * @param {Object} params
   * @param {string} params.groupID
   * @param {string} params.host
   * @param {number} params.port
   * @param {number} [params.pageNum]
   * @param {number} [params.itemsPerPage]
   * @param {object} [options]
   * @param {number} [options.timeout] - A request specific timeout
   * @param {external:Span} [options.span] - An OpenTracing span - For example from the parent request
   * @param {module:atlas-api-client.RetryPolicies} [options.retryPolicy] - A request specific retryPolicy
   * @param {function} [cb]
   * @returns {Promise}
   * @fulfill {Object}
   * @reject {module:atlas-api-client.Errors.BadRequest}
   * @reject {module:atlas-api-client.Errors.Unauthorized}
   * @reject {module:atlas-api-client.Errors.NotFound}
   * @reject {module:atlas-api-client.Errors.InternalError}
   * @reject {Error}
   */
  getProcessDatabases(params, options, cb) {
    return this._hystrixCommand.execute(this._getProcessDatabases, arguments);
  }
  _getProcessDatabases(params, options, cb) {
    if (!cb && typeof options === "function") {
      cb = options;
      options = undefined;
    }

    return new Promise((resolve, reject) => {
      const rejecter = (err) => {
        reject(err);
        if (cb) {
          cb(err);
        }
      };
      const resolver = (data) => {
        resolve(data);
        if (cb) {
          cb(null, data);
        }
      };


      if (!options) {
        options = {};
      }

      const timeout = options.timeout || this.timeout;
      const tracer = options.tracer || this.tracer;
      const span = options.span;

      const headers = {};
      if (!params.groupID) {
        rejecter(new Error("groupID must be non-empty because it's a path parameter"));
        return;
      }
      if (!params.host) {
        rejecter(new Error("host must be non-empty because it's a path parameter"));
        return;
      }
      if (!params.port) {
        rejecter(new Error("port must be non-empty because it's a path parameter"));
        return;
      }

      const query = {};
      if (typeof params.pageNum !== "undefined") {
        query["pageNum"] = params.pageNum;
      }
  
      if (typeof params.itemsPerPage !== "undefined") {
        query["itemsPerPage"] = params.itemsPerPage;
      }
  

      if (span) {
        // Need to get tracer to inject. Use HTTP headers format so we can properly escape special characters
        tracer.inject(span, opentracing.FORMAT_HTTP_HEADERS, headers);
        span.logEvent("GET /api/atlas/v1.0/groups/{groupID}/processes/{host}:{port}/databases");
        span.setTag("span.kind", "client");
      }

	  const requestOptions = {
        method: "GET",
        uri: this.address + "/api/atlas/v1.0/groups/" + params.groupID + "/processes/" + params.host + ":" + params.port + "/databases",
        json: true,
        timeout,
        headers,
        qs: query,
        useQuerystring: true,
      };
      if (this.keepalive) {
        requestOptions.forever = true;
      }
  

      const retryPolicy = options.retryPolicy || this.retryPolicy || singleRetryPolicy;
      const backoffs = retryPolicy.backoffs();
      const logger = this.logger;
  
      let retries = 0;
      (function requestOnce() {
        request(requestOptions, (err, response, body) => {
          if (retries < backoffs.length && retryPolicy.retry(requestOptions, err, response, body)) {
            const backoff = backoffs[retries];
            retries += 1;
            setTimeout(requestOnce, backoff);
            return;
          }
          if (err) {
            err._fromRequest = true;
            responseLog(logger, requestOptions, response, err)
            rejecter(err);
            return;
          }

          switch (response.statusCode) {
            case 200:
              resolver(body);
              break;
            
            case 400:
              var err = new Errors.BadRequest(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 401:
              var err = new Errors.Unauthorized(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 404:
              var err = new Errors.NotFound(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 500:
              var err = new Errors.InternalError(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            default:
              var err = new Error("Received unexpected statusCode " + response.statusCode);
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
          }
        });
      }());
    });
  }

  /**
   * Get the measurements of the specified database for a Atlas MongoDB process.
   * @param {Object} params
   * @param {string} params.groupID
   * @param {string} params.host
   * @param {number} params.port
   * @param {string} params.databaseID
   * @param {string} params.granularity
   * @param {string} [params.period]
   * @param {string} [params.start]
   * @param {string} [params.end]
   * @param {string[]} [params.m]
   * @param {number} [params.pageNum]
   * @param {number} [params.itemsPerPage]
   * @param {object} [options]
   * @param {number} [options.timeout] - A request specific timeout
   * @param {external:Span} [options.span] - An OpenTracing span - For example from the parent request
   * @param {module:atlas-api-client.RetryPolicies} [options.retryPolicy] - A request specific retryPolicy
   * @param {function} [cb]
   * @returns {Promise}
   * @fulfill {Object}
   * @reject {module:atlas-api-client.Errors.BadRequest}
   * @reject {module:atlas-api-client.Errors.Unauthorized}
   * @reject {module:atlas-api-client.Errors.NotFound}
   * @reject {module:atlas-api-client.Errors.InternalError}
   * @reject {Error}
   */
  getProcessDatabaseMeasurements(params, options, cb) {
    return this._hystrixCommand.execute(this._getProcessDatabaseMeasurements, arguments);
  }
  _getProcessDatabaseMeasurements(params, options, cb) {
    if (!cb && typeof options === "function") {
      cb = options;
      options = undefined;
    }

    return new Promise((resolve, reject) => {
      const rejecter = (err) => {
        reject(err);
        if (cb) {
          cb(err);
        }
      };
      const resolver = (data) => {
        resolve(data);
        if (cb) {
          cb(null, data);
        }
      };


      if (!options) {
        options = {};
      }

      const timeout = options.timeout || this.timeout;
      const tracer = options.tracer || this.tracer;
      const span = options.span;

      const headers = {};
      if (!params.groupID) {
        rejecter(new Error("groupID must be non-empty because it's a path parameter"));
        return;
      }
      if (!params.host) {
        rejecter(new Error("host must be non-empty because it's a path parameter"));
        return;
      }
      if (!params.port) {
        rejecter(new Error("port must be non-empty because it's a path parameter"));
        return;
      }
      if (!params.databaseID) {
        rejecter(new Error("databaseID must be non-empty because it's a path parameter"));
        return;
      }

      const query = {};
      query["granularity"] = params.granularity;
  
      if (typeof params.period !== "undefined") {
        query["period"] = params.period;
      }
  
      if (typeof params.start !== "undefined") {
        query["start"] = params.start;
      }
  
      if (typeof params.end !== "undefined") {
        query["end"] = params.end;
      }
  
      if (typeof params.m !== "undefined") {
        query["m"] = params.m;
      }
  
      if (typeof params.pageNum !== "undefined") {
        query["pageNum"] = params.pageNum;
      }
  
      if (typeof params.itemsPerPage !== "undefined") {
        query["itemsPerPage"] = params.itemsPerPage;
      }
  

      if (span) {
        // Need to get tracer to inject. Use HTTP headers format so we can properly escape special characters
        tracer.inject(span, opentracing.FORMAT_HTTP_HEADERS, headers);
        span.logEvent("GET /api/atlas/v1.0/groups/{groupID}/processes/{host}:{port}/databases/{databaseID}/measurements");
        span.setTag("span.kind", "client");
      }

	  const requestOptions = {
        method: "GET",
        uri: this.address + "/api/atlas/v1.0/groups/" + params.groupID + "/processes/" + params.host + ":" + params.port + "/databases/" + params.databaseID + "/measurements",
        json: true,
        timeout,
        headers,
        qs: query,
        useQuerystring: true,
      };
      if (this.keepalive) {
        requestOptions.forever = true;
      }
  

      const retryPolicy = options.retryPolicy || this.retryPolicy || singleRetryPolicy;
      const backoffs = retryPolicy.backoffs();
      const logger = this.logger;
  
      let retries = 0;
      (function requestOnce() {
        request(requestOptions, (err, response, body) => {
          if (retries < backoffs.length && retryPolicy.retry(requestOptions, err, response, body)) {
            const backoff = backoffs[retries];
            retries += 1;
            setTimeout(requestOnce, backoff);
            return;
          }
          if (err) {
            err._fromRequest = true;
            responseLog(logger, requestOptions, response, err)
            rejecter(err);
            return;
          }

          switch (response.statusCode) {
            case 200:
              resolver(body);
              break;
            
            case 400:
              var err = new Errors.BadRequest(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 401:
              var err = new Errors.Unauthorized(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 404:
              var err = new Errors.NotFound(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 500:
              var err = new Errors.InternalError(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            default:
              var err = new Error("Received unexpected statusCode " + response.statusCode);
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
          }
        });
      }());
    });
  }

  /**
   * Get the available disks for a Atlas MongoDB Process
   * @param {Object} params
   * @param {string} params.groupID
   * @param {string} params.host
   * @param {number} params.port
   * @param {number} [params.pageNum]
   * @param {number} [params.itemsPerPage]
   * @param {object} [options]
   * @param {number} [options.timeout] - A request specific timeout
   * @param {external:Span} [options.span] - An OpenTracing span - For example from the parent request
   * @param {module:atlas-api-client.RetryPolicies} [options.retryPolicy] - A request specific retryPolicy
   * @param {function} [cb]
   * @returns {Promise}
   * @fulfill {Object}
   * @reject {module:atlas-api-client.Errors.BadRequest}
   * @reject {module:atlas-api-client.Errors.Unauthorized}
   * @reject {module:atlas-api-client.Errors.NotFound}
   * @reject {module:atlas-api-client.Errors.InternalError}
   * @reject {Error}
   */
  getProcessDisks(params, options, cb) {
    return this._hystrixCommand.execute(this._getProcessDisks, arguments);
  }
  _getProcessDisks(params, options, cb) {
    if (!cb && typeof options === "function") {
      cb = options;
      options = undefined;
    }

    return new Promise((resolve, reject) => {
      const rejecter = (err) => {
        reject(err);
        if (cb) {
          cb(err);
        }
      };
      const resolver = (data) => {
        resolve(data);
        if (cb) {
          cb(null, data);
        }
      };


      if (!options) {
        options = {};
      }

      const timeout = options.timeout || this.timeout;
      const tracer = options.tracer || this.tracer;
      const span = options.span;

      const headers = {};
      if (!params.groupID) {
        rejecter(new Error("groupID must be non-empty because it's a path parameter"));
        return;
      }
      if (!params.host) {
        rejecter(new Error("host must be non-empty because it's a path parameter"));
        return;
      }
      if (!params.port) {
        rejecter(new Error("port must be non-empty because it's a path parameter"));
        return;
      }

      const query = {};
      if (typeof params.pageNum !== "undefined") {
        query["pageNum"] = params.pageNum;
      }
  
      if (typeof params.itemsPerPage !== "undefined") {
        query["itemsPerPage"] = params.itemsPerPage;
      }
  

      if (span) {
        // Need to get tracer to inject. Use HTTP headers format so we can properly escape special characters
        tracer.inject(span, opentracing.FORMAT_HTTP_HEADERS, headers);
        span.logEvent("GET /api/atlas/v1.0/groups/{groupID}/processes/{host}:{port}/disks");
        span.setTag("span.kind", "client");
      }

	  const requestOptions = {
        method: "GET",
        uri: this.address + "/api/atlas/v1.0/groups/" + params.groupID + "/processes/" + params.host + ":" + params.port + "/disks",
        json: true,
        timeout,
        headers,
        qs: query,
        useQuerystring: true,
      };
      if (this.keepalive) {
        requestOptions.forever = true;
      }
  

      const retryPolicy = options.retryPolicy || this.retryPolicy || singleRetryPolicy;
      const backoffs = retryPolicy.backoffs();
      const logger = this.logger;
  
      let retries = 0;
      (function requestOnce() {
        request(requestOptions, (err, response, body) => {
          if (retries < backoffs.length && retryPolicy.retry(requestOptions, err, response, body)) {
            const backoff = backoffs[retries];
            retries += 1;
            setTimeout(requestOnce, backoff);
            return;
          }
          if (err) {
            err._fromRequest = true;
            responseLog(logger, requestOptions, response, err)
            rejecter(err);
            return;
          }

          switch (response.statusCode) {
            case 200:
              resolver(body);
              break;
            
            case 400:
              var err = new Errors.BadRequest(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 401:
              var err = new Errors.Unauthorized(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 404:
              var err = new Errors.NotFound(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 500:
              var err = new Errors.InternalError(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            default:
              var err = new Error("Received unexpected statusCode " + response.statusCode);
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
          }
        });
      }());
    });
  }

  /**
   * Get the measurements of the specified disk for a Atlas MongoDB process.
   * @param {Object} params
   * @param {string} params.groupID
   * @param {string} params.host
   * @param {number} params.port
   * @param {string} params.diskName
   * @param {string} params.granularity
   * @param {string} [params.period]
   * @param {string} [params.start]
   * @param {string} [params.end]
   * @param {string[]} [params.m]
   * @param {number} [params.pageNum]
   * @param {number} [params.itemsPerPage]
   * @param {object} [options]
   * @param {number} [options.timeout] - A request specific timeout
   * @param {external:Span} [options.span] - An OpenTracing span - For example from the parent request
   * @param {module:atlas-api-client.RetryPolicies} [options.retryPolicy] - A request specific retryPolicy
   * @param {function} [cb]
   * @returns {Promise}
   * @fulfill {Object}
   * @reject {module:atlas-api-client.Errors.BadRequest}
   * @reject {module:atlas-api-client.Errors.Unauthorized}
   * @reject {module:atlas-api-client.Errors.NotFound}
   * @reject {module:atlas-api-client.Errors.InternalError}
   * @reject {Error}
   */
  getProcessDiskMeasurements(params, options, cb) {
    return this._hystrixCommand.execute(this._getProcessDiskMeasurements, arguments);
  }
  _getProcessDiskMeasurements(params, options, cb) {
    if (!cb && typeof options === "function") {
      cb = options;
      options = undefined;
    }

    return new Promise((resolve, reject) => {
      const rejecter = (err) => {
        reject(err);
        if (cb) {
          cb(err);
        }
      };
      const resolver = (data) => {
        resolve(data);
        if (cb) {
          cb(null, data);
        }
      };


      if (!options) {
        options = {};
      }

      const timeout = options.timeout || this.timeout;
      const tracer = options.tracer || this.tracer;
      const span = options.span;

      const headers = {};
      if (!params.groupID) {
        rejecter(new Error("groupID must be non-empty because it's a path parameter"));
        return;
      }
      if (!params.host) {
        rejecter(new Error("host must be non-empty because it's a path parameter"));
        return;
      }
      if (!params.port) {
        rejecter(new Error("port must be non-empty because it's a path parameter"));
        return;
      }
      if (!params.diskName) {
        rejecter(new Error("diskName must be non-empty because it's a path parameter"));
        return;
      }

      const query = {};
      query["granularity"] = params.granularity;
  
      if (typeof params.period !== "undefined") {
        query["period"] = params.period;
      }
  
      if (typeof params.start !== "undefined") {
        query["start"] = params.start;
      }
  
      if (typeof params.end !== "undefined") {
        query["end"] = params.end;
      }
  
      if (typeof params.m !== "undefined") {
        query["m"] = params.m;
      }
  
      if (typeof params.pageNum !== "undefined") {
        query["pageNum"] = params.pageNum;
      }
  
      if (typeof params.itemsPerPage !== "undefined") {
        query["itemsPerPage"] = params.itemsPerPage;
      }
  

      if (span) {
        // Need to get tracer to inject. Use HTTP headers format so we can properly escape special characters
        tracer.inject(span, opentracing.FORMAT_HTTP_HEADERS, headers);
        span.logEvent("GET /api/atlas/v1.0/groups/{groupID}/processes/{host}:{port}/disks/{diskName}/measurements");
        span.setTag("span.kind", "client");
      }

	  const requestOptions = {
        method: "GET",
        uri: this.address + "/api/atlas/v1.0/groups/" + params.groupID + "/processes/" + params.host + ":" + params.port + "/disks/" + params.diskName + "/measurements",
        json: true,
        timeout,
        headers,
        qs: query,
        useQuerystring: true,
      };
      if (this.keepalive) {
        requestOptions.forever = true;
      }
  

      const retryPolicy = options.retryPolicy || this.retryPolicy || singleRetryPolicy;
      const backoffs = retryPolicy.backoffs();
      const logger = this.logger;
  
      let retries = 0;
      (function requestOnce() {
        request(requestOptions, (err, response, body) => {
          if (retries < backoffs.length && retryPolicy.retry(requestOptions, err, response, body)) {
            const backoff = backoffs[retries];
            retries += 1;
            setTimeout(requestOnce, backoff);
            return;
          }
          if (err) {
            err._fromRequest = true;
            responseLog(logger, requestOptions, response, err)
            rejecter(err);
            return;
          }

          switch (response.statusCode) {
            case 200:
              resolver(body);
              break;
            
            case 400:
              var err = new Errors.BadRequest(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 401:
              var err = new Errors.Unauthorized(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 404:
              var err = new Errors.NotFound(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 500:
              var err = new Errors.InternalError(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            default:
              var err = new Error("Received unexpected statusCode " + response.statusCode);
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
          }
        });
      }());
    });
  }

  /**
   * Get measurements for a specific Atlas MongoDB process (mongod or mongos).
   * @param {Object} params
   * @param {string} params.groupID
   * @param {string} params.host
   * @param {number} params.port
   * @param {string} params.granularity
   * @param {string} [params.period]
   * @param {string} [params.start]
   * @param {string} [params.end]
   * @param {string[]} [params.m]
   * @param {number} [params.pageNum]
   * @param {number} [params.itemsPerPage]
   * @param {object} [options]
   * @param {number} [options.timeout] - A request specific timeout
   * @param {external:Span} [options.span] - An OpenTracing span - For example from the parent request
   * @param {module:atlas-api-client.RetryPolicies} [options.retryPolicy] - A request specific retryPolicy
   * @param {function} [cb]
   * @returns {Promise}
   * @fulfill {Object}
   * @reject {module:atlas-api-client.Errors.BadRequest}
   * @reject {module:atlas-api-client.Errors.Unauthorized}
   * @reject {module:atlas-api-client.Errors.NotFound}
   * @reject {module:atlas-api-client.Errors.InternalError}
   * @reject {Error}
   */
  getProcessMeasurements(params, options, cb) {
    return this._hystrixCommand.execute(this._getProcessMeasurements, arguments);
  }
  _getProcessMeasurements(params, options, cb) {
    if (!cb && typeof options === "function") {
      cb = options;
      options = undefined;
    }

    return new Promise((resolve, reject) => {
      const rejecter = (err) => {
        reject(err);
        if (cb) {
          cb(err);
        }
      };
      const resolver = (data) => {
        resolve(data);
        if (cb) {
          cb(null, data);
        }
      };


      if (!options) {
        options = {};
      }

      const timeout = options.timeout || this.timeout;
      const tracer = options.tracer || this.tracer;
      const span = options.span;

      const headers = {};
      if (!params.groupID) {
        rejecter(new Error("groupID must be non-empty because it's a path parameter"));
        return;
      }
      if (!params.host) {
        rejecter(new Error("host must be non-empty because it's a path parameter"));
        return;
      }
      if (!params.port) {
        rejecter(new Error("port must be non-empty because it's a path parameter"));
        return;
      }

      const query = {};
      query["granularity"] = params.granularity;
  
      if (typeof params.period !== "undefined") {
        query["period"] = params.period;
      }
  
      if (typeof params.start !== "undefined") {
        query["start"] = params.start;
      }
  
      if (typeof params.end !== "undefined") {
        query["end"] = params.end;
      }
  
      if (typeof params.m !== "undefined") {
        query["m"] = params.m;
      }
  
      if (typeof params.pageNum !== "undefined") {
        query["pageNum"] = params.pageNum;
      }
  
      if (typeof params.itemsPerPage !== "undefined") {
        query["itemsPerPage"] = params.itemsPerPage;
      }
  

      if (span) {
        // Need to get tracer to inject. Use HTTP headers format so we can properly escape special characters
        tracer.inject(span, opentracing.FORMAT_HTTP_HEADERS, headers);
        span.logEvent("GET /api/atlas/v1.0/groups/{groupID}/processes/{host}:{port}/measurements");
        span.setTag("span.kind", "client");
      }

	  const requestOptions = {
        method: "GET",
        uri: this.address + "/api/atlas/v1.0/groups/" + params.groupID + "/processes/" + params.host + ":" + params.port + "/measurements",
        json: true,
        timeout,
        headers,
        qs: query,
        useQuerystring: true,
      };
      if (this.keepalive) {
        requestOptions.forever = true;
      }
  

      const retryPolicy = options.retryPolicy || this.retryPolicy || singleRetryPolicy;
      const backoffs = retryPolicy.backoffs();
      const logger = this.logger;
  
      let retries = 0;
      (function requestOnce() {
        request(requestOptions, (err, response, body) => {
          if (retries < backoffs.length && retryPolicy.retry(requestOptions, err, response, body)) {
            const backoff = backoffs[retries];
            retries += 1;
            setTimeout(requestOnce, backoff);
            return;
          }
          if (err) {
            err._fromRequest = true;
            responseLog(logger, requestOptions, response, err)
            rejecter(err);
            return;
          }

          switch (response.statusCode) {
            case 200:
              resolver(body);
              break;
            
            case 400:
              var err = new Errors.BadRequest(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 401:
              var err = new Errors.Unauthorized(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 404:
              var err = new Errors.NotFound(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            case 500:
              var err = new Errors.InternalError(body || {});
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
            
            default:
              var err = new Error("Received unexpected statusCode " + response.statusCode);
              responseLog(logger, requestOptions, response, err);
              rejecter(err);
              return;
          }
        });
      }());
    });
  }
};

module.exports = AtlasAPIClient;

/**
 * Retry policies available to use.
 * @alias module:atlas-api-client.RetryPolicies
 */
module.exports.RetryPolicies = {
  Single: singleRetryPolicy,
  Exponential: exponentialRetryPolicy,
  None: noRetryPolicy,
};

/**
 * Errors returned by methods.
 * @alias module:atlas-api-client.Errors
 */
module.exports.Errors = Errors;

module.exports.DefaultCircuitOptions = defaultCircuitOptions;
