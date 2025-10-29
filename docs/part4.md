* If "autoCloseType" is not sent, orders with both of the types will be returned
* If "startTime" is not sent, data within 7 days before "endTime" can be queried



## User Commission Rate (USER_DATA)

> **Response:**

```javascript
{
	"symbol": "BTCUSDT",
  	"makerCommissionRate": "0.0002",  // 0.02%
  	"takerCommissionRate": "0.0004"   // 0.04%
}
```

``
GET /fapi/v1/commissionRate (HMAC SHA256)
``

**Weight:**
20


**Parameters:**

Name | Type | Mandatory | Description
------------ | ------------ | ------------ | ------------
symbol | STRING | YES	
recvWindow | LONG | NO	
timestamp | LONG | YES





# User Data Streams

* The base API endpoint is: **https://fapi.asterdex.com**
* A User Data Stream `listenKey` is valid for 60 minutes after creation.
* Doing a `PUT` on a `listenKey` will extend its validity for 60 minutes.
* Doing a `DELETE` on a `listenKey` will close the stream and invalidate the `listenKey`.
* Doing a `POST` on an account with an active `listenKey` will return the currently active `listenKey` and extend its validity for 60 minutes.
* The baseurl for websocket is **wss://fstream.asterdex.com**
* User Data Streams are accessed at **/ws/\<listenKey\>**
* User data stream payloads are **not guaranteed** to be in order during heavy periods; **make sure to order your updates using E**
* A single connection to **fstream.asterdex.com** is only valid for 24 hours; expect to be disconnected at the 24 hour mark


## Start User Data Stream (USER_STREAM)


> **Response:**

```javascript
{
  "listenKey": "pqia91ma19a5s61cv6a81va65sdf19v8a65a1a5s61cv6a81va65sdf19v8a65a1"
}
```

``
POST /fapi/v1/listenKey
``

Start a new user data stream. The stream will close after 60 minutes unless a keepalive is sent. If the account has an active `listenKey`, that `listenKey` will be returned and its validity will be extended for 60 minutes.

**Weight:**
1

**Parameters:**

None



## Keepalive User Data Stream (USER_STREAM)

> **Response:**

```javascript
{}
```

``
PUT /fapi/v1/listenKey
``

Keepalive a user data stream to prevent a time out. User data streams will close after 60 minutes. It's recommended to send a ping about every 60 minutes.

**Weight:**
1

**Parameters:**

None



## Close User Data Stream (USER_STREAM)


> **Response:**

```javascript
{}
```

``
DELETE /fapi/v1/listenKey
``

Close out a user data stream.

**Weight:**
1

**Parameters:**

None


## Event: User Data Stream Expired

> **Payload:**

```javascript
{
	'e': 'listenKeyExpired',      // event type
	'E': 1576653824250				// event time
}
```

When the `listenKey` used for the user data stream turns expired, this event will be pushed.

**Notice:**

