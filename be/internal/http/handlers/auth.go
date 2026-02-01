package handlers

import (
	"net/http"

	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/http/response"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
)

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

func LoginHandler(authService service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, ok := middleware.GetBody[LoginRequest](r)
		if !ok {
			response.SendError(w, http.StatusInternalServerError, "invalid request context")
			return
		}

		// Get user agent and IP address from request
		userAgent := r.UserAgent()
		ipAddress := r.RemoteAddr

		accessToken, refreshToken, expiresIn, err := authService.Login(
			r.Context(),
			req.Email,
			req.Password,
			userAgent,
			ipAddress,
		)
		if err != nil {
			var statusCode int
			switch err {
			case service.ErrUserNotFound, service.ErrInvalidPassword:
				statusCode = http.StatusUnauthorized
			default:
				statusCode = http.StatusInternalServerError
			}
			response.SendError(w, statusCode, err.Error())
			return
		}

		resp := LoginResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    expiresIn,
		}

		response.SendSuccess(w, http.StatusOK, resp)
	}
}

func RefreshTokenHandler(authService service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, ok := middleware.GetBody[RefreshTokenRequest](r)
		if !ok {
			response.SendError(w, http.StatusInternalServerError, "invalid request context")
			return
		}

		userAgent := r.UserAgent()
		ipAddress := r.RemoteAddr

		accessToken, expiresIn, err := authService.RefreshToken(
			r.Context(),
			req.RefreshToken,
			userAgent,
			ipAddress,
		)
		if err != nil {
			var statusCode int
			switch err {
			case service.ErrSessionNotFound, service.ErrInvalidRefreshToken:
				statusCode = http.StatusUnauthorized
			default:
				statusCode = http.StatusInternalServerError
			}
			response.SendError(w, statusCode, err.Error())
			return
		}

		resp := RefreshTokenResponse{
			AccessToken: accessToken,
			ExpiresIn:   expiresIn,
		}

		response.SendSuccess(w, http.StatusOK, resp)
	}
}

func LogoutHandler(authService service.AuthService, jwtService service.JWTService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, ok := middleware.GetBody[LogoutRequest](r)
		if !ok {
			response.SendError(w, http.StatusInternalServerError, "invalid request context")
			return
		}

		// Get user ID from context
		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		// Get JTI from context (set by auth middleware)
		jti, ok := middleware.GetJTI(r)
		if !ok {
			jti = "" // If not available, skip blacklisting
		}

		err := authService.Logout(r.Context(), req.RefreshToken, jti, userID)
		if err != nil {
			response.SendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		response.SendSuccess(w, http.StatusOK, map[string]string{
			"message": "logged out successfully",
		})
	}
}

func LogoutAllHandler(authService service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from context
		userID, ok := middleware.GetUserID(r)
		if !ok {
			response.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		err := authService.LogoutAllSessions(r.Context(), userID)
		if err != nil {
			response.SendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		response.SendSuccess(w, http.StatusOK, map[string]string{
			"message": "all sessions logged out successfully",
		})
	}
}
