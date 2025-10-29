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

