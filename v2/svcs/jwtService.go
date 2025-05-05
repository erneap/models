package svcs

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/erneap/models/v2/users"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateToken(userid primitive.ObjectID, email string) (string, error) {
	key := []byte(strings.TrimSpace(os.Getenv("JWT_SECRET")))
	expireTime := time.Now().Add(6 * time.Hour)
	claims := &users.JWTClaim{
		UserID:       userid.Hex(),
		EmailAddress: email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ValidateToken(signedToken string) (*users.JWTClaim, error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&users.JWTClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(strings.TrimSpace(os.Getenv("JWT_SECRET"))), nil
		},
	)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*users.JWTClaim)
	if !ok {
		return nil, errors.New("couldn't parse claims")
	}
	if claims.ExpiresAt < time.Now().Local().Unix() {
		return nil, errors.New("token expired")
	}
	return claims, nil
}

func GetRequestor(context *gin.Context) string {
	tokenString := context.GetHeader("Authorization")
	if tokenString == "" {
		return ""
	}
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return ""
	}
	return claims.UserID
}

func CheckJWT(app string) gin.HandlerFunc {
	return func(context *gin.Context) {
		tokenString := context.GetHeader("Authorization")
		userID := GetRequestor(context)
		user, _ := GetUserByID(userID)
		if tokenString == "" {
			CreateDBLogEntry("authentication", app, "CheckJWT Error", "",
				"No Authentication Token Passed", context)
			context.JSON(http.StatusUnauthorized, gin.H{"error": "request does not contain an access token"})
			context.Abort()
			return
		}
		claims, err := ValidateToken(tokenString)
		if err != nil {
			CreateDBLogEntry("authentication", app, "CheckJWT Error", user.LastName,
				"Validation Error", context)
			context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			context.Abort()
			return
		}

		// replace token by passing a new token in the response header
		CreateDBLogEntry("authentication", app, "CheckJWT", user.LastName,
			"Token Verified", context)
		id, _ := primitive.ObjectIDFromHex(claims.UserID)
		tokenString, _ = CreateToken(id, claims.EmailAddress)
		context.Writer.Header().Set("Token", tokenString)
		context.Next()
	}
}

func CheckRole(prog, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			CreateDBLogEntry("authentication", prog, "CheckRole Error", "",
				"No Authentication Token Passed", c)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "request does not contain an access token"})
			c.Abort()
			return
		}
		_, err := ValidateToken(tokenString)
		userID := GetRequestor(c)
		user, err2 := GetUserByID(userID)
		if err != nil {
			CreateDBLogEntry("authentication", prog, "CheckRole Error", user.LastName,
				"Validation Error", c)
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		if err2 != nil {
			CreateDBLogEntry("authentication", prog, "CheckRole Error", userID,
				"User Not Found", c)
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found: " + err.Error()})
			c.Abort()
			return
		}
		if !user.IsInGroup(prog, role) {
			CreateDBLogEntry("authentication", prog, "CheckRole Error", user.LastName,
				"User Not In Group", c)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not in group"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func CheckRoles(prog string, roles []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			CreateDBLogEntry("authentication", prog, "CheckRoles Error", "",
				"No Authentication Token passed", c)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "request does not contain an access token"})
			c.Abort()
			return
		}
		claims, err := ValidateToken(tokenString)
		if err != nil {
			CreateDBLogEntry("authentication", prog, "CheckRoles Error", "",
				"Validation error", c)
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		user, err := GetUserByID(claims.UserID)
		if err != nil {
			CreateDBLogEntry("authentication", prog, "CheckRoles Error", claims.UserID,
				"User Not Found", c)
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found: " + err.Error()})
			c.Abort()
			return
		}
		inRole := false
		for i := 0; i < len(roles) && !inRole; i++ {
			if user.IsInGroup(prog, roles[i]) {
				inRole = true
			}
		}
		if !inRole {
			CreateDBLogEntry("authentication", prog, "CheckRoles Error", user.LastName,
				"User not in Groups", c)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not in group"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func CheckRoleList(app string, roles []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			CreateDBLogEntry("authentication", app, "CheckRoleList Error", "",
				"No Authentication Token passed", c)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "request does not contain an access token"})
			c.Abort()
			return
		}
		claims, err := ValidateToken(tokenString)
		if err != nil {
			CreateDBLogEntry("authentication", app, "CheckRoleList Error", "",
				"Validation Error: "+err.Error(), c)
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		user, err := GetUserByID(claims.UserID)
		if err != nil {
			CreateDBLogEntry("authentication", app, "CheckRoleList Error", claims.UserID,
				"User Not Found: "+err.Error(), c)
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found: " + err.Error()})
			c.Abort()
			return
		}
		inRole := false
		for i := 0; i < len(roles) && !inRole; i++ {
			parts := strings.Split(roles[i], "-")
			if user.IsInGroup(parts[0], parts[1]) {
				inRole = true
			}
		}
		if !inRole {
			CreateDBLogEntry("authentication", app, "CheckRoleList Error", user.LastName,
				"User Not in list of roles provided", c)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not in group"})
			c.Abort()
			return
		}
		c.Next()
	}
}
