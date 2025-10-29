- [General Info](#general-info)
    - [General API Information](#general-api-information)
        - [HTTP Return Codes](#http-return-codes)
        - [Error Codes and Messages](#error-codes-and-messages)
        - [General Information on Endpoints](#general-information-on-endpoints)
    - [LIMITS](#limits)
        - [IP Limits](#ip-limits)
        - [Order Rate Limits](#order-rate-limits)
    - [Endpoint Security Type](#endpoint-security-type)
    - [SIGNED (TRADE and USER_DATA) Endpoint Security](#signed-trade-and-user_data-endpoint-security)
        - [Timing Security](#timing-security)
        - [SIGNED Endpoint Examples for POST /fapi/v1/order](#signed-endpoint-examples-for-post-fapiv1order)
            - [Example 1: As a query string](#example-1-as-a-query-string)
            - [Example 2: As a request body](#example-2-as-a-request-body)
            - [Example 3: Mixed query string and request body](#example-3-mixed-query-string-and-request-body)
    - [Public Endpoints Info](#public-endpoints-info)
        - [Terminology](#terminology)
        - [ENUM definitions](#enum-definitions)
    - [Filters](#filters)
        - [Symbol filters](#symbol-filters)
            - [PRICE_FILTER](#price_filter)
            - [LOT_SIZE](#lot_size)
            - [MARKET_LOT_SIZE](#market_lot_size)
            - [MAX_NUM_ORDERS](#max_num_orders)
            - [MAX_NUM_ALGO_ORDERS](#max_num_algo_orders)
            - [PERCENT_PRICE](#percent_price)
            - [MIN_NOTIONAL](#min_notional)
- [Market Data Endpoints](#market-data-endpoints)
    - [Test Connectivity](#test-connectivity)
    - [Check Server Time](#check-server-time)
    - [Exchange Information](#exchange-information)
    - [Order Book](#order-book)
    - [Recent Trades List](#recent-trades-list)
    - [Old Trades Lookup (MARKET_DATA)](#old-trades-lookup-market_data)
    - [Compressed/Aggregate Trades List](#compressedaggregate-trades-list)
    - [Kline/Candlestick Data](#klinecandlestick-data)
    - [Index Price Kline/Candlestick Data](#index-price-klinecandlestick-data)
    - [Mark Price Kline/Candlestick Data](#mark-price-klinecandlestick-data)
    - [Mark Price](#mark-price)
    - [Get Funding Rate History](#get-funding-rate-history)
    - [Get Funding Rate Config](#get-funding-rate-config)
    - [24hr Ticker Price Change Statistics](#24hr-ticker-price-change-statistics)
    - [Symbol Price Ticker](#symbol-price-ticker)
    - [Symbol Order Book Ticker](#symbol-order-book-ticker)
- [Websocket Market Streams](#websocket-market-streams)
    - [Live Subscribing/Unsubscribing to streams](#live-subscribingunsubscribing-to-streams)
        - [Subscribe to a stream](#subscribe-to-a-stream)
        - [Unsubscribe to a stream](#unsubscribe-to-a-stream)
        - [Listing Subscriptions](#listing-subscriptions)
        - [Setting Properties](#setting-properties)
        - [Retrieving Properties](#retrieving-properties)
        - [Error Messages](#error-messages)
    - [Aggregate Trade Streams](#aggregate-trade-streams)
    - [Mark Price Stream](#mark-price-stream)
    - [Mark Price Stream for All market](#mark-price-stream-for-all-market)
    - [Kline/Candlestick Streams](#klinecandlestick-streams)
    - [Individual Symbol Mini Ticker Stream](#individual-symbol-mini-ticker-stream)
    - [All Market Mini Tickers Stream](#all-market-mini-tickers-stream)
    - [Individual Symbol Ticker Streams](#individual-symbol-ticker-streams)
    - [All Market Tickers Streams](#all-market-tickers-streams)
    - [Individual Symbol Book Ticker Streams](#individual-symbol-book-ticker-streams)
    - [All Book Tickers Stream](#all-book-tickers-stream)
    - [Liquidation Order Streams](#liquidation-order-streams)
    - [All Market Liquidation Order Streams](#all-market-liquidation-order-streams)
    - [Partial Book Depth Streams](#partial-book-depth-streams)
    - [Diff. Book Depth Streams](#diff-book-depth-streams)
    - [How to manage a local order book correctly](#how-to-manage-a-local-order-book-correctly)
- [Account/Trades Endpoints](#accounttrades-endpoints)
    - [Change Position Mode(TRADE)](#change-position-modetrade)
    - [Get Current Position Mode(USER_DATA)](#get-current-position-modeuser_data)
    - [Change Multi-Assets Mode (TRADE)](#change-multi-assets-mode-trade)
    - [Get Current Multi-Assets Mode (USER_DATA)](#get-current-multi-assets-mode-user_data)
    - [New Order  (TRADE)](#new-order--trade)
    - [Place Multiple Orders  (TRADE)](#place-multiple-orders--trade)
    - [Transfer Between Futures And Spot (USER_DATA)](#transfer-between-futures-and-spot-user_data)
    - [Query Order (USER_DATA)](#query-order-user_data)
    - [Cancel Order (TRADE)](#cancel-order-trade)
    - [Cancel All Open Orders (TRADE)](#cancel-all-open-orders-trade)
    - [Cancel Multiple Orders (TRADE)](#cancel-multiple-orders-trade)
    - [Auto-Cancel All Open Orders (TRADE)](#auto-cancel-all-open-orders-trade)
    - [Query Current Open Order (USER_DATA)](#query-current-open-order-user_data)
    - [Current All Open Orders (USER_DATA)](#current-all-open-orders-user_data)
    - [All Orders (USER_DATA)](#all-orders-user_data)
    - [Futures Account Balance V2 (USER_DATA)](#futures-account-balance-v2-user_data)
    - [Account Information V2 (USER_DATA)](#account-information-v2-user_data)
    - [Change Initial Leverage (TRADE)](#change-initial-leverage-trade)
    - [Change Margin Type (TRADE)](#change-margin-type-trade)
    - [Modify Isolated Position Margin (TRADE)](#modify-isolated-position-margin-trade)
    - [Get Position Margin Change History (TRADE)](#get-position-margin-change-history-trade)
    - [Position Information V2 (USER_DATA)](#position-information-v2-user_data)
    - [Account Trade List (USER_DATA)](#account-trade-list-user_data)
    - [Get Income History(USER_DATA)](#get-income-historyuser_data)
    - [Notional and Leverage Brackets (USER_DATA)](#notional-and-leverage-brackets-user_data)
    - [Position ADL Quantile Estimation (USER_DATA)](#position-adl-quantile-estimation-user_data)
    - [User's Force Orders (USER_DATA)](#users-force-orders-user_data)
    - [User Commission Rate (USER_DATA)](#user-commission-rate-user_data)
- [User Data Streams](#user-data-streams)
    - [Start User Data Stream (USER_STREAM)](#start-user-data-stream-user_stream)
    - [Keepalive User Data Stream (USER_STREAM)](#keepalive-user-data-stream-user_stream)
    - [Close User Data Stream (USER_STREAM)](#close-user-data-stream-user_stream)
    - [Event: User Data Stream Expired](#event-user-data-stream-expired)
    - [Event: Margin Call](#event-margin-call)
    - [Event: Balance and Position Update](#event-balance-and-position-update)
    - [Event: Order Update](#event-order-update)
    - [Event: Account Configuration Update previous Leverage Update](#event-account-configuration-update-previous-leverage-update)
- [Error Codes](#error-codes)
    - [10xx - General Server or Network issues](#10xx---general-server-or-network-issues)
    - [11xx - Request issues](#11xx---request-issues)
    - [20xx - Processing Issues](#20xx---processing-issues)
    - [40xx - Filters and other Issues](#40xx---filters-and-other-issues)

# General Info

## General API Information

* Some endpoints will require an API Key. Please refer to [this page](https://www.asterdex.com/)
* The base endpoint is: **https://fapi.asterdex.com**
* All endpoints return either a JSON object or array.
* Data is returned in **ascending** order. Oldest first, newest last.
* All time and timestamp related fields are in milliseconds.
* All data types adopt definition in JAVA.

### HTTP Return Codes

* HTTP `4XX` return codes are used for for malformed requests;
  the issue is on the sender's side.
* HTTP `403` return code is used when the WAF Limit (Web Application Firewall) has been violated.
* HTTP `429` return code is used when breaking a request rate limit.
* HTTP `418` return code is used when an IP has been auto-banned for continuing to send requests after receiving `429` codes.
* HTTP `5XX` return codes are used for internal errors; the issue is on
  Aster's side.
* HTTP `503` return code is used when the API successfully sent the message but not get a response within the timeout period.   
  It is important to **NOT** treat this as a failure operation; the execution status is
  **UNKNOWN** and could have been a success.

### Error Codes and Messages

* Any endpoint can return an ERROR

> ***The error payload is as follows:***

```javascript
{
  "code": -1121,
  "msg": "Invalid symbol."
}
```

* Specific error codes and messages defined in [Error Codes](#error-codes).

### General Information on Endpoints

* For `GET` endpoints, parameters must be sent as a `query string`.
* For `POST`, `PUT`, and `DELETE` endpoints, the parameters may be sent as a
  `query string` or in the `request body` with content type
  `application/x-www-form-urlencoded`. You may mix parameters between both the
  `query string` and `request body` if you wish to do so.
* Parameters may be sent in any order.
* If a parameter sent in both the `query string` and `request body`, the
  `query string` parameter will be used.

## LIMITS
* The `/fapi/v1/exchangeInfo` `rateLimits` array contains objects related to the exchange's `RAW_REQUEST`, `REQUEST_WEIGHT`, and `ORDER` rate limits. These are further defined in the `ENUM definitions` section under `Rate limiters (rateLimitType)`.
* A `429` will be returned when either rate limit is violated.

<aside class="notice">
Aster Finance has the right to further tighten the rate limits on users with intent to attack.
</aside>

### IP Limits
* Every request will contain `X-MBX-USED-WEIGHT-(intervalNum)(intervalLetter)` in the response headers which has the current used weight for the IP for all request rate limiters defined.
* Each route has a `weight` which determines for the number of requests each endpoint counts for. Heavier endpoints and endpoints that do operations on multiple symbols will have a heavier `weight`.
* When a 429 is received, it's your obligation as an API to back off and not spam the API.
* **Repeatedly violating rate limits and/or failing to back off after receiving 429s will result in an automated IP ban (HTTP status 418).**
* IP bans are tracked and **scale in duration** for repeat offenders, **from 2 minutes to 3 days**.
* **The limits on the API are based on the IPs, not the API keys.**

<aside class="notice">
It is strongly recommended to use websocket stream for getting data as much as possible, which can not only ensure the timeliness of the message, but also reduce the access restriction pressure caused by the request.
</aside>

### Order Rate Limits
* Every order response will contain a `X-MBX-ORDER-COUNT-(intervalNum)(intervalLetter)` header which has the current order count for the account for all order rate limiters defined.
* Rejected/unsuccessful orders are not guaranteed to have `X-MBX-ORDER-COUNT-**` headers in the response.
* **The order rate limit is counted against each account**.

## Endpoint Security Type
* Each endpoint has a security type that determines the how you will
  interact with it.
* API-keys are passed into the Rest API via the `X-MBX-APIKEY`
  header.
* API-keys and secret-keys **are case sensitive**.
* API-keys can be configured to only access certain types of secure endpoints.
  For example, one API-key could be used for TRADE only, while another API-key
  can access everything except for TRADE routes.
* By default, API-keys can access all secure routes.

Security Type | Description
------------ | ------------
NONE | Endpoint can be accessed freely.
TRADE | Endpoint requires sending a valid API-Key and signature.
USER_DATA | Endpoint requires sending a valid API-Key and signature.
USER_STREAM | Endpoint requires sending a valid API-Key.
MARKET_DATA | Endpoint requires sending a valid API-Key.


* `TRADE` and `USER_DATA` endpoints are `SIGNED` endpoints.

## SIGNED (TRADE and USER_DATA) Endpoint Security
* `SIGNED` endpoints require an additional parameter, `signature`, to be
  sent in the  `query string` or `request body`.
* Endpoints use `HMAC SHA256` signatures. The `HMAC SHA256 signature` is a keyed `HMAC SHA256` operation.
  Use your `secretKey` as the key and `totalParams` as the value for the HMAC operation.
* The `signature` is **not case sensitive**.
* Please make sure the `signature` is the end part of your `query string` or `request body`.
* `totalParams` is defined as the `query string` concatenated with the
  `request body`.

### Timing Security
* A `SIGNED` endpoint also requires a parameter, `timestamp`, to be sent which
  should be the millisecond timestamp of when the request was created and sent.
* An additional parameter, `recvWindow`, may be sent to specify the number of
  milliseconds after `timestamp` the request is valid for. If `recvWindow`
  is not sent, **it defaults to 5000**.

> The logic is as follows:

```javascript
  if (timestamp < (serverTime + 1000) && (serverTime - timestamp) <= recvWindow){
    // process request
  } 
  else {
    // reject request
  }
```

**Serious trading is about timing.** Networks can be unstable and unreliable,
which can lead to requests taking varying amounts of time to reach the
servers. With `recvWindow`, you can specify that the request must be
processed within a certain number of milliseconds or be rejected by the
server.

<aside class="notice">
It is recommended to use a small recvWindow of 5000 or less!
</aside>



``
GET /fapi/v1/positionMargin/history (HMAC SHA256)
``

**Weight:**
1

**Parameters:**

Name | Type | Mandatory | Description
------------ | ------------ | ------------ | ------------
symbol | STRING | YES	
type | INT	 | NO | 1: Add position margin，2: Reduce position margin
startTime | LONG | NO	
endTime | LONG | NO	
limit | INT | NO | Default: 500
recvWindow | LONG | NO	
timestamp | LONG | YES	






## Position Information V2 (USER_DATA)


> **Response:**

> For One-way position mode:

```javascript
[
  	{
  		"entryPrice": "0.00000",
  		"marginType": "isolated", 
  		"isAutoAddMargin": "false",
  		"isolatedMargin": "0.00000000",	
  		"leverage": "10", 
  		"liquidationPrice": "0", 
  		"markPrice": "6679.50671178",	
  		"maxNotionalValue": "20000000", 
  		"positionAmt": "0.000", 
  		"symbol": "BTCUSDT", 
  		"unRealizedProfit": "0.00000000", 
  		"positionSide": "BOTH",
  		"updateTime": 0
  	}
]
```

> For Hedge position mode:

```javascript
[
  	{
  		"entryPrice": "6563.66500", 
  		"marginType": "isolated", 
  		"isAutoAddMargin": "false",
  		"isolatedMargin": "15517.54150468",
  		"leverage": "10",
  		"liquidationPrice": "5930.78",
  		"markPrice": "6679.50671178",	
  		"maxNotionalValue": "20000000", 
  		"positionAmt": "20.000", 
  		"symbol": "BTCUSDT", 
  		"unRealizedProfit": "2316.83423560"
  		"positionSide": "LONG", 
  		"updateTime": 1625474304765
  	},
  	{
  		"entryPrice": "0.00000",
  		"marginType": "isolated", 
  		"isAutoAddMargin": "false",
  		"isolatedMargin": "5413.95799991", 
  		"leverage": "10", 
  		"liquidationPrice": "7189.95", 
  		"markPrice": "6679.50671178",	
  		"maxNotionalValue": "20000000", 
  		"positionAmt": "-10.000", 
  		"symbol": "BTCUSDT",
  		"unRealizedProfit": "-1156.46711780" 
  		"positionSide": "SHORT",
  		"updateTime": 0
  	}
]
```

``
GET /fapi/v2/positionRisk (HMAC SHA256)
``

Get current position information.

**Weight:**
5

**Parameters:**

Name | Type | Mandatory | Description
------------ | ------------ | ------------ | ------------
symbol | STRING | NO
recvWindow | LONG | NO |
timestamp | LONG | YES |

**Note**    
Please use with user data stream `ACCOUNT_UPDATE` to meet your timeliness and accuracy needs.



## Account Trade List (USER_DATA)


> **Response:**

```javascript
[
  {
  	"buyer": false,
  	"commission": "-0.07819010",
  	"commissionAsset": "USDT",
  	"id": 698759,
  	"maker": false,
  	"orderId": 25851813,
  	"price": "7819.01",
  	"qty": "0.002",
  	"quoteQty": "15.63802",
  	"realizedPnl": "-0.91539999",
  	"side": "SELL",
  	"positionSide": "SHORT",
  	"symbol": "BTCUSDT",
  	"time": 1569514978020
  }
]
```

``
GET /fapi/v1/userTrades  (HMAC SHA256)
``

Get trades for a specific account and symbol.

**Weight:**
5

**Parameters:**

Name | Type | Mandatory | Description
------------ | ------------ | ------------ | ------------
symbol | STRING | YES |
startTime | LONG | NO |
endTime | LONG | NO |
fromId | LONG | NO | Trade id to fetch from. Default gets most recent trades.
limit | INT | NO | Default 500; max 1000.
recvWindow | LONG | NO |
timestamp | LONG | YES |

* If `startTime` and `endTime` are both not sent, then the last 7 days' data will be returned.
* The time between `startTime` and `endTime` cannot be longer than 7 days.
* The parameter `fromId` cannot be sent with `startTime` or `endTime`.


## Get Income History(USER_DATA)


> **Response:**

```javascript
[
	{
    	"symbol": "",					// trade symbol, if existing
    	"incomeType": "TRANSFER",	// income type
    	"income": "-0.37500000",  // income amount
    	"asset": "USDT",				// income asset
    	"info":"TRANSFER",			// extra information
    	"time": 1570608000000,		
    	"tranId":"9689322392",		// transaction id
    	"tradeId":""					// trade id, if existing
	},
	{
   		"symbol": "BTCUSDT",
    	"incomeType": "COMMISSION", 
    	"income": "-0.01000000",
    	"asset": "USDT",
    	"info":"COMMISSION",
    	"time": 1570636800000,
    	"tranId":"9689322392",
    	"tradeId":"2059192"
	}
]
```

``
GET /fapi/v1/income (HMAC SHA256)
``

**Weight:**
30

**Parameters:**

Name | Type | Mandatory | Description
------------ | ------------ | ------------ | ------------
symbol | STRING | NO|
incomeType | STRING | NO | "TRANSFER"，"WELCOME_BONUS", "REALIZED_PNL"，"FUNDING_FEE", "COMMISSION", "INSURANCE_CLEAR", and "MARKET_MERCHANT_RETURN_REWARD"
startTime | LONG | NO | Timestamp in ms to get funding from INCLUSIVE.
endTime | LONG | NO | Timestamp in ms to get funding until INCLUSIVE.
limit | INT | NO | Default 100; max 1000
recvWindow|LONG|NO|
timestamp|LONG|YES|

* If neither `startTime` nor `endTime` is sent, the recent 7-day data will be returned.
* If `incomeType ` is not sent, all kinds of flow will be returned
* "trandId" is unique in the same incomeType for a user


## Notional and Leverage Brackets (USER_DATA)


> **Response:**

```javascript
[
    {
        "symbol": "ETHUSDT",
        "brackets": [
            {
                "bracket": 1,   // Notional bracket
                "initialLeverage": 75,  // Max initial leverage for this bracket
                "notionalCap": 10000,  // Cap notional of this bracket
                "notionalFloor": 0,  // Notional threshold of this bracket 
                "maintMarginRatio": 0.0065, // Maintenance ratio for this bracket
                "cum":0 // Auxiliary number for quick calculation 
               
            },
        ]
    }
]
```

> **OR** (if symbol sent)

```javascript

{
    "symbol": "ETHUSDT",
    "brackets": [
        {
            "bracket": 1,
            "initialLeverage": 75,
            "notionalCap": 10000,
            "notionalFloor": 0,
            "maintMarginRatio": 0.0065,
            "cum":0
        },
    ]
}
```


``
GET /fapi/v1/leverageBracket
``


**Weight:** 1

**Parameters:**

Name | Type | Mandatory | Description
------------ | ------------ | ------------ | ------------
symbol	| STRING | NO
recvWindow|LONG|NO|
timestamp|LONG|YES|



## Position ADL Quantile Estimation (USER_DATA)


> **Response:**

```javascript
[
	{
		"symbol": "ETHUSDT", 
		"adlQuantile": 
			{
				// if the positions of the symbol are crossed margined in Hedge Mode, "LONG" and "SHORT" will be returned a same quantile value, and "HEDGE" will be returned instead of "BOTH".
				"LONG": 3,  
				"SHORT": 3, 
				"HEDGE": 0   // only a sign, ignore the value
			}
		},
 	{
 		"symbol": "BTCUSDT", 
 		"adlQuantile": 
 			{
 				// for positions of the symbol are in One-way Mode or isolated margined in Hedge Mode
 				"LONG": 1, 	// adl quantile for "LONG" position in hedge mode
 				"SHORT": 2, 	// adl qauntile for "SHORT" position in hedge mode
 				"BOTH": 0		// adl qunatile for position in one-way mode
 			}
 	}
 ]
```

``
GET /fapi/v1/adlQuantile
``


**Weight:** 5

**Parameters:**

Name | Type | Mandatory | Description
------------ | ------------ | ------------ | ------------
symbol	| STRING | NO
recvWindow|LONG|NO|
timestamp|LONG|YES|

* Values update every 30s.

* Values 0, 1, 2, 3, 4 shows the queue position and possibility of ADL from low to high.

* For positions of the symbol are in One-way Mode or isolated margined in Hedge Mode, "LONG", "SHORT", and "BOTH" will be returned to show the positions' adl quantiles of different position sides.

* If the positions of the symbol are crossed margined in Hedge Mode:
    * "HEDGE" as a sign will be returned instead of "BOTH";
    * A same value caculated on unrealized pnls on long and short sides' positions will be shown for "LONG" and "SHORT" when there are positions in both of long and short sides.



## User's Force Orders (USER_DATA)


> **Response:**

```javascript
[
  {
  	"orderId": 6071832819, 
  	"symbol": "BTCUSDT", 
  	"status": "FILLED", 
  	"clientOrderId": "autoclose-1596107620040000020", 
  	"price": "10871.09", 
  	"avgPrice": "10913.21000", 
  	"origQty": "0.001", 
  	"executedQty": "0.001", 
  	"cumQuote": "10.91321", 
  	"timeInForce": "IOC", 
  	"type": "LIMIT", 
  	"reduceOnly": false, 
  	"closePosition": false, 
  	"side": "SELL", 
  	"positionSide": "BOTH", 
  	"stopPrice": "0", 
  	"workingType": "CONTRACT_PRICE", 
  	"origType": "LIMIT", 
  	"time": 1596107620044, 
  	"updateTime": 1596107620087
  }
  {
   	"orderId": 6072734303, 
   	"symbol": "BTCUSDT", 
   	"status": "FILLED", 
   	"clientOrderId": "adl_autoclose", 
   	"price": "11023.14", 
   	"avgPrice": "10979.82000", 
   	"origQty": "0.001", 
   	"executedQty": "0.001", 
   	"cumQuote": "10.97982", 
   	"timeInForce": "GTC", 
   	"type": "LIMIT", 
   	"reduceOnly": false, 
   	"closePosition": false, 
   	"side": "BUY", 
   	"positionSide": "SHORT", 
   	"stopPrice": "0", 
   	"workingType": "CONTRACT_PRICE", 
   	"origType": "LIMIT", 
   	"time": 1596110725059, 
   	"updateTime": 1596110725071
  }
]
```


``
GET /fapi/v1/forceOrders
``


**Weight:** 20 with symbol, 50 without symbol

**Parameters:**

  Name      |  Type  | Mandatory |                         Description
------------- | ------ | --------- | -----------------------------------------------------------
symbol        | STRING | NO        |
autoCloseType | ENUM   | NO        | "LIQUIDATION" for liquidation orders, "ADL" for ADL orders.
startTime     | LONG   | NO        |
endTime       | LONG   | NO        |
limit         | INT    | NO        | Default 50; max 100.
recvWindow    | LONG   | NO        |
timestamp     | LONG   | YES       |

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

* This event is not related to the websocket disconnection.
* This event will be received only when a valid `listenKey` in connection got expired.
* No more user data event will be updated after this event received until a new valid `listenKey` used.





## Event: Margin Call

> **Payload:**

```javascript
{
    "e":"MARGIN_CALL",    	// Event Type
    "E":1587727187525,		// Event Time
    "cw":"3.16812045",		// Cross Wallet Balance. Only pushed with crossed position margin call
    "p":[					// Position(s) of Margin Call
      {
        "s":"ETHUSDT",		// Symbol
        "ps":"LONG",		// Position Side
        "pa":"1.327",		// Position Amount
        "mt":"CROSSED",		// Margin Type
        "iw":"0",			// Isolated Wallet (if isolated position)
        "mp":"187.17127",	// Mark Price
        "up":"-1.166074",	// Unrealized PnL
        "mm":"1.614445"		// Maintenance Margin Required
      }
    ]
}  
 
```


* When the user's position risk ratio is too high, this stream will be pushed.
* This message is only used as risk guidance information and is not recommended for investment strategies.
* In the case of a highly volatile market, there may be the possibility that the user's position has been liquidated at the same time when this stream is pushed out.





## Event: Balance and Position Update


> **Payload:**

```javascript
{
  "e": "ACCOUNT_UPDATE",				// Event Type
  "E": 1564745798939,            		// Event Time
  "T": 1564745798938 ,           		// Transaction
  "a":                          		// Update Data
    {
      "m":"ORDER",						// Event reason type
      "B":[                     		// Balances
        {
          "a":"USDT",           		// Asset
          "wb":"122624.12345678",    	// Wallet Balance
          "cw":"100.12345678",			// Cross Wallet Balance
          "bc":"50.12345678"			// Balance Change except PnL and Commission
        },
        {
          "a":"BUSD",           
          "wb":"1.00000000",
          "cw":"0.00000000",         
          "bc":"-49.12345678"
        }
      ],
      "P":[
        {
          "s":"BTCUSDT",          	// Symbol
          "pa":"0",               	// Position Amount
          "ep":"0.00000",            // Entry Price
          "cr":"200",             	// (Pre-fee) Accumulated Realized
          "up":"0",						// Unrealized PnL
          "mt":"isolated",				// Margin Type
          "iw":"0.00000000",			// Isolated Wallet (if isolated position)
          "ps":"BOTH"					// Position Side
        }，
        {
        	"s":"BTCUSDT",
        	"pa":"20",
        	"ep":"6563.66500",
        	"cr":"0",
        	"up":"2850.21200",
        	"mt":"isolated",
        	"iw":"13200.70726908",
        	"ps":"LONG"
      	 },
        {
        	"s":"BTCUSDT",
        	"pa":"-10",
        	"ep":"6563.86000",
        	"cr":"-45.04000000",
        	"up":"-1423.15600",
        	"mt":"isolated",
        	"iw":"6570.42511771",
        	"ps":"SHORT"
        }
      ]
    }
}
```

Event type is `ACCOUNT_UPDATE`.

* When balance or position get updated, this event will be pushed.
    * `ACCOUNT_UPDATE` will be pushed only when update happens on user's account, including changes on balances, positions, or margin type.
    * Unfilled orders or cancelled orders will not make the event `ACCOUNT_UPDATE` pushed, since there's no change on positions.
    * Only positions of symbols with non-zero isolatd wallet or non-zero position amount will be pushed in the "position" part of the event `ACCOUNT_UPDATE` when any position changes.

* When "FUNDING FEE" changes to the user's balance, the event will be pushed with the brief message:
    * When "FUNDING FEE" occurs in a **crossed position**, `ACCOUNT_UPDATE` will be pushed with only the balance `B`(including the "FUNDING FEE" asset only), without any position `P` message.
    * When "FUNDING FEE" occurs in an **isolated position**, `ACCOUNT_UPDATE` will be pushed with only the balance `B`(including the "FUNDING FEE" asset only) and the relative position message `P`( including the isolated position on which the "FUNDING FEE" occurs only, without any other position message).

* The field "m" represents the reason type for the event and may shows the following possible types:
    * DEPOSIT
    * WITHDRAW
    * ORDER
    * FUNDING_FEE
    * WITHDRAW_REJECT
    * ADJUSTMENT
    * INSURANCE_CLEAR
    * ADMIN_DEPOSIT
    * ADMIN_WITHDRAW
    * MARGIN_TRANSFER
    * MARGIN_TYPE_CHANGE
    * ASSET_TRANSFER
    * OPTIONS_PREMIUM_FEE
    * OPTIONS_SETTLE_PROFIT
    * AUTO_EXCHANGE

* The field "bc" represents the balance change except for PnL and commission.

## Event: Order Update


> **Payload:**

```javascript
{
  
  "e":"ORDER_TRADE_UPDATE",		// Event Type
  "E":1568879465651,			// Event Time
  "T":1568879465650,			// Transaction Time
  "o":{								
    "s":"BTCUSDT",				// Symbol
    "c":"TEST",					// Client Order Id
      // special client order id:
      // starts with "autoclose-": liquidation order
      // "adl_autoclose": ADL auto close order
    "S":"SELL",					// Side
    "o":"TRAILING_STOP_MARKET",	// Order Type
    "f":"GTC",					// Time in Force
    "q":"0.001",				// Original Quantity
    "p":"0",					// Original Price
    "ap":"0",					// Average Price
    "sp":"7103.04",				// Stop Price. Please ignore with TRAILING_STOP_MARKET order
    "x":"NEW",					// Execution Type
    "X":"NEW",					// Order Status
    "i":8886774,				// Order Id
    "l":"0",					// Order Last Filled Quantity
    "z":"0",					// Order Filled Accumulated Quantity
    "L":"0",					// Last Filled Price
    "N":"USDT",            	// Commission Asset, will not push if no commission
    "n":"0",               	// Commission, will not push if no commission
    "T":1568879465651,			// Order Trade Time
    "t":0,			        	// Trade Id
    "b":"0",			    	// Bids Notional
    "a":"9.91",					// Ask Notional
    "m":false,					// Is this trade the maker side?
    "R":false,					// Is this reduce only
    "wt":"CONTRACT_PRICE", 		// Stop Price Working Type
    "ot":"TRAILING_STOP_MARKET",	// Original Order Type
    "ps":"LONG",						// Position Side
    "cp":false,						// If Close-All, pushed with conditional order
    "AP":"7476.89",				// Activation Price, only puhed with TRAILING_STOP_MARKET order
    "cr":"5.0",					// Callback Rate, only puhed with TRAILING_STOP_MARKET order
    "rp":"0"							// Realized Profit of the trade
  }
  
}
```


When new order created, order status changed will push such event.
event type is `ORDER_TRADE_UPDATE`.





**Side**

* BUY
* SELL

**Order Type**

* MARKET
* LIMIT
* STOP
* TAKE_PROFIT
* LIQUIDATION

**Execution Type**

* NEW
* CANCELED
* CALCULATED		 - Liquidation Execution
* EXPIRED
* TRADE

**Order Status**

* NEW
* PARTIALLY_FILLED
* FILLED
* CANCELED
* EXPIRED
* NEW_INSURANCE     - Liquidation with Insurance Fund
* NEW_ADL				- Counterparty Liquidation`

**Time in force**

* GTC
* IOC
* FOK
* GTX
* HIDDEN

**Working Type**

* MARK_PRICE
* CONTRACT_PRICE



## Event: Account Configuration Update previous Leverage Update

> **Payload:**

```javascript
{
    "e":"ACCOUNT_CONFIG_UPDATE",       // Event Type
    "E":1611646737479,		           // Event Time
    "T":1611646737476,		           // Transaction Time
    "ac":{								
    "s":"BTCUSDT",					   // symbol
    "l":25						       // leverage
     
    }
}  
 
```

> **Or**

```javascript
{
    "e":"ACCOUNT_CONFIG_UPDATE",       // Event Type
    "E":1611646737479,		           // Event Time
    "T":1611646737476,		           // Transaction Time
    "ai":{							   // User's Account Configuration
    "j":true,						   // Multi-Assets Mode
    "f":true,                          // Specified token fee deduction
    "d":true                           // Position mode: true for dual-side (hedge) mode, false for single-side (one-way) mode
    }
}  
```

When the account configuration is changed, the event type will be pushed as `ACCOUNT_CONFIG_UPDATE`

When the leverage of a trade pair changes, the payload will contain the object `ac` to represent the account configuration of the trade pair, where `s` represents the specific trade pair and `l` represents the leverage

When the user Multi-Assets margin mode changes the payload will contain the object `ai` representing the user account configuration, where `j` represents the user Multi-Assets margin mode



# Error Codes

> Here is the error JSON payload:

```javascript
{
  "code":-1121,
  "msg":"Invalid symbol."
}
```

Errors consist of two parts: an error code and a message.    
Codes are universal,but messages can vary.



## 10xx - General Server or Network issues
> -1000 UNKNOWN
* An unknown error occured while processing the request.

> -1001 DISCONNECTED
* Internal error; unable to process your request. Please try again.

> -1002 UNAUTHORIZED
* You are not authorized to execute this request.

> -1003 TOO_MANY_REQUESTS
* Too many requests queued.
* Too many requests; please use the websocket for live updates.
* Too many requests; current limit is %s requests per minute. Please use the websocket for live updates to avoid polling the API.
* Way too many requests; IP banned until %s. Please use the websocket for live updates to avoid bans.

> -1004 DUPLICATE_IP
* This IP is already on the white list

> -1005 NO_SUCH_IP
* No such IP has been white listed

> -1006 UNEXPECTED_RESP
* An unexpected response was received from the message bus. Execution status unknown.

> -1007 TIMEOUT
* Timeout waiting for response from backend server. Send status unknown; execution status unknown.

> -1010 ERROR_MSG_RECEIVED
* ERROR_MSG_RECEIVED.

> -1011 NON_WHITE_LIST
* This IP cannot access this route.

> -1013 INVALID_MESSAGE
* INVALID_MESSAGE.

> -1014 UNKNOWN_ORDER_COMPOSITION
* Unsupported order combination.

> -1015 TOO_MANY_ORDERS
* Too many new orders.
* Too many new orders; current limit is %s orders per %s.

> -1016 SERVICE_SHUTTING_DOWN
* This service is no longer available.

> -1020 UNSUPPORTED_OPERATION
* This operation is not supported.

> -1021 INVALID_TIMESTAMP
* Timestamp for this request is outside of the recvWindow.
* Timestamp for this request was 1000ms ahead of the server's time.

> -1022 INVALID_SIGNATURE
* Signature for this request is not valid.

> -1023 START_TIME_GREATER_THAN_END_TIME
* Start time is greater than end time.


## 11xx - Request issues
> -1100 ILLEGAL_CHARS
* Illegal characters found in a parameter.
* Illegal characters found in parameter '%s'; legal range is '%s'.

> -1101 TOO_MANY_PARAMETERS
* Too many parameters sent for this endpoint.
* Too many parameters; expected '%s' and received '%s'.
* Duplicate values for a parameter detected.

> -1102 MANDATORY_PARAM_EMPTY_OR_MALFORMED
* A mandatory parameter was not sent, was empty/null, or malformed.
* Mandatory parameter '%s' was not sent, was empty/null, or malformed.
* Param '%s' or '%s' must be sent, but both were empty/null!

> -1103 UNKNOWN_PARAM
* An unknown parameter was sent.

> -1104 UNREAD_PARAMETERS
* Not all sent parameters were read.
* Not all sent parameters were read; read '%s' parameter(s) but was sent '%s'.

> -1105 PARAM_EMPTY
* A parameter was empty.
* Parameter '%s' was empty.

> -1106 PARAM_NOT_REQUIRED
* A parameter was sent when not required.
* Parameter '%s' sent when not required.

> -1108 BAD_ASSET
* Invalid asset.

> -1109 BAD_ACCOUNT
* Invalid account.

> -1110 BAD_INSTRUMENT_TYPE
* Invalid symbolType.

> -1111 BAD_PRECISION
* Precision is over the maximum defined for this asset.

> -1112 NO_DEPTH
* No orders on book for symbol.

> -1113 WITHDRAW_NOT_NEGATIVE
* Withdrawal amount must be negative.

> -1114 TIF_NOT_REQUIRED
* TimeInForce parameter sent when not required.

> -1115 INVALID_TIF
* Invalid timeInForce.

> -1116 INVALID_ORDER_TYPE
* Invalid orderType.

> -1117 INVALID_SIDE
* Invalid side.

> -1118 EMPTY_NEW_CL_ORD_ID
* New client order ID was empty.

> -1119 EMPTY_ORG_CL_ORD_ID
* Original client order ID was empty.

> -1120 BAD_INTERVAL
* Invalid interval.

> -1121 BAD_SYMBOL
* Invalid symbol.

> -1125 INVALID_LISTEN_KEY
* This listenKey does not exist.

> -1127 MORE_THAN_XX_HOURS
* Lookup interval is too big.
* More than %s hours between startTime and endTime.

> -1128 OPTIONAL_PARAMS_BAD_COMBO
* Combination of optional parameters invalid.

> -1130 INVALID_PARAMETER
* Invalid data sent for a parameter.
* Data sent for parameter '%s' is not valid.

> -1136 INVALID_NEW_ORDER_RESP_TYPE
* Invalid newOrderRespType.


## 20xx - Processing Issues

> -2010 NEW_ORDER_REJECTED
* NEW_ORDER_REJECTED

> -2011 CANCEL_REJECTED
* CANCEL_REJECTED

> -2013 NO_SUCH_ORDER
* Order does not exist.

> -2014 BAD_API_KEY_FMT
* API-key format invalid.

> -2015 REJECTED_MBX_KEY
* Invalid API-key, IP, or permissions for action.

> -2016 NO_TRADING_WINDOW
* No trading window could be found for the symbol. Try ticker/24hrs instead.

> -2018 BALANCE_NOT_SUFFICIENT
* Balance is insufficient.

> -2019 MARGIN_NOT_SUFFICIEN
* Margin is insufficient.

> -2020 UNABLE_TO_FILL
* Unable to fill.

> -2021 ORDER_WOULD_IMMEDIATELY_TRIGGER
* Order would immediately trigger.

> -2022 REDUCE_ONLY_REJECT
* ReduceOnly Order is rejected.

> -2023 USER_IN_LIQUIDATION
* User in liquidation mode now.

> -2024 POSITION_NOT_SUFFICIENT
* Position is not sufficient.

> -2025 MAX_OPEN_ORDER_EXCEEDED
* Reach max open order limit.

> -2026 REDUCE_ONLY_ORDER_TYPE_NOT_SUPPORTED
* This OrderType is not supported when reduceOnly.

> -2027 MAX_LEVERAGE_RATIO
* Exceeded the maximum allowable position at current leverage.


> -2028 MIN_LEVERAGE_RATIO
* Leverage is smaller than permitted: insufficient margin balance.


## 40xx - Filters and other Issues
> -4000 INVALID_ORDER_STATUS
* Invalid order status.

> -4001 PRICE_LESS_THAN_ZERO
* Price less than 0.

> -4002 PRICE_GREATER_THAN_MAX_PRICE
* Price greater than max price.

> -4003 QTY_LESS_THAN_ZERO
* Quantity less than zero.

> -4004 QTY_LESS_THAN_MIN_QTY
* Quantity less than min quantity.

> -4005 QTY_GREATER_THAN_MAX_QTY
* Quantity greater than max quantity.

> -4006 STOP_PRICE_LESS_THAN_ZERO
* Stop price less than zero.

> -4007 STOP_PRICE_GREATER_THAN_MAX_PRICE
* Stop price greater than max price.

> -4008 TICK_SIZE_LESS_THAN_ZERO
* Tick size less than zero.

> -4009 MAX_PRICE_LESS_THAN_MIN_PRICE
* Max price less than min price.

> -4010 MAX_QTY_LESS_THAN_MIN_QTY
* Max qty less than min qty.

> -4011 STEP_SIZE_LESS_THAN_ZERO
* Step size less than zero.

> -4012 MAX_NUM_ORDERS_LESS_THAN_ZERO
* Max mum orders less than zero.

> -4013 PRICE_LESS_THAN_MIN_PRICE
* Price less than min price.

> -4014 PRICE_NOT_INCREASED_BY_TICK_SIZE
* Price not increased by tick size.

> -4015 INVALID_CL_ORD_ID_LEN
* Client order id is not valid.
* Client order id length should not be more than 36 chars

> -4016 PRICE_HIGHTER_THAN_MULTIPLIER_UP
* Price is higher than mark price multiplier cap.

> -4017 MULTIPLIER_UP_LESS_THAN_ZERO
* Multiplier up less than zero.

> -4018 MULTIPLIER_DOWN_LESS_THAN_ZERO
* Multiplier down less than zero.

> -4019 COMPOSITE_SCALE_OVERFLOW
* Composite scale too large.

> -4020 TARGET_STRATEGY_INVALID
* Target strategy invalid for orderType '%s',reduceOnly '%b'.

> -4021 INVALID_DEPTH_LIMIT
* Invalid depth limit.
* '%s' is not valid depth limit.

> -4022 WRONG_MARKET_STATUS
* market status sent is not valid.

> -4023 QTY_NOT_INCREASED_BY_STEP_SIZE
* Qty not increased by step size.

> -4024 PRICE_LOWER_THAN_MULTIPLIER_DOWN
* Price is lower than mark price multiplier floor.

> -4025 MULTIPLIER_DECIMAL_LESS_THAN_ZERO
* Multiplier decimal less than zero.

> -4026 COMMISSION_INVALID
* Commission invalid.
* `%s` less than zero.
* `%s` absolute value greater than `%s`

> -4027 INVALID_ACCOUNT_TYPE
* Invalid account type.

> -4028 INVALID_LEVERAGE
* Invalid leverage
* Leverage `%s` is not valid
* Leverage `%s` already exist with `%s`

> -4029 INVALID_TICK_SIZE_PRECISION
* Tick size precision is invalid.

> -4030 INVALID_STEP_SIZE_PRECISION
* Step size precision is invalid.

> -4031 INVALID_WORKING_TYPE
* Invalid parameter working type
* Invalid parameter working type: `%s`

> -4032 EXCEED_MAX_CANCEL_ORDER_SIZE
* Exceed maximum cancel order size.
* Invalid parameter working type: `%s`

> -4033 INSURANCE_ACCOUNT_NOT_FOUND
* Insurance account not found.

> -4044 INVALID_BALANCE_TYPE
* Balance Type is invalid.

> -4045 MAX_STOP_ORDER_EXCEEDED
* Reach max stop order limit.

> -4046 NO_NEED_TO_CHANGE_MARGIN_TYPE
* No need to change margin type.

> -4047 THERE_EXISTS_OPEN_ORDERS
* Margin type cannot be changed if there exists open orders.

> -4048 THERE_EXISTS_QUANTITY
* Margin type cannot be changed if there exists position.

> -4049 ADD_ISOLATED_MARGIN_REJECT
* Add margin only support for isolated position.

> -4050 CROSS_BALANCE_INSUFFICIENT
* Cross balance insufficient.

> -4051 ISOLATED_BALANCE_INSUFFICIENT
* Isolated balance insufficient.

> -4052 NO_NEED_TO_CHANGE_AUTO_ADD_MARGIN
* No need to change auto add margin.

> -4053 AUTO_ADD_CROSSED_MARGIN_REJECT
* Auto add margin only support for isolated position.

> -4054 ADD_ISOLATED_MARGIN_NO_POSITION_REJECT
* Cannot add position margin: position is 0.

> -4055 AMOUNT_MUST_BE_POSITIVE
* Amount must be positive.

> -4056 INVALID_API_KEY_TYPE
* Invalid api key type.

> -4057 INVALID_RSA_PUBLIC_KEY
* Invalid api public key

> -4058 MAX_PRICE_TOO_LARGE
* maxPrice and priceDecimal too large,please check.

> -4059 NO_NEED_TO_CHANGE_POSITION_SIDE
* No need to change position side.

> -4060 INVALID_POSITION_SIDE
* Invalid position side.

> -4061 POSITION_SIDE_NOT_MATCH
* Order's position side does not match user's setting.

> -4062 REDUCE_ONLY_CONFLICT
* Invalid or improper reduceOnly value.

> -4063 INVALID_OPTIONS_REQUEST_TYPE
* Invalid options request type

> -4064 INVALID_OPTIONS_TIME_FRAME
* Invalid options time frame

> -4065 INVALID_OPTIONS_AMOUNT
* Invalid options amount

> -4066 INVALID_OPTIONS_EVENT_TYPE
* Invalid options event type

> -4067 POSITION_SIDE_CHANGE_EXISTS_OPEN_ORDERS
* Position side cannot be changed if there exists open orders.

> -4068 POSITION_SIDE_CHANGE_EXISTS_QUANTITY
* Position side cannot be changed if there exists position.

> -4069 INVALID_OPTIONS_PREMIUM_FEE
* Invalid options premium fee

> -4070 INVALID_CL_OPTIONS_ID_LEN
* Client options id is not valid.
* Client options id length should be less than 32 chars

> -4071 INVALID_OPTIONS_DIRECTION
* Invalid options direction

> -4072 OPTIONS_PREMIUM_NOT_UPDATE
* premium fee is not updated, reject order

> -4073 OPTIONS_PREMIUM_INPUT_LESS_THAN_ZERO
* input premium fee is less than 0, reject order

> -4074 OPTIONS_AMOUNT_BIGGER_THAN_UPPER
* Order amount is bigger than upper boundary or less than 0, reject order

> -4075 OPTIONS_PREMIUM_OUTPUT_ZERO
* output premium fee is less than 0, reject order

> -4076 OPTIONS_PREMIUM_TOO_DIFF
* original fee is too much higher than last fee

> -4077 OPTIONS_PREMIUM_REACH_LIMIT
* place order amount has reached to limit, reject order

> -4078 OPTIONS_COMMON_ERROR
* options internal error

> -4079 INVALID_OPTIONS_ID
* invalid options id
* invalid options id: %s
* duplicate options id %d for user %d

> -4080 OPTIONS_USER_NOT_FOUND
* user not found
* user not found with id: %s

> -4081 OPTIONS_NOT_FOUND
* options not found
* options not found with id: %s

> -4082 INVALID_BATCH_PLACE_ORDER_SIZE
* Invalid number of batch place orders.
* Invalid number of batch place orders: %s

> -4083 PLACE_BATCH_ORDERS_FAIL
* Fail to place batch orders.

> -4084 UPCOMING_METHOD
* Method is not allowed currently. Upcoming soon.

> -4085 INVALID_NOTIONAL_LIMIT_COEF
* Invalid notional limit coefficient

> -4086 INVALID_PRICE_SPREAD_THRESHOLD
* Invalid price spread threshold

> -4087 REDUCE_ONLY_ORDER_PERMISSION
* User can only place reduce only order

> -4088 NO_PLACE_ORDER_PERMISSION
* User can not place order currently

> -4104 INVALID_CONTRACT_TYPE
* Invalid contract type

> -4114 INVALID_CLIENT_TRAN_ID_LEN
* clientTranId  is not valid
* Client tran id length should be less than 64 chars

> -4115 DUPLICATED_CLIENT_TRAN_ID
* clientTranId  is duplicated
* Client tran id should be unique within 7 days

> -4118 REDUCE_ONLY_MARGIN_CHECK_FAILED
* ReduceOnly Order Failed. Please check your existing position and open orders

> -4131 MARKET_ORDER_REJECT
* The counterparty's best price does not meet the PERCENT_PRICE filter limit

> -4135 INVALID_ACTIVATION_PRICE
* Invalid activation price

> -4137 QUANTITY_EXISTS_WITH_CLOSE_POSITION
* Quantity must be zero with closePosition equals true

> -4138 REDUCE_ONLY_MUST_BE_TRUE
* Reduce only must be true with closePosition equals true

> -4139 ORDER_TYPE_CANNOT_BE_MKT
* Order type can not be market if it's unable to cancel

> -4140 INVALID_OPENING_POSITION_STATUS
* Invalid symbol status for opening position

> -4141 SYMBOL_ALREADY_CLOSED
* Symbol is closed

> -4142 STRATEGY_INVALID_TRIGGER_PRICE
* REJECT: take profit or stop order will be triggered immediately

> -4144 INVALID_PAIR
* Invalid pair

> -4161 ISOLATED_LEVERAGE_REJECT_WITH_POSITION
* Leverage reduction is not supported in Isolated Margin Mode with open positions

> -4164 MIN_NOTIONAL
* Order's notional must be no smaller than 5.0 (unless you choose reduce only)
* Order's notional must be no smaller than %s (unless you choose reduce only)

> -4165 INVALID_TIME_INTERVAL
* Invalid time interval
* Maximum time interval is %s days

> -4183 PRICE_HIGHTER_THAN_STOP_MULTIPLIER_UP
* Price is higher than stop price multiplier cap.
* Limit price can't be higher than %s.

> -4184 PRICE_LOWER_THAN_STOP_MULTIPLIER_DOWN
* Price is lower than stop price multiplier floor.
* Limit price can't be lower than %s.