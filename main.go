package main

import (
	"fmt"
	"net/http"
	"os"
	// "path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

//add CORS (Cross-Origin Resource Sharing) headers to HTTP responses
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		//Access-Control-Allow-Origin header is set to *, which allows any domain to access the resources on the server.
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		// The Access-Control-Allow-Credentials header is set to true, which allows the server to include cookies in the requests and responses.
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		//The Access-Control-Allow-Headers header lists the allowed request headers
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		// Access-Control-Allow-Methods header lists the allowed HTTP methods.
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// middleware function:  checks whether an incoming request has a valid JWT (JSON Web Token) in its Authorization header. 
func ValidateToken() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		fmt.Println(ctx.FullPath())
		//if the incoming request path starts with`/public`, skip the validation and pass request to the next handler
		if strings.HasPrefix(ctx.FullPath(), "/public/") {
			ctx.Next()
			return
		}
		//if Authorization header exists, it is split into two parts (a prefix and the actual token) using the space character as a separator.
		headerToken := ctx.GetHeader("Authorization")

		signedToken := strings.Split(headerToken, " ")
		
		//if the req has invalid Authorization header, abort with status code 401
		if len(signedToken) != 2 {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token string in header",
			})
			return
		}

		//pass token using  jwt.ParseWithClaims() function from the "github.com/dgrijalva/jwt-go" library
		token, err := jwt.ParseWithClaims(
			signedToken[1],
		//validate token  with the JWT registered claims and provide secret key
			&jwt.RegisteredClaims{},
			func(token *jwt.Token) (interface{}, error) {
				return []byte(os.Getenv("AUTH_SECRET")), nil
			},
		)
		//throw status code 401 if theres an error
		if err != nil {
			// utils.Log.Error(err)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}
		//If the token is successfully parsed, the claims are extracted and added to the context using ctx.Set()
		claims, ok := token.Claims.(*jwt.RegisteredClaims)
		ctx.Set("user_id", claims.ID)
		if !ok {
			// utils.Log.Error("couldn't parse claims")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "couldn't parse claims",
			})
			return
		}
		// ExpirsAt time has already been passed
		if claims.ExpiresAt.Before(time.Now()) {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Token expired!",
			})
			return
		}
		// /If the token is valid and has not expired, the function calls ctx.Next() to pass the request to the next handler in the chain.
		ctx.Next()

	}
}

// set up the main router for the Gin web framework. 
func SetupRouter() *gin.Engine {
	//gin initiallization and middleware configuration
	r:= gin.Default()
	// set up Cross-Origin Resource Sharing (CORS)
	r.Use(CORSMiddleware())
	// /validate the JWT token for authenticated routes
	r.Use(ValidateToken())
	r.Use(gin.Recovery())

	// loop over `Routes` map to deetermine HTTP metthod used andd add route to router with corresponding method and handler function
	for path, handlers := range Routes {
		for method, handler := range handlers {
			switch method {
			case "GET":
				r.GET(path, handler)

			case "POST":
				r.POST(path, handler)

			case "PUT":
				r.PUT(path, handler)

			case "PATCH":
				r.PATCH(path, handler)

			case "DELETE":
				r.DELETE(path, handler)
			}
		}
	}

	return r
}

func main(

)