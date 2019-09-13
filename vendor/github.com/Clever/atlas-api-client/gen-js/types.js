module.exports.Errors = {};

/**
 * BadRequest
 * @extends Error
 * @memberof module:atlas-api-client
 * @alias module:atlas-api-client.Errors.BadRequest
 * @property {string} detail
 * @property {number} error
 * @property {string} message
 * @property {string} reason
 */
module.exports.Errors.BadRequest = class extends Error {
  constructor(body) {
    super(body.message);
    for (const k of Object.keys(body)) {
      this[k] = body[k];
    }
  }
};

/**
 * Unauthorized
 * @extends Error
 * @memberof module:atlas-api-client
 * @alias module:atlas-api-client.Errors.Unauthorized
 * @property {string} detail
 * @property {number} error
 * @property {string} message
 * @property {string} reason
 */
module.exports.Errors.Unauthorized = class extends Error {
  constructor(body) {
    super(body.message);
    for (const k of Object.keys(body)) {
      this[k] = body[k];
    }
  }
};

/**
 * NotFound
 * @extends Error
 * @memberof module:atlas-api-client
 * @alias module:atlas-api-client.Errors.NotFound
 * @property {string} detail
 * @property {number} error
 * @property {string} message
 * @property {string} reason
 */
module.exports.Errors.NotFound = class extends Error {
  constructor(body) {
    super(body.message);
    for (const k of Object.keys(body)) {
      this[k] = body[k];
    }
  }
};

/**
 * InternalError
 * @extends Error
 * @memberof module:atlas-api-client
 * @alias module:atlas-api-client.Errors.InternalError
 * @property {string} detail
 * @property {number} error
 * @property {string} message
 * @property {string} reason
 */
module.exports.Errors.InternalError = class extends Error {
  constructor(body) {
    super(body.message);
    for (const k of Object.keys(body)) {
      this[k] = body[k];
    }
  }
};

