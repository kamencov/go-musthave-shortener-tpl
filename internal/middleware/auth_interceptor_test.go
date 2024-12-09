package middleware

import (
	"github.com/golang/mock/gomock"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/errorscustom"
	logger2 "github.com/kamencov/go-musthave-shortener-tpl/internal/logger"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/service/auth"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"testing"
)

func TestAuthMiddleware_UnaryInterceptor(t *testing.T) {
	cases := []struct {
		name                         string
		method                       string
		userID                       string
		token                        string
		ctx                          context.Context
		expectedVerifyUserErr        error
		expectedCreatTokenForUserErr error
		expectedCode                 codes.Code
	}{
		{
			name:                  "successful_unary_auth_token_true",
			method:                "/shortener.ShortenerService/ShortenURL",
			userID:                "test",
			ctx:                   context.Background(),
			expectedVerifyUserErr: nil,
			expectedCode:          codes.OK,
		},
		{
			name:                         "successful_unary_auth_token_false",
			method:                       "/shortener.ShortenerService/ShortenURL",
			userID:                       "test",
			token:                        "token",
			ctx:                          context.Background(),
			expectedVerifyUserErr:        errorscustom.ErrBadVarifyToken,
			expectedCreatTokenForUserErr: nil,
			expectedCode:                 codes.OK,
		},
		{
			name:   "successful_metadata_auth_token_true",
			method: "/shortener.ShortenerService/ShortenURL",
			userID: "test",
			token:  "token",
			ctx: metadata.NewIncomingContext(context.Background(),
				metadata.Pairs("authorization", "token")),
			expectedCode: codes.OK,
		},
		{
			name:                         "error_create_token",
			method:                       "/shortener.ShortenerService/ShortenURL",
			userID:                       "test",
			token:                        "token",
			ctx:                          context.Background(),
			expectedVerifyUserErr:        errorscustom.ErrBadVarifyToken,
			expectedCreatTokenForUserErr: errorscustom.ErrBadVarifyToken,
			expectedCode:                 codes.Internal,
		},
		{
			name:   "successful_unary_check_auth_token_true",
			method: "/shortener.ShortenerService/GetUsersURLs",
			userID: "test",
			token:  "token",
			ctx: metadata.NewIncomingContext(context.Background(),
				metadata.Pairs("authorization", "token")),
			expectedVerifyUserErr: nil,
			expectedCode:          codes.OK,
		},
		{
			name:   "successful_unary_check_auth_token_false",
			method: "/shortener.ShortenerService/GetUsersURLs",
			userID: "test",
			token:  "token",
			ctx: metadata.NewIncomingContext(context.Background(),
				metadata.Pairs("authorization", "token")),
			expectedVerifyUserErr: errorscustom.ErrBadVarifyToken,
			expectedCode:          codes.Unauthenticated,
		},
		{
			name:   "incorrect_token",
			method: "/shortener.ShortenerService/GetUsersURLs",
			userID: "test",
			token:  "token",
			ctx: metadata.NewIncomingContext(context.Background(),
				metadata.Pairs("authorization", "")),
			expectedCode: codes.Unauthenticated,
		},
	}

	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			logger := logger2.NewLogger()

			authServiceMock := auth.NewMockAuthService(ctrl)
			authServiceMock.EXPECT().
				VerifyUser(gomock.Any()).
				Return(cc.userID, cc.expectedVerifyUserErr).AnyTimes()
			authServiceMock.EXPECT().
				CreatTokenForUser(gomock.Any()).
				Return(cc.token, cc.expectedCreatTokenForUserErr).AnyTimes()
			authorization := NewAuthMiddleware(authServiceMock, logger)

			req := "test-request"

			info := &grpc.UnaryServerInfo{
				FullMethod: cc.method,
			}

			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				userID := ctx.Value(UserIDContextKey)
				return map[string]interface{}{"userID": userID}, nil
			}

			_, err := authorization.UnaryInterceptor(cc.ctx, req, info, handler)

			if err != nil {
				code, ok := status.FromError(err)
				if !ok {
					t.Errorf("unexpected error type: %v", err)
				}
				if code.Code() != cc.expectedCode {
					t.Errorf("unexpected error code: got %v, want %v", code.Code(), cc.expectedCode)
				}

			} else if cc.expectedCode != codes.OK {
				t.Errorf("expected error code %v, got none", cc.expectedCode)
			}
		})
	}
}
