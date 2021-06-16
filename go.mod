module zoove

go 1.16

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/fasthttp/websocket v1.4.3
	github.com/gofiber/fiber/v2 v2.12.0
	github.com/gofiber/jwt/v2 v2.2.2
	github.com/gofiber/websocket/v2 v2.0.5
	github.com/gomodule/redigo v1.8.4
	github.com/google/uuid v1.2.0
	github.com/iancoleman/strcase v0.1.3
	github.com/joho/godotenv v1.3.0
	github.com/prisma/prisma-client-go v0.9.0
	github.com/shopspring/decimal v1.2.0
	github.com/soveran/redisurl v0.0.0-20180322091936-eb325bc7a4b8
	github.com/takuoki/gocase v1.0.0
	github.com/zmb3/spotify v1.2.0
	golang.org/x/oauth2 v0.0.0-20210514164344-f6687ab2804c
)

replace github.com/prisma/prisma-client-go => ./prisma-client
