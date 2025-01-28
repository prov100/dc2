package common

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

func Router(router *http.ServeMux) http.Handler {
	fmt.Println("common./middleware.go Router() started")
	fmt.Println("common./middleware.go Router() started router is", router)
	return HandleCacheControl(router)
}

func HandleCacheControl(next *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		fmt.Println("common./middleware.go HandleCacheControl() started rw is", rw)
		fmt.Println("common./middleware.go HandleCacheControl() started req is", req)
		headers := rw.Header()
		headers.Set("Cache-Control", "no-cache, no-store, max-age=0, must-revalidate")
		headers.Set("Pragma", "no-cache")
		headers.Set("Expires", "0")
		next.ServeHTTP(rw, req)
	})
}

// AddMiddleware - adds middleware to a Handler
func AddMiddleware(h http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
	for _, mw := range middleware {
		h = mw(h)
	}
	return h
}

// CustomClaims contains custom data we want from the token.
type CustomClaims struct {
	Permissions []string `json:"permissions"`
	Email       string   `json:"email"`
}

// Validate does nothing for this example, but we need
// it to satisfy validator.CustomClaims interface.
func (c CustomClaims) Validate(ctx context.Context) error {
	return nil
}

func (c CustomClaims) HasPermissions(expectedClaims []string) bool {
	fmt.Println("internal/common/middleware.go HasPermissions() started")
	fmt.Println("internal/common/middleware.go HasPermissions() expectedClaims", expectedClaims)
	if len(expectedClaims) == 0 {
		return false
	}
	for _, scope := range expectedClaims {
		if !Contains(c.Permissions, scope) {
			return false
		}
	}
	fmt.Println("internal/common/middleware.go HasPermissions() ended")
	return true
}

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func ValidatePermissions(expectedClaims []string, audience string, domain string) func(next http.Handler) http.Handler {
	fmt.Println("internal/common/middleware.go ValidatePermissions1111")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			/*fmt.Println("ValidatePermissions")
			fmt.Println("ValidatePermissions expectedClaims", expectedClaims)
			fmt.Println("ValidatePermissions r.Context()", r.Context())
			fmt.Println("ValidatePermissions r.Context().Value(jwtmiddleware.ContextKey{})", r.Context().Value(jwtmiddleware.ContextKey{}))
			token := r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
			fmt.Println("ValidatePermissions token", token)
			fmt.Println("ValidatePermissions token.RegisteredClaims", token.RegisteredClaims)
			claims := token.CustomClaims.(*CustomClaims)
			fmt.Println("ValidatePermissions claims", claims)*/
			/*authHeaderParts := strings.Fields(r.Header.Get("Authorization"))
			fmt.Println("authHeaderParts[1]", authHeaderParts[1])

			issuerURL, err := url.Parse("https://" + domain + "/")
			if err != nil {
				fmt.Println("Failed to parse the issuer url: %v", err)
				http.Error(w, "Failed to parse the issuer url", http.StatusUnauthorized)
				log.Error("Error",
					zap.Int("msgnum", 752),
					zap.Error(err))
				return
			}
			fmt.Println("internal/common/middleware.go issuerURL is", issuerURL)

			provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)
			jwtValidator, err := validator.New(
				provider.KeyFunc,
				validator.RS256,
				issuerURL.String(),
				[]string{audience},
				validator.WithCustomClaims(func() validator.CustomClaims {
					return new(CustomClaims)
				}),
			)
			if err != nil {
				fmt.Println("err", err)
				return
			}
			tokenClaims, err := jwtValidator.ValidateToken(r.Context(), authHeaderParts[1])
			if err != nil {
				fmt.Println("err", err)
				return
			}
			m := tokenClaims.(*validator.ValidatedClaims)
			fmt.Println("tokenClaims.CustomClaims", m.CustomClaims)
			claims := m.CustomClaims.(*CustomClaims)*/

			tokenString, err := getToken(r)
			if err != nil {
				http.Error(w, "Error parsing token", http.StatusUnauthorized)
				return
			}
			_, claims, err := getClaims(audience, domain, tokenString, w, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			fmt.Println("claims", claims)
			fmt.Println("claims.Permissions", claims.Permissions)
			if len(claims.Permissions) == 0 {
				fmt.Println("in len(claims.Permissions) err is permission denied")
				RenderJSON(w, "Permission Denied")
				return
			}
			fmt.Println("ValidatePermissions111111")
			if !claims.HasPermissions(expectedClaims) {
				RenderJSON(w, "Permission Denied")
				return
			}
			fmt.Println("ValidatePermissions end")
			next.ServeHTTP(w, r)
		})
	}
}

