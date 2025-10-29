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
