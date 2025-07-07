package handlers

import (
	"VoiceSculptor/internal/apidocs"
	"VoiceSculptor/internal/models"
	"net/http"
)

func (h *Handlers) GetDocs() []apidocs.UriDoc {
	// Define the API documentation
	uriDocs := []apidocs.UriDoc{ // test
		{
			Group:   "User Authorization",
			Path:    "/api/auth/login",
			Method:  http.MethodPost,
			Desc:    "User login with email and password",
			Request: apidocs.GetDocDefine(models.LoginForm{}),
			Response: &apidocs.DocField{
				Type: "object",
				Fields: []apidocs.DocField{
					{Name: "email", Type: apidocs.TYPE_STRING},
					{Name: "activation", Type: apidocs.TYPE_BOOLEAN, CanNull: true},
				},
			},
		},
		{
			Group:        "User Authorization",
			Path:         "/api/auth/logout",
			Method:       http.MethodPost,
			AuthRequired: true,
			Desc:         "User logout, if `?next={NEXT_URL}`is not empty, redirect to {NEXT_URL}",
		},
		{
			Group:        "User Authorization",
			Path:         "/api/auth/register",
			Method:       http.MethodPost,
			AuthRequired: false,
			Desc:         "User register with email and password",
			Request:      apidocs.GetDocDefine(models.RegisterUserForm{}),
			Response: &apidocs.DocField{
				Type: "object",
				Fields: []apidocs.DocField{
					{Name: "email", Type: apidocs.TYPE_STRING, Desc: "The email address"},
					{Name: "activation", Type: apidocs.TYPE_BOOLEAN, Desc: "Is the account activated"},
					{Name: "expired", Type: apidocs.TYPE_STRING, Default: "180d", CanNull: true, Desc: "If email verification is required, it will be verified within the valid time"},
				},
			},
		},
		{
			Group:        "User Authorization",
			Path:         "/api/auth/reset_password",
			Method:       http.MethodPost,
			AuthRequired: false,
			Desc:         "Send a verification code to the email address, and then click the link in the email to reset the password",
			Request:      apidocs.GetDocDefine(models.ResetPasswordForm{}),
			Response: &apidocs.DocField{
				Type: "object",
				Fields: []apidocs.DocField{
					{Name: "expired", Type: apidocs.TYPE_STRING, Default: "30m", Desc: "Must be verified within the valid time"},
				},
			},
		},
		{
			Group:        "User Authorization",
			Path:         "/api/auth/reset_password_done",
			Method:       http.MethodPost,
			AuthRequired: false,
			Desc:         "Setup new password",
			Request:      apidocs.GetDocDefine(models.ResetPasswordDoneForm{}),
			Response: &apidocs.DocField{
				Type: apidocs.TYPE_BOOLEAN,
				Desc: "true if success",
			},
		},
		{
			Group:        "User Authorization",
			Path:         "/api/auth/change_password",
			Method:       http.MethodPost,
			AuthRequired: false,
			Desc:         "Setup new password when user is logged in",
			Request:      apidocs.GetDocDefine(models.ChangePasswordForm{}),
			Response: &apidocs.DocField{
				Type: apidocs.TYPE_BOOLEAN,
				Desc: "true if success",
			},
		},
		{
			Group:        "User Authorization",
			Path:         "/api/auth/send/email",
			Method:       http.MethodPost,
			AuthRequired: false,
			Desc:         "Send email verification code",
			Request:      apidocs.GetDocDefine(models.SendEmailVerifyEmail{}),
			Response: &apidocs.DocField{
				Type: "object",
				Fields: []apidocs.DocField{
					{Name: "expired", Type: apidocs.TYPE_STRING, Default: "30m", Desc: "Must be verified within the valid time"},
				},
			},
		},
		{
			Group:        "System Module",
			Path:         "/api/system/health",
			Method:       http.MethodGet,
			Summary:      "数据库健康状态",
			AuthRequired: false,
			Desc:         `检查数据库健康状态`,
		},
	}
	return uriDocs
}