func EnsureValidToken(audience string, domain string) func(http.Handler) http.Handler {
	fmt.Println("internal/common/middleware.go EnsureValidToken1111")
	return func(next http.Handler) http.Handler {
		fmt.Println("internal/common/middleware.go EnsureValidToken2222")
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("r is", r)
			fmt.Println("internal/common/middleware.go EnsureValidToken3333")
			fmt.Println("internal/common/middleware.go audience is", audience)
			fmt.Println("internal/common/middleware.go domain is", domain)
			// domain := "dev-llzybv0gk4ybnuqm.us.auth0.com"

			/*issuerURL, err := url.Parse("https://" + domain + "/")
			if err != nil {
				http.Error(w, "Failed to parse the issuer url", http.StatusUnauthorized)
				RenderJSON(w, "Permission Denied")
				return
			}
			fmt.Println("internal/common/middleware.go issuerURL is", issuerURL)

			provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)
			// audience := "https://hello-world.example.com"
			jwtValidator, err := validator.New(
				provider.KeyFunc,
				validator.RS256,
				issuerURL.String(),
				[]string{audience},
				validator.WithCustomClaims(func() validator.CustomClaims {
					return new(CustomClaims)
				}),
			)
			if err != nil {
				fmt.Println("Failed to set up the jwt validator")
				http.Error(w, "Failed to set up the jwt validator", http.StatusUnauthorized)
				return
			}
			fmt.Println("internal/common/middleware.go EnsureValidToken4444444444")
			authHeaderParts := strings.Fields(r.Header.Get("Authorization"))
			fmt.Println("internal/common/middleware.go authHeaderParts", authHeaderParts)
			fmt.Println("internal/common/middleware.go authHeaderParts[1]", authHeaderParts[1])
			tokenString = authHeaderParts[1]
			if len(authHeaderParts) > 0 && strings.ToLower(authHeaderParts[0]) != "bearer" {
				http.Error(w, "Error parsing token", http.StatusUnauthorized)
				return
			}

			errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
				fmt.Println("Encountered error while validating JWT", err)
				if errors.Is(err, jwtmiddleware.ErrJWTMissing) {
					http.Error(w, "Error parsing token", http.StatusUnauthorized)
					return
				}
				if errors.Is(err, jwtmiddleware.ErrJWTInvalid) {
					http.Error(w, "Error parsing token", http.StatusUnauthorized)
					return
				}
			}

			middleware := jwtmiddleware.New(
				jwtValidator.ValidateToken,
				jwtmiddleware.WithErrorHandler(errorHandler),
			)
			fmt.Println("internal/common/middleware.go EnsureValidToken middleware", middleware)
			fmt.Println("internal/common/middleware.go tokenString", tokenString)

			tokenClaims, err := jwtValidator.ValidateToken(r.Context(), tokenString)
			if err != nil {
				http.Error(w, "Error parsing token", http.StatusUnauthorized)
				return
			}
			m := tokenClaims.(*validator.ValidatedClaims)
			fmt.Println("tokenClaims.CustomClaims", m.CustomClaims)
			claims := m.CustomClaims.(*CustomClaims)
			fmt.Println("email is", claims.Email)*/

			// fmt.Println("tokenClaims.RegisteredClaims", m.RegisteredClaims)
			tokenString, err := getToken(r)
			if err != nil {
				http.Error(w, "Error parsing token", http.StatusUnauthorized)
				return
			}
			middleware, claims, err := getClaims(audience, domain, tokenString, w, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			v := ContextStruct{}

			// for now i hardcoded email but we need this frm custom claims
			// v.Email = "sprov300@gmail.com"
			v.Email = claims.Email
			v.TokenString = tokenString

			fmt.Println("v.Email", v.Email)

			ctx := context.WithValue(r.Context(), KeyEmailToken, v)

			// middleware.CheckJWT(next).ServeHTTP(w, r)
			middleware.CheckJWT(next).ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func getToken(r *http.Request) (string, error) {
	authHeaderParts := strings.Fields(r.Header.Get("Authorization"))
	fmt.Println("internal/common/middleware.go authHeaderParts", authHeaderParts)
	fmt.Println("internal/common/middleware.go authHeaderParts[1]", authHeaderParts[1])
	if len(authHeaderParts) > 0 && strings.ToLower(authHeaderParts[0]) != "bearer" {
		return "", errors.New("Error parsing token")
	}
	return authHeaderParts[1], nil
}

func getClaims(audience string, domain string, tokenString string, w http.ResponseWriter, r *http.Request) (*jwtmiddleware.JWTMiddleware, *CustomClaims, error) {
	issuerURL, err := url.Parse("https://" + domain + "/")
	if err != nil {
		return nil, nil, errors.New("Failed to parse the issuer url")
	}
	fmt.Println("internal/common/middleware.go issuerURL is", issuerURL)

	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)
	// audience := "https://hello-world.example.com"
	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{audience},
		validator.WithCustomClaims(func() validator.CustomClaims {
			return new(CustomClaims)
		}),
	)
	if err != nil {
		fmt.Println("Failed to set up the jwt validator")
		return nil, nil, errors.New("Failed to set up the jwt validator")
	}
	fmt.Println("internal/common/middleware.go EnsureValidToken4444444444")

	errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		fmt.Println("Encountered error while validating JWT", err)
		if errors.Is(err, jwtmiddleware.ErrJWTMissing) {
			http.Error(w, "Error parsing token", http.StatusUnauthorized)
			return
		}
		if errors.Is(err, jwtmiddleware.ErrJWTInvalid) {
			http.Error(w, "Error parsing token", http.StatusUnauthorized)
			return
		}
	}

	middleware := jwtmiddleware.New(
		jwtValidator.ValidateToken,
		jwtmiddleware.WithErrorHandler(errorHandler),
	)
	fmt.Println("internal/common/middleware.go EnsureValidToken middleware", middleware)
	fmt.Println("internal/common/middleware.go tokenString", tokenString)

	tokenClaims, err := jwtValidator.ValidateToken(r.Context(), tokenString)
	if err != nil {
		return nil, nil, err
	}
	m := tokenClaims.(*validator.ValidatedClaims)
	fmt.Println("tokenClaims.CustomClaims", m.CustomClaims)
	claims := m.CustomClaims.(*CustomClaims)
	fmt.Println("email is", claims.Email)
	return middleware, claims, nil
}

