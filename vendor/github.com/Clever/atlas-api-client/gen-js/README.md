## Modules

<dl>
<dt><a href="#module_atlas-api-client">atlas-api-client</a></dt>
<dd><p>atlas-api-client client library.</p>
</dd>
</dl>

## Functions

<dl>
<dt><a href="#responseLog">responseLog()</a></dt>
<dd><p>Request status log is used to
to output the status of a request returned
by the client.</p>
</dd>
</dl>

<a name="module_atlas-api-client"></a>

## atlas-api-client
atlas-api-client client library.


* [atlas-api-client](#module_atlas-api-client)
    * [AtlasAPIClient](#exp_module_atlas-api-client--AtlasAPIClient) ⏏
        * [new AtlasAPIClient(options)](#new_module_atlas-api-client--AtlasAPIClient_new)
        * _instance_
            * [.getClusters(groupID, [options], [cb])](#module_atlas-api-client--AtlasAPIClient+getClusters) ⇒ <code>Promise</code>
            * [.createCluster(params, [options], [cb])](#module_atlas-api-client--AtlasAPIClient+createCluster) ⇒ <code>Promise</code>
            * [.deleteCluster(params, [options], [cb])](#module_atlas-api-client--AtlasAPIClient+deleteCluster) ⇒ <code>Promise</code>
            * [.getCluster(params, [options], [cb])](#module_atlas-api-client--AtlasAPIClient+getCluster) ⇒ <code>Promise</code>
            * [.updateCluster(params, [options], [cb])](#module_atlas-api-client--AtlasAPIClient+updateCluster) ⇒ <code>Promise</code>
            * [.getContainers(groupID, [options], [cb])](#module_atlas-api-client--AtlasAPIClient+getContainers) ⇒ <code>Promise</code>
            * [.createContainer(params, [options], [cb])](#module_atlas-api-client--AtlasAPIClient+createContainer) ⇒ <code>Promise</code>
            * [.getContainer(params, [options], [cb])](#module_atlas-api-client--AtlasAPIClient+getContainer) ⇒ <code>Promise</code>
            * [.updateContainer(params, [options], [cb])](#module_atlas-api-client--AtlasAPIClient+updateContainer) ⇒ <code>Promise</code>
            * [.getDatabaseUsers(groupID, [options], [cb])](#module_atlas-api-client--AtlasAPIClient+getDatabaseUsers) ⇒ <code>Promise</code>
            * [.createDatabaseUser(params, [options], [cb])](#module_atlas-api-client--AtlasAPIClient+createDatabaseUser) ⇒ <code>Promise</code>
            * [.deleteDatabaseUser(params, [options], [cb])](#module_atlas-api-client--AtlasAPIClient+deleteDatabaseUser) ⇒ <code>Promise</code>
            * [.getDatabaseUser(params, [options], [cb])](#module_atlas-api-client--AtlasAPIClient+getDatabaseUser) ⇒ <code>Promise</code>
            * [.updateDatabaseUser(params, [options], [cb])](#module_atlas-api-client--AtlasAPIClient+updateDatabaseUser) ⇒ <code>Promise</code>
            * [.getProcesses(groupID, [options], [cb])](#module_atlas-api-client--AtlasAPIClient+getProcesses) ⇒ <code>Promise</code>
            * [.getProcessDatabases(params, [options], [cb])](#module_atlas-api-client--AtlasAPIClient+getProcessDatabases) ⇒ <code>Promise</code>
            * [.getProcessDatabaseMeasurements(params, [options], [cb])](#module_atlas-api-client--AtlasAPIClient+getProcessDatabaseMeasurements) ⇒ <code>Promise</code>
            * [.getProcessDisks(params, [options], [cb])](#module_atlas-api-client--AtlasAPIClient+getProcessDisks) ⇒ <code>Promise</code>
            * [.getProcessDiskMeasurements(params, [options], [cb])](#module_atlas-api-client--AtlasAPIClient+getProcessDiskMeasurements) ⇒ <code>Promise</code>
            * [.getProcessMeasurements(params, [options], [cb])](#module_atlas-api-client--AtlasAPIClient+getProcessMeasurements) ⇒ <code>Promise</code>
        * _static_
            * [.RetryPolicies](#module_atlas-api-client--AtlasAPIClient.RetryPolicies)
                * [.Exponential](#module_atlas-api-client--AtlasAPIClient.RetryPolicies.Exponential)
                * [.Single](#module_atlas-api-client--AtlasAPIClient.RetryPolicies.Single)
                * [.None](#module_atlas-api-client--AtlasAPIClient.RetryPolicies.None)
            * [.Errors](#module_atlas-api-client--AtlasAPIClient.Errors)
                * [.BadRequest](#module_atlas-api-client--AtlasAPIClient.Errors.BadRequest) ⇐ <code>Error</code>
                * [.Unauthorized](#module_atlas-api-client--AtlasAPIClient.Errors.Unauthorized) ⇐ <code>Error</code>
                * [.NotFound](#module_atlas-api-client--AtlasAPIClient.Errors.NotFound) ⇐ <code>Error</code>
                * [.InternalError](#module_atlas-api-client--AtlasAPIClient.Errors.InternalError) ⇐ <code>Error</code>
            * [.DefaultCircuitOptions](#module_atlas-api-client--AtlasAPIClient.DefaultCircuitOptions)

<a name="exp_module_atlas-api-client--AtlasAPIClient"></a>

### AtlasAPIClient ⏏
atlas-api-client client

**Kind**: Exported class  
<a name="new_module_atlas-api-client--AtlasAPIClient_new"></a>

#### new AtlasAPIClient(options)
Create a new client object.


| Param | Type | Default | Description |
| --- | --- | --- | --- |
| options | <code>Object</code> |  | Options for constructing a client object. |
| [options.address] | <code>string</code> |  | URL where the server is located. Must provide this or the discovery argument |
| [options.discovery] | <code>bool</code> |  | Use clever-discovery to locate the server. Must provide this or the address argument |
| [options.timeout] | <code>number</code> |  | The timeout to use for all client requests, in milliseconds. This can be overridden on a per-request basis. Default is 5000ms. |
| [options.keepalive] | <code>bool</code> |  | Set keepalive to true for client requests. This sets the forever: true attribute in request. Defaults to false |
| [options.retryPolicy] | <code>[RetryPolicies](#module_atlas-api-client--AtlasAPIClient.RetryPolicies)</code> | <code>RetryPolicies.Single</code> | The logic to determine which requests to retry, as well as how many times to retry. |
| [options.logger] | <code>module:kayvee.Logger</code> | <code>logger.New(&quot;atlas-api-client-wagclient&quot;)</code> | The Kayvee logger to use in the client. |
| [options.circuit] | <code>Object</code> |  | Options for constructing the client's circuit breaker. |
| [options.circuit.forceClosed] | <code>bool</code> |  | When set to true the circuit will always be closed. Default: true. |
| [options.circuit.maxConcurrentRequests] | <code>number</code> |  | the maximum number of concurrent requests the client can make at the same time. Default: 100. |
| [options.circuit.requestVolumeThreshold] | <code>number</code> |  | The minimum number of requests needed before a circuit can be tripped due to health. Default: 20. |
| [options.circuit.sleepWindow] | <code>number</code> |  | how long, in milliseconds, to wait after a circuit opens before testing for recovery. Default: 5000. |
| [options.circuit.errorPercentThreshold] | <code>number</code> |  | the threshold to place on the rolling error rate. Once the error rate exceeds this percentage, the circuit opens. Default: 90. |

<a name="module_atlas-api-client--AtlasAPIClient+getClusters"></a>

#### atlasAPIClient.getClusters(groupID, [options], [cb]) ⇒ <code>Promise</code>
Get All Clusters

**Kind**: instance method of <code>[AtlasAPIClient](#exp_module_atlas-api-client--AtlasAPIClient)</code>  
**Fulfill**: <code>Object</code>  
**Reject**: <code>[BadRequest](#module_atlas-api-client--AtlasAPIClient.Errors.BadRequest)</code>  
**Reject**: <code>[Unauthorized](#module_atlas-api-client--AtlasAPIClient.Errors.Unauthorized)</code>  
**Reject**: <code>[NotFound](#module_atlas-api-client--AtlasAPIClient.Errors.NotFound)</code>  
**Reject**: <code>[InternalError](#module_atlas-api-client--AtlasAPIClient.Errors.InternalError)</code>  
**Reject**: <code>Error</code>  

| Param | Type | Description |
| --- | --- | --- |
| groupID | <code>string</code> |  |
| [options] | <code>object</code> |  |
| [options.timeout] | <code>number</code> | A request specific timeout |
| [options.span] | <code>[Span](https://doc.esdoc.org/github.com/opentracing/opentracing-javascript/class/src/span.js~Span.html)</code> | An OpenTracing span - For example from the parent request |
| [options.retryPolicy] | <code>[RetryPolicies](#module_atlas-api-client--AtlasAPIClient.RetryPolicies)</code> | A request specific retryPolicy |
| [cb] | <code>function</code> |  |

<a name="module_atlas-api-client--AtlasAPIClient+createCluster"></a>

#### atlasAPIClient.createCluster(params, [options], [cb]) ⇒ <code>Promise</code>
Create a Cluster

**Kind**: instance method of <code>[AtlasAPIClient](#exp_module_atlas-api-client--AtlasAPIClient)</code>  
**Fulfill**: <code>Object</code>  
**Reject**: <code>[BadRequest](#module_atlas-api-client--AtlasAPIClient.Errors.BadRequest)</code>  
**Reject**: <code>[Unauthorized](#module_atlas-api-client--AtlasAPIClient.Errors.Unauthorized)</code>  
**Reject**: <code>[NotFound](#module_atlas-api-client--AtlasAPIClient.Errors.NotFound)</code>  
**Reject**: <code>[InternalError](#module_atlas-api-client--AtlasAPIClient.Errors.InternalError)</code>  
**Reject**: <code>Error</code>  

| Param | Type | Description |
| --- | --- | --- |
| params | <code>Object</code> |  |
| params.groupID | <code>string</code> |  |
| params.createOrUpdateClusterRequest |  |  |
| [options] | <code>object</code> |  |
| [options.timeout] | <code>number</code> | A request specific timeout |
| [options.span] | <code>[Span](https://doc.esdoc.org/github.com/opentracing/opentracing-javascript/class/src/span.js~Span.html)</code> | An OpenTracing span - For example from the parent request |
| [options.retryPolicy] | <code>[RetryPolicies](#module_atlas-api-client--AtlasAPIClient.RetryPolicies)</code> | A request specific retryPolicy |
| [cb] | <code>function</code> |  |

<a name="module_atlas-api-client--AtlasAPIClient+deleteCluster"></a>

#### atlasAPIClient.deleteCluster(params, [options], [cb]) ⇒ <code>Promise</code>
Deletes a cluster

**Kind**: instance method of <code>[AtlasAPIClient](#exp_module_atlas-api-client--AtlasAPIClient)</code>  
**Fulfill**: <code>undefined</code>  
**Reject**: <code>[BadRequest](#module_atlas-api-client--AtlasAPIClient.Errors.BadRequest)</code>  
**Reject**: <code>[Unauthorized](#module_atlas-api-client--AtlasAPIClient.Errors.Unauthorized)</code>  
**Reject**: <code>[NotFound](#module_atlas-api-client--AtlasAPIClient.Errors.NotFound)</code>  
**Reject**: <code>[InternalError](#module_atlas-api-client--AtlasAPIClient.Errors.InternalError)</code>  
**Reject**: <code>Error</code>  

| Param | Type | Description |
| --- | --- | --- |
| params | <code>Object</code> |  |
| params.groupID | <code>string</code> |  |
| params.clusterName | <code>string</code> |  |
| [options] | <code>object</code> |  |
| [options.timeout] | <code>number</code> | A request specific timeout |
| [options.span] | <code>[Span](https://doc.esdoc.org/github.com/opentracing/opentracing-javascript/class/src/span.js~Span.html)</code> | An OpenTracing span - For example from the parent request |
| [options.retryPolicy] | <code>[RetryPolicies](#module_atlas-api-client--AtlasAPIClient.RetryPolicies)</code> | A request specific retryPolicy |
| [cb] | <code>function</code> |  |

<a name="module_atlas-api-client--AtlasAPIClient+getCluster"></a>

#### atlasAPIClient.getCluster(params, [options], [cb]) ⇒ <code>Promise</code>
Gets a cluster

**Kind**: instance method of <code>[AtlasAPIClient](#exp_module_atlas-api-client--AtlasAPIClient)</code>  
**Fulfill**: <code>Object</code>  
**Reject**: <code>[BadRequest](#module_atlas-api-client--AtlasAPIClient.Errors.BadRequest)</code>  
**Reject**: <code>[NotFound](#module_atlas-api-client--AtlasAPIClient.Errors.NotFound)</code>  
**Reject**: <code>[InternalError](#module_atlas-api-client--AtlasAPIClient.Errors.InternalError)</code>  
**Reject**: <code>Error</code>  

| Param | Type | Description |
| --- | --- | --- |
| params | <code>Object</code> |  |
| params.groupID | <code>string</code> |  |
| params.clusterName | <code>string</code> |  |
| [options] | <code>object</code> |  |
| [options.timeout] | <code>number</code> | A request specific timeout |
| [options.span] | <code>[Span](https://doc.esdoc.org/github.com/opentracing/opentracing-javascript/class/src/span.js~Span.html)</code> | An OpenTracing span - For example from the parent request |
| [options.retryPolicy] | <code>[RetryPolicies](#module_atlas-api-client--AtlasAPIClient.RetryPolicies)</code> | A request specific retryPolicy |
| [cb] | <code>function</code> |  |

<a name="module_atlas-api-client--AtlasAPIClient+updateCluster"></a>

#### atlasAPIClient.updateCluster(params, [options], [cb]) ⇒ <code>Promise</code>
Update a Cluster

**Kind**: instance method of <code>[AtlasAPIClient](#exp_module_atlas-api-client--AtlasAPIClient)</code>  
**Fulfill**: <code>Object</code>  
**Reject**: <code>[BadRequest](#module_atlas-api-client--AtlasAPIClient.Errors.BadRequest)</code>  
**Reject**: <code>[Unauthorized](#module_atlas-api-client--AtlasAPIClient.Errors.Unauthorized)</code>  
**Reject**: <code>[NotFound](#module_atlas-api-client--AtlasAPIClient.Errors.NotFound)</code>  
**Reject**: <code>[InternalError](#module_atlas-api-client--AtlasAPIClient.Errors.InternalError)</code>  
**Reject**: <code>Error</code>  

| Param | Type | Description |
| --- | --- | --- |
| params | <code>Object</code> |  |
| params.groupID | <code>string</code> |  |
| params.clusterName | <code>string</code> |  |
| params.createOrUpdateClusterRequest |  |  |
| [options] | <code>object</code> |  |
| [options.timeout] | <code>number</code> | A request specific timeout |
| [options.span] | <code>[Span](https://doc.esdoc.org/github.com/opentracing/opentracing-javascript/class/src/span.js~Span.html)</code> | An OpenTracing span - For example from the parent request |
| [options.retryPolicy] | <code>[RetryPolicies](#module_atlas-api-client--AtlasAPIClient.RetryPolicies)</code> | A request specific retryPolicy |
| [cb] | <code>function</code> |  |

<a name="module_atlas-api-client--AtlasAPIClient+getContainers"></a>

#### atlasAPIClient.getContainers(groupID, [options], [cb]) ⇒ <code>Promise</code>
Get All Containers

**Kind**: instance method of <code>[AtlasAPIClient](#exp_module_atlas-api-client--AtlasAPIClient)</code>  
**Fulfill**: <code>Object</code>  
**Reject**: <code>[BadRequest](#module_atlas-api-client--AtlasAPIClient.Errors.BadRequest)</code>  
**Reject**: <code>[Unauthorized](#module_atlas-api-client--AtlasAPIClient.Errors.Unauthorized)</code>  
**Reject**: <code>[NotFound](#module_atlas-api-client--AtlasAPIClient.Errors.NotFound)</code>  
**Reject**: <code>[InternalError](#module_atlas-api-client--AtlasAPIClient.Errors.InternalError)</code>  
**Reject**: <code>Error</code>  

| Param | Type | Description |
| --- | --- | --- |
| groupID | <code>string</code> |  |
| [options] | <code>object</code> |  |
| [options.timeout] | <code>number</code> | A request specific timeout |
| [options.span] | <code>[Span](https://doc.esdoc.org/github.com/opentracing/opentracing-javascript/class/src/span.js~Span.html)</code> | An OpenTracing span - For example from the parent request |
| [options.retryPolicy] | <code>[RetryPolicies](#module_atlas-api-client--AtlasAPIClient.RetryPolicies)</code> | A request specific retryPolicy |
| [cb] | <code>function</code> |  |

<a name="module_atlas-api-client--AtlasAPIClient+createContainer"></a>

#### atlasAPIClient.createContainer(params, [options], [cb]) ⇒ <code>Promise</code>
Create a Container

**Kind**: instance method of <code>[AtlasAPIClient](#exp_module_atlas-api-client--AtlasAPIClient)</code>  
**Fulfill**: <code>Object</code>  
**Reject**: <code>[BadRequest](#module_atlas-api-client--AtlasAPIClient.Errors.BadRequest)</code>  
**Reject**: <code>[Unauthorized](#module_atlas-api-client--AtlasAPIClient.Errors.Unauthorized)</code>  
**Reject**: <code>[NotFound](#module_atlas-api-client--AtlasAPIClient.Errors.NotFound)</code>  
**Reject**: <code>[InternalError](#module_atlas-api-client--AtlasAPIClient.Errors.InternalError)</code>  
**Reject**: <code>Error</code>  

| Param | Type | Description |
| --- | --- | --- |
| params | <code>Object</code> |  |
| params.groupID | <code>string</code> |  |
| params.createOrUpdateContainerRequest |  |  |
| [options] | <code>object</code> |  |
| [options.timeout] | <code>number</code> | A request specific timeout |
| [options.span] | <code>[Span](https://doc.esdoc.org/github.com/opentracing/opentracing-javascript/class/src/span.js~Span.html)</code> | An OpenTracing span - For example from the parent request |
| [options.retryPolicy] | <code>[RetryPolicies](#module_atlas-api-client--AtlasAPIClient.RetryPolicies)</code> | A request specific retryPolicy |
| [cb] | <code>function</code> |  |

<a name="module_atlas-api-client--AtlasAPIClient+getContainer"></a>

#### atlasAPIClient.getContainer(params, [options], [cb]) ⇒ <code>Promise</code>
Gets a container

**Kind**: instance method of <code>[AtlasAPIClient](#exp_module_atlas-api-client--AtlasAPIClient)</code>  
**Fulfill**: <code>Object</code>  
**Reject**: <code>[BadRequest](#module_atlas-api-client--AtlasAPIClient.Errors.BadRequest)</code>  
**Reject**: <code>[NotFound](#module_atlas-api-client--AtlasAPIClient.Errors.NotFound)</code>  
**Reject**: <code>[InternalError](#module_atlas-api-client--AtlasAPIClient.Errors.InternalError)</code>  
**Reject**: <code>Error</code>  

| Param | Type | Description |
| --- | --- | --- |
| params | <code>Object</code> |  |
| params.groupID | <code>string</code> |  |
| params.containerID | <code>string</code> |  |
| [options] | <code>object</code> |  |
| [options.timeout] | <code>number</code> | A request specific timeout |
| [options.span] | <code>[Span](https://doc.esdoc.org/github.com/opentracing/opentracing-javascript/class/src/span.js~Span.html)</code> | An OpenTracing span - For example from the parent request |
| [options.retryPolicy] | <code>[RetryPolicies](#module_atlas-api-client--AtlasAPIClient.RetryPolicies)</code> | A request specific retryPolicy |
| [cb] | <code>function</code> |  |

<a name="module_atlas-api-client--AtlasAPIClient+updateContainer"></a>

#### atlasAPIClient.updateContainer(params, [options], [cb]) ⇒ <code>Promise</code>
Update a Container

**Kind**: instance method of <code>[AtlasAPIClient](#exp_module_atlas-api-client--AtlasAPIClient)</code>  
**Fulfill**: <code>Object</code>  
**Reject**: <code>[BadRequest](#module_atlas-api-client--AtlasAPIClient.Errors.BadRequest)</code>  
**Reject**: <code>[Unauthorized](#module_atlas-api-client--AtlasAPIClient.Errors.Unauthorized)</code>  
**Reject**: <code>[NotFound](#module_atlas-api-client--AtlasAPIClient.Errors.NotFound)</code>  
**Reject**: <code>[InternalError](#module_atlas-api-client--AtlasAPIClient.Errors.InternalError)</code>  
**Reject**: <code>Error</code>  

| Param | Type | Description |
| --- | --- | --- |
| params | <code>Object</code> |  |
| params.groupID | <code>string</code> |  |
| params.containerID | <code>string</code> |  |
| params.createOrUpdateContainerRequest |  |  |
| [options] | <code>object</code> |  |
| [options.timeout] | <code>number</code> | A request specific timeout |
| [options.span] | <code>[Span](https://doc.esdoc.org/github.com/opentracing/opentracing-javascript/class/src/span.js~Span.html)</code> | An OpenTracing span - For example from the parent request |
| [options.retryPolicy] | <code>[RetryPolicies](#module_atlas-api-client--AtlasAPIClient.RetryPolicies)</code> | A request specific retryPolicy |
| [cb] | <code>function</code> |  |

<a name="module_atlas-api-client--AtlasAPIClient+getDatabaseUsers"></a>

#### atlasAPIClient.getDatabaseUsers(groupID, [options], [cb]) ⇒ <code>Promise</code>
Get All DatabaseUsers

**Kind**: instance method of <code>[AtlasAPIClient](#exp_module_atlas-api-client--AtlasAPIClient)</code>  
**Fulfill**: <code>Object</code>  
**Reject**: <code>[BadRequest](#module_atlas-api-client--AtlasAPIClient.Errors.BadRequest)</code>  
**Reject**: <code>[Unauthorized](#module_atlas-api-client--AtlasAPIClient.Errors.Unauthorized)</code>  
**Reject**: <code>[NotFound](#module_atlas-api-client--AtlasAPIClient.Errors.NotFound)</code>  
**Reject**: <code>[InternalError](#module_atlas-api-client--AtlasAPIClient.Errors.InternalError)</code>  
**Reject**: <code>Error</code>  

| Param | Type | Description |
| --- | --- | --- |
| groupID | <code>string</code> |  |
| [options] | <code>object</code> |  |
| [options.timeout] | <code>number</code> | A request specific timeout |
| [options.span] | <code>[Span](https://doc.esdoc.org/github.com/opentracing/opentracing-javascript/class/src/span.js~Span.html)</code> | An OpenTracing span - For example from the parent request |
| [options.retryPolicy] | <code>[RetryPolicies](#module_atlas-api-client--AtlasAPIClient.RetryPolicies)</code> | A request specific retryPolicy |
| [cb] | <code>function</code> |  |

<a name="module_atlas-api-client--AtlasAPIClient+createDatabaseUser"></a>

#### atlasAPIClient.createDatabaseUser(params, [options], [cb]) ⇒ <code>Promise</code>
Create a DatabaseUser

**Kind**: instance method of <code>[AtlasAPIClient](#exp_module_atlas-api-client--AtlasAPIClient)</code>  
**Fulfill**: <code>Object</code>  
**Reject**: <code>[BadRequest](#module_atlas-api-client--AtlasAPIClient.Errors.BadRequest)</code>  
**Reject**: <code>[Unauthorized](#module_atlas-api-client--AtlasAPIClient.Errors.Unauthorized)</code>  
**Reject**: <code>[NotFound](#module_atlas-api-client--AtlasAPIClient.Errors.NotFound)</code>  
**Reject**: <code>[InternalError](#module_atlas-api-client--AtlasAPIClient.Errors.InternalError)</code>  
**Reject**: <code>Error</code>  

| Param | Type | Description |
| --- | --- | --- |
| params | <code>Object</code> |  |
| params.groupID | <code>string</code> |  |
| params.createDatabaseUserRequest |  |  |
| [options] | <code>object</code> |  |
| [options.timeout] | <code>number</code> | A request specific timeout |
| [options.span] | <code>[Span](https://doc.esdoc.org/github.com/opentracing/opentracing-javascript/class/src/span.js~Span.html)</code> | An OpenTracing span - For example from the parent request |
| [options.retryPolicy] | <code>[RetryPolicies](#module_atlas-api-client--AtlasAPIClient.RetryPolicies)</code> | A request specific retryPolicy |
| [cb] | <code>function</code> |  |

<a name="module_atlas-api-client--AtlasAPIClient+deleteDatabaseUser"></a>

#### atlasAPIClient.deleteDatabaseUser(params, [options], [cb]) ⇒ <code>Promise</code>
Deletes a DatabaseUser

**Kind**: instance method of <code>[AtlasAPIClient](#exp_module_atlas-api-client--AtlasAPIClient)</code>  
**Fulfill**: <code>undefined</code>  
**Reject**: <code>[BadRequest](#module_atlas-api-client--AtlasAPIClient.Errors.BadRequest)</code>  
**Reject**: <code>[Unauthorized](#module_atlas-api-client--AtlasAPIClient.Errors.Unauthorized)</code>  
**Reject**: <code>[NotFound](#module_atlas-api-client--AtlasAPIClient.Errors.NotFound)</code>  
**Reject**: <code>[InternalError](#module_atlas-api-client--AtlasAPIClient.Errors.InternalError)</code>  
**Reject**: <code>Error</code>  

| Param | Type | Description |
| --- | --- | --- |
| params | <code>Object</code> |  |
| params.groupID | <code>string</code> |  |
| params.username | <code>string</code> |  |
| [options] | <code>object</code> |  |
| [options.timeout] | <code>number</code> | A request specific timeout |
| [options.span] | <code>[Span](https://doc.esdoc.org/github.com/opentracing/opentracing-javascript/class/src/span.js~Span.html)</code> | An OpenTracing span - For example from the parent request |
| [options.retryPolicy] | <code>[RetryPolicies](#module_atlas-api-client--AtlasAPIClient.RetryPolicies)</code> | A request specific retryPolicy |
| [cb] | <code>function</code> |  |

<a name="module_atlas-api-client--AtlasAPIClient+getDatabaseUser"></a>

#### atlasAPIClient.getDatabaseUser(params, [options], [cb]) ⇒ <code>Promise</code>
Gets a database user

**Kind**: instance method of <code>[AtlasAPIClient](#exp_module_atlas-api-client--AtlasAPIClient)</code>  
**Fulfill**: <code>Object</code>  
**Reject**: <code>[BadRequest](#module_atlas-api-client--AtlasAPIClient.Errors.BadRequest)</code>  
**Reject**: <code>[NotFound](#module_atlas-api-client--AtlasAPIClient.Errors.NotFound)</code>  
**Reject**: <code>[InternalError](#module_atlas-api-client--AtlasAPIClient.Errors.InternalError)</code>  
**Reject**: <code>Error</code>  

| Param | Type | Description |
| --- | --- | --- |
| params | <code>Object</code> |  |
| params.groupID | <code>string</code> |  |
| params.username | <code>string</code> |  |
| [options] | <code>object</code> |  |
| [options.timeout] | <code>number</code> | A request specific timeout |
| [options.span] | <code>[Span](https://doc.esdoc.org/github.com/opentracing/opentracing-javascript/class/src/span.js~Span.html)</code> | An OpenTracing span - For example from the parent request |
| [options.retryPolicy] | <code>[RetryPolicies](#module_atlas-api-client--AtlasAPIClient.RetryPolicies)</code> | A request specific retryPolicy |
| [cb] | <code>function</code> |  |

<a name="module_atlas-api-client--AtlasAPIClient+updateDatabaseUser"></a>

#### atlasAPIClient.updateDatabaseUser(params, [options], [cb]) ⇒ <code>Promise</code>
Update a DatabaseUser

**Kind**: instance method of <code>[AtlasAPIClient](#exp_module_atlas-api-client--AtlasAPIClient)</code>  
**Fulfill**: <code>Object</code>  
**Reject**: <code>[BadRequest](#module_atlas-api-client--AtlasAPIClient.Errors.BadRequest)</code>  
**Reject**: <code>[Unauthorized](#module_atlas-api-client--AtlasAPIClient.Errors.Unauthorized)</code>  
**Reject**: <code>[NotFound](#module_atlas-api-client--AtlasAPIClient.Errors.NotFound)</code>  
**Reject**: <code>[InternalError](#module_atlas-api-client--AtlasAPIClient.Errors.InternalError)</code>  
**Reject**: <code>Error</code>  

| Param | Type | Description |
| --- | --- | --- |
| params | <code>Object</code> |  |
| params.groupID | <code>string</code> |  |
| params.username | <code>string</code> |  |
| params.updateDatabaseUserRequest |  |  |
| [options] | <code>object</code> |  |
| [options.timeout] | <code>number</code> | A request specific timeout |
| [options.span] | <code>[Span](https://doc.esdoc.org/github.com/opentracing/opentracing-javascript/class/src/span.js~Span.html)</code> | An OpenTracing span - For example from the parent request |
| [options.retryPolicy] | <code>[RetryPolicies](#module_atlas-api-client--AtlasAPIClient.RetryPolicies)</code> | A request specific retryPolicy |
| [cb] | <code>function</code> |  |

<a name="module_atlas-api-client--AtlasAPIClient+getProcesses"></a>

#### atlasAPIClient.getProcesses(groupID, [options], [cb]) ⇒ <code>Promise</code>
Get All Processes

**Kind**: instance method of <code>[AtlasAPIClient](#exp_module_atlas-api-client--AtlasAPIClient)</code>  
**Fulfill**: <code>Object</code>  
**Reject**: <code>[BadRequest](#module_atlas-api-client--AtlasAPIClient.Errors.BadRequest)</code>  
**Reject**: <code>[Unauthorized](#module_atlas-api-client--AtlasAPIClient.Errors.Unauthorized)</code>  
**Reject**: <code>[NotFound](#module_atlas-api-client--AtlasAPIClient.Errors.NotFound)</code>  
**Reject**: <code>[InternalError](#module_atlas-api-client--AtlasAPIClient.Errors.InternalError)</code>  
**Reject**: <code>Error</code>  

| Param | Type | Description |
| --- | --- | --- |
| groupID | <code>string</code> |  |
| [options] | <code>object</code> |  |
| [options.timeout] | <code>number</code> | A request specific timeout |
| [options.span] | <code>[Span](https://doc.esdoc.org/github.com/opentracing/opentracing-javascript/class/src/span.js~Span.html)</code> | An OpenTracing span - For example from the parent request |
| [options.retryPolicy] | <code>[RetryPolicies](#module_atlas-api-client--AtlasAPIClient.RetryPolicies)</code> | A request specific retryPolicy |
| [cb] | <code>function</code> |  |

<a name="module_atlas-api-client--AtlasAPIClient+getProcessDatabases"></a>

#### atlasAPIClient.getProcessDatabases(params, [options], [cb]) ⇒ <code>Promise</code>
Get the available databases for a Atlas MongoDB Process

**Kind**: instance method of <code>[AtlasAPIClient](#exp_module_atlas-api-client--AtlasAPIClient)</code>  
**Fulfill**: <code>Object</code>  
**Reject**: <code>[BadRequest](#module_atlas-api-client--AtlasAPIClient.Errors.BadRequest)</code>  
**Reject**: <code>[Unauthorized](#module_atlas-api-client--AtlasAPIClient.Errors.Unauthorized)</code>  
**Reject**: <code>[NotFound](#module_atlas-api-client--AtlasAPIClient.Errors.NotFound)</code>  
**Reject**: <code>[InternalError](#module_atlas-api-client--AtlasAPIClient.Errors.InternalError)</code>  
**Reject**: <code>Error</code>  

| Param | Type | Description |
| --- | --- | --- |
| params | <code>Object</code> |  |
| params.groupID | <code>string</code> |  |
| params.host | <code>string</code> |  |
| params.port | <code>number</code> |  |
| [params.pageNum] | <code>number</code> |  |
| [params.itemsPerPage] | <code>number</code> |  |
| [options] | <code>object</code> |  |
| [options.timeout] | <code>number</code> | A request specific timeout |
| [options.span] | <code>[Span](https://doc.esdoc.org/github.com/opentracing/opentracing-javascript/class/src/span.js~Span.html)</code> | An OpenTracing span - For example from the parent request |
| [options.retryPolicy] | <code>[RetryPolicies](#module_atlas-api-client--AtlasAPIClient.RetryPolicies)</code> | A request specific retryPolicy |
| [cb] | <code>function</code> |  |

<a name="module_atlas-api-client--AtlasAPIClient+getProcessDatabaseMeasurements"></a>

#### atlasAPIClient.getProcessDatabaseMeasurements(params, [options], [cb]) ⇒ <code>Promise</code>
Get the measurements of the specified database for a Atlas MongoDB process.

**Kind**: instance method of <code>[AtlasAPIClient](#exp_module_atlas-api-client--AtlasAPIClient)</code>  
**Fulfill**: <code>Object</code>  
**Reject**: <code>[BadRequest](#module_atlas-api-client--AtlasAPIClient.Errors.BadRequest)</code>  
**Reject**: <code>[Unauthorized](#module_atlas-api-client--AtlasAPIClient.Errors.Unauthorized)</code>  
**Reject**: <code>[NotFound](#module_atlas-api-client--AtlasAPIClient.Errors.NotFound)</code>  
**Reject**: <code>[InternalError](#module_atlas-api-client--AtlasAPIClient.Errors.InternalError)</code>  
**Reject**: <code>Error</code>  

| Param | Type | Description |
| --- | --- | --- |
| params | <code>Object</code> |  |
| params.groupID | <code>string</code> |  |
| params.host | <code>string</code> |  |
| params.port | <code>number</code> |  |
| params.databaseID | <code>string</code> |  |
| params.granularity | <code>string</code> |  |
| [params.period] | <code>string</code> |  |
| [params.start] | <code>string</code> |  |
| [params.end] | <code>string</code> |  |
| [params.m] | <code>Array.&lt;string&gt;</code> |  |
| [params.pageNum] | <code>number</code> |  |
| [params.itemsPerPage] | <code>number</code> |  |
| [options] | <code>object</code> |  |
| [options.timeout] | <code>number</code> | A request specific timeout |
| [options.span] | <code>[Span](https://doc.esdoc.org/github.com/opentracing/opentracing-javascript/class/src/span.js~Span.html)</code> | An OpenTracing span - For example from the parent request |
| [options.retryPolicy] | <code>[RetryPolicies](#module_atlas-api-client--AtlasAPIClient.RetryPolicies)</code> | A request specific retryPolicy |
| [cb] | <code>function</code> |  |

<a name="module_atlas-api-client--AtlasAPIClient+getProcessDisks"></a>

#### atlasAPIClient.getProcessDisks(params, [options], [cb]) ⇒ <code>Promise</code>
Get the available disks for a Atlas MongoDB Process

**Kind**: instance method of <code>[AtlasAPIClient](#exp_module_atlas-api-client--AtlasAPIClient)</code>  
**Fulfill**: <code>Object</code>  
**Reject**: <code>[BadRequest](#module_atlas-api-client--AtlasAPIClient.Errors.BadRequest)</code>  
**Reject**: <code>[Unauthorized](#module_atlas-api-client--AtlasAPIClient.Errors.Unauthorized)</code>  
**Reject**: <code>[NotFound](#module_atlas-api-client--AtlasAPIClient.Errors.NotFound)</code>  
**Reject**: <code>[InternalError](#module_atlas-api-client--AtlasAPIClient.Errors.InternalError)</code>  
**Reject**: <code>Error</code>  

| Param | Type | Description |
| --- | --- | --- |
| params | <code>Object</code> |  |
| params.groupID | <code>string</code> |  |
| params.host | <code>string</code> |  |
| params.port | <code>number</code> |  |
| [params.pageNum] | <code>number</code> |  |
| [params.itemsPerPage] | <code>number</code> |  |
| [options] | <code>object</code> |  |
| [options.timeout] | <code>number</code> | A request specific timeout |
| [options.span] | <code>[Span](https://doc.esdoc.org/github.com/opentracing/opentracing-javascript/class/src/span.js~Span.html)</code> | An OpenTracing span - For example from the parent request |
| [options.retryPolicy] | <code>[RetryPolicies](#module_atlas-api-client--AtlasAPIClient.RetryPolicies)</code> | A request specific retryPolicy |
| [cb] | <code>function</code> |  |

<a name="module_atlas-api-client--AtlasAPIClient+getProcessDiskMeasurements"></a>

#### atlasAPIClient.getProcessDiskMeasurements(params, [options], [cb]) ⇒ <code>Promise</code>
Get the measurements of the specified disk for a Atlas MongoDB process.

**Kind**: instance method of <code>[AtlasAPIClient](#exp_module_atlas-api-client--AtlasAPIClient)</code>  
**Fulfill**: <code>Object</code>  
**Reject**: <code>[BadRequest](#module_atlas-api-client--AtlasAPIClient.Errors.BadRequest)</code>  
**Reject**: <code>[Unauthorized](#module_atlas-api-client--AtlasAPIClient.Errors.Unauthorized)</code>  
**Reject**: <code>[NotFound](#module_atlas-api-client--AtlasAPIClient.Errors.NotFound)</code>  
**Reject**: <code>[InternalError](#module_atlas-api-client--AtlasAPIClient.Errors.InternalError)</code>  
**Reject**: <code>Error</code>  

| Param | Type | Description |
| --- | --- | --- |
| params | <code>Object</code> |  |
| params.groupID | <code>string</code> |  |
| params.host | <code>string</code> |  |
| params.port | <code>number</code> |  |
| params.diskName | <code>string</code> |  |
| params.granularity | <code>string</code> |  |
| [params.period] | <code>string</code> |  |
| [params.start] | <code>string</code> |  |
| [params.end] | <code>string</code> |  |
| [params.m] | <code>Array.&lt;string&gt;</code> |  |
| [params.pageNum] | <code>number</code> |  |
| [params.itemsPerPage] | <code>number</code> |  |
| [options] | <code>object</code> |  |
| [options.timeout] | <code>number</code> | A request specific timeout |
| [options.span] | <code>[Span](https://doc.esdoc.org/github.com/opentracing/opentracing-javascript/class/src/span.js~Span.html)</code> | An OpenTracing span - For example from the parent request |
| [options.retryPolicy] | <code>[RetryPolicies](#module_atlas-api-client--AtlasAPIClient.RetryPolicies)</code> | A request specific retryPolicy |
| [cb] | <code>function</code> |  |

<a name="module_atlas-api-client--AtlasAPIClient+getProcessMeasurements"></a>

#### atlasAPIClient.getProcessMeasurements(params, [options], [cb]) ⇒ <code>Promise</code>
Get measurements for a specific Atlas MongoDB process (mongod or mongos).

**Kind**: instance method of <code>[AtlasAPIClient](#exp_module_atlas-api-client--AtlasAPIClient)</code>  
**Fulfill**: <code>Object</code>  
**Reject**: <code>[BadRequest](#module_atlas-api-client--AtlasAPIClient.Errors.BadRequest)</code>  
**Reject**: <code>[Unauthorized](#module_atlas-api-client--AtlasAPIClient.Errors.Unauthorized)</code>  
**Reject**: <code>[NotFound](#module_atlas-api-client--AtlasAPIClient.Errors.NotFound)</code>  
**Reject**: <code>[InternalError](#module_atlas-api-client--AtlasAPIClient.Errors.InternalError)</code>  
**Reject**: <code>Error</code>  

| Param | Type | Description |
| --- | --- | --- |
| params | <code>Object</code> |  |
| params.groupID | <code>string</code> |  |
| params.host | <code>string</code> |  |
| params.port | <code>number</code> |  |
| params.granularity | <code>string</code> |  |
| [params.period] | <code>string</code> |  |
| [params.start] | <code>string</code> |  |
| [params.end] | <code>string</code> |  |
| [params.m] | <code>Array.&lt;string&gt;</code> |  |
| [params.pageNum] | <code>number</code> |  |
| [params.itemsPerPage] | <code>number</code> |  |
| [options] | <code>object</code> |  |
| [options.timeout] | <code>number</code> | A request specific timeout |
| [options.span] | <code>[Span](https://doc.esdoc.org/github.com/opentracing/opentracing-javascript/class/src/span.js~Span.html)</code> | An OpenTracing span - For example from the parent request |
| [options.retryPolicy] | <code>[RetryPolicies](#module_atlas-api-client--AtlasAPIClient.RetryPolicies)</code> | A request specific retryPolicy |
| [cb] | <code>function</code> |  |

<a name="module_atlas-api-client--AtlasAPIClient.RetryPolicies"></a>

#### AtlasAPIClient.RetryPolicies
Retry policies available to use.

**Kind**: static property of <code>[AtlasAPIClient](#exp_module_atlas-api-client--AtlasAPIClient)</code>  

* [.RetryPolicies](#module_atlas-api-client--AtlasAPIClient.RetryPolicies)
    * [.Exponential](#module_atlas-api-client--AtlasAPIClient.RetryPolicies.Exponential)
    * [.Single](#module_atlas-api-client--AtlasAPIClient.RetryPolicies.Single)
    * [.None](#module_atlas-api-client--AtlasAPIClient.RetryPolicies.None)

<a name="module_atlas-api-client--AtlasAPIClient.RetryPolicies.Exponential"></a>

##### RetryPolicies.Exponential
The exponential retry policy will retry five times with an exponential backoff.

**Kind**: static constant of <code>[RetryPolicies](#module_atlas-api-client--AtlasAPIClient.RetryPolicies)</code>  
<a name="module_atlas-api-client--AtlasAPIClient.RetryPolicies.Single"></a>

##### RetryPolicies.Single
Use this retry policy to retry a request once.

**Kind**: static constant of <code>[RetryPolicies](#module_atlas-api-client--AtlasAPIClient.RetryPolicies)</code>  
<a name="module_atlas-api-client--AtlasAPIClient.RetryPolicies.None"></a>

##### RetryPolicies.None
Use this retry policy to turn off retries.

**Kind**: static constant of <code>[RetryPolicies](#module_atlas-api-client--AtlasAPIClient.RetryPolicies)</code>  
<a name="module_atlas-api-client--AtlasAPIClient.Errors"></a>

#### AtlasAPIClient.Errors
Errors returned by methods.

**Kind**: static property of <code>[AtlasAPIClient](#exp_module_atlas-api-client--AtlasAPIClient)</code>  

* [.Errors](#module_atlas-api-client--AtlasAPIClient.Errors)
    * [.BadRequest](#module_atlas-api-client--AtlasAPIClient.Errors.BadRequest) ⇐ <code>Error</code>
    * [.Unauthorized](#module_atlas-api-client--AtlasAPIClient.Errors.Unauthorized) ⇐ <code>Error</code>
    * [.NotFound](#module_atlas-api-client--AtlasAPIClient.Errors.NotFound) ⇐ <code>Error</code>
    * [.InternalError](#module_atlas-api-client--AtlasAPIClient.Errors.InternalError) ⇐ <code>Error</code>

<a name="module_atlas-api-client--AtlasAPIClient.Errors.BadRequest"></a>

##### Errors.BadRequest ⇐ <code>Error</code>
BadRequest

**Kind**: static class of <code>[Errors](#module_atlas-api-client--AtlasAPIClient.Errors)</code>  
**Extends:** <code>Error</code>  
**Properties**

| Name | Type |
| --- | --- |
| detail | <code>string</code> | 
| error | <code>number</code> | 
| message | <code>string</code> | 
| reason | <code>string</code> | 

<a name="module_atlas-api-client--AtlasAPIClient.Errors.Unauthorized"></a>

##### Errors.Unauthorized ⇐ <code>Error</code>
Unauthorized

**Kind**: static class of <code>[Errors](#module_atlas-api-client--AtlasAPIClient.Errors)</code>  
**Extends:** <code>Error</code>  
**Properties**

| Name | Type |
| --- | --- |
| detail | <code>string</code> | 
| error | <code>number</code> | 
| message | <code>string</code> | 
| reason | <code>string</code> | 

<a name="module_atlas-api-client--AtlasAPIClient.Errors.NotFound"></a>

##### Errors.NotFound ⇐ <code>Error</code>
NotFound

**Kind**: static class of <code>[Errors](#module_atlas-api-client--AtlasAPIClient.Errors)</code>  
**Extends:** <code>Error</code>  
**Properties**

| Name | Type |
| --- | --- |
| detail | <code>string</code> | 
| error | <code>number</code> | 
| message | <code>string</code> | 
| reason | <code>string</code> | 

<a name="module_atlas-api-client--AtlasAPIClient.Errors.InternalError"></a>

##### Errors.InternalError ⇐ <code>Error</code>
InternalError

**Kind**: static class of <code>[Errors](#module_atlas-api-client--AtlasAPIClient.Errors)</code>  
**Extends:** <code>Error</code>  
**Properties**

| Name | Type |
| --- | --- |
| detail | <code>string</code> | 
| error | <code>number</code> | 
| message | <code>string</code> | 
| reason | <code>string</code> | 

<a name="module_atlas-api-client--AtlasAPIClient.DefaultCircuitOptions"></a>

#### AtlasAPIClient.DefaultCircuitOptions
Default circuit breaker options.

**Kind**: static constant of <code>[AtlasAPIClient](#exp_module_atlas-api-client--AtlasAPIClient)</code>  
<a name="responseLog"></a>

## responseLog()
Request status log is used to
to output the status of a request returned
by the client.

**Kind**: global function  