/*// AuthenticateMiddleware - Authenticate Token from request
// AuthenticateMiddleware - Authenticate Token from request
func AuthenticateMiddleware(tokenService *TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString, err := GetAuthBearerToken(r)
			if err != nil {
				log.Error("Error",
					zap.Int("msgnum", 751),
					zap.Error(err))
				http.Error(w, "Error parsing token", http.StatusUnauthorized)
				return
			}
			jwtOpt := GetJWTOpt()

			claims, err := tokenService.ParseToken(tokenString, jwtOpt.AccessSecret)
			if err != nil {
				log.Error("Error",
					zap.Int("msgnum", 751),
					zap.Error(err))
				http.Error(w, "Error parsing token", http.StatusUnauthorized)
				return
			}

			err = tokenService.ValidateToken(claims, false)
			if err != nil {
				log.Error("Error",
					zap.Int("msgnum", 751),
					zap.Error(err))
				http.Error(w, "Error validating token", http.StatusUnauthorized)
				return
			}

			v := ContextStruct{}

			v.Email = claims.EmailAddr
			v.TokenString = tokenString

			ctx := context.WithValue(r.Context(), KeyEmailToken, v)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}*/

// CorsMiddleware - Enable CORS with various options
/*func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers",
				"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Access-Control-Allow-Origin")
		}
		// Stop here if its Preflighted OPTIONS request
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Expose-Headers", "Authorization")
			w.Header().Set("Access-Control-Max-Age", "86400")
			return
		}
		next.ServeHTTP(w, r)
	})
}*/

/*func ValidatePermissions(expectedClaims []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("ValidatePermissions")
		fmt.Println("ValidatePermissions expectedClaims", expectedClaims)
		token := r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
		fmt.Println("ValidatePermissions token", token)
		fmt.Println("ValidatePermissions token.RegisteredClaims", token.RegisteredClaims)
		claims := token.CustomClaims.(*CustomClaims)
		fmt.Println("ValidatePermissions claims", claims)
		if !claims.HasPermissions(expectedClaims) {
			http.Error(w, "Permission Denied", http.StatusUnauthorized)
			log.Error("Error",
				zap.Int("msgnum", 750),
				zap.Error(errors.New("Permission Denied")))
			return
		}
		fmt.Println("ValidatePermissions end")
		next.ServeHTTP(w, r)
	})
}*/

// HasScope checks whether our claims have a specific scope.
/*func (c CustomClaims) HasScope(expectedScope string) bool {
	result := strings.Split(c.Scope, " ")
	for i := range result {
		if result[i] == expectedScope {
			return true
		}
	}

	return false
}*/
