package auth

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticator_Authorize(t *testing.T) {
	type args struct {
		username string
		password string
	}
	testCases := []struct {
		name    string
		fake    func(t *testing.T) (Dependencies, func())
		args    args
		want    Credentials
		wantErr bool
	}{
		{
			name: "success-normal",
			fake: func(t *testing.T) (Dependencies, func()) {
				ctrl := gomock.NewController(t)
				mockConfig := NewMockConfig(ctrl)
				mockConfig.EXPECT().GetUsername().Return("user", nil).AnyTimes()
				mockConfig.EXPECT().GetPassword().Return("password", nil).AnyTimes()
				mockConfig.EXPECT().GetUserToken().Return("token", nil).AnyTimes()
				mockCipher := NewMockCipher(ctrl)
				mockCipher.EXPECT().Check(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				mockTokenOperator := NewMockTokenOperator(ctrl)
				gomock.InOrder(
					mockTokenOperator.EXPECT().Generate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("access", nil),
					mockTokenOperator.EXPECT().Generate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("refresh", nil),
				)
				return Dependencies{
					Config:        mockConfig,
					Cipher:        mockCipher,
					TokenOperator: mockTokenOperator,
				}, ctrl.Finish
			},
			args: args{
				username: "user",
				password: "password",
			},
			want: Credentials{
				AccessToken:  "access",
				RefreshToken: "refresh",
			},
		},
		{
			name: "fail-empty-username",
			fake: func(t *testing.T) (Dependencies, func()) {
				ctrl := gomock.NewController(t)
				mockConfig := NewMockConfig(ctrl)
				mockConfig.EXPECT().GetUserToken().Return("token", nil).AnyTimes()
				return Dependencies{Config: mockConfig}, ctrl.Finish
			},
			args:    args{},
			wantErr: true,
		},
		{
			name: "fail-username-not-match",
			fake: func(t *testing.T) (Dependencies, func()) {
				ctrl := gomock.NewController(t)
				mockConfig := NewMockConfig(ctrl)
				mockConfig.EXPECT().GetUserToken().Return("token", nil).AnyTimes()
				mockConfig.EXPECT().GetUsername().Return("user", nil).AnyTimes()
				return Dependencies{Config: mockConfig}, ctrl.Finish
			},
			args: args{
				username: "user1",
				password: "pwd",
			},
			wantErr: true,
		},
		{
			name: "fail-password-not-match",
			fake: func(t *testing.T) (Dependencies, func()) {
				ctrl := gomock.NewController(t)
				mockConfig := NewMockConfig(ctrl)
				mockConfig.EXPECT().GetUserToken().Return("token", nil).AnyTimes()
				mockConfig.EXPECT().GetUsername().Return("user", nil).AnyTimes()
				mockConfig.EXPECT().GetPassword().Return("password", nil).AnyTimes()
				mockCipher := NewMockCipher(ctrl)
				mockCipher.EXPECT().Check(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(assert.AnError).AnyTimes()
				return Dependencies{
					Config: mockConfig,
					Cipher: mockCipher,
				}, ctrl.Finish
			},
			args: args{
				username: "user",
				password: "password",
			},
			wantErr: true,
		},
		{
			name: "success-first-init",
			fake: func(t *testing.T) (Dependencies, func()) {
				ctrl := gomock.NewController(t)
				mockConfig := NewMockConfig(ctrl)
				mockConfig.EXPECT().GetUsername().Return("", nil).AnyTimes()
				mockConfig.EXPECT().SetUsername(defaultUsername).AnyTimes()
				mockConfig.EXPECT().GetPassword().Return("", nil).AnyTimes()
				mockConfig.EXPECT().SetPassword(defaultPassword).AnyTimes()
				mockConfig.EXPECT().GetUserToken().Return("", nil).AnyTimes()
				mockConfig.EXPECT().SetUserToken(gomock.Any()).AnyTimes()
				mockCipher := NewMockCipher(ctrl)
				mockCipher.EXPECT().Check(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				mockCipher.EXPECT().GenerateKey(gomock.Any()).Return("token", nil).AnyTimes()
				mockTokenOperator := NewMockTokenOperator(ctrl)
				gomock.InOrder(
					mockTokenOperator.EXPECT().Generate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("access", nil),
					mockTokenOperator.EXPECT().Generate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("refresh", nil),
				)
				return Dependencies{
					Config:        mockConfig,
					Cipher:        mockCipher,
					TokenOperator: mockTokenOperator,
				}, ctrl.Finish
			},
			args: args{
				username: "admin",
				password: "password",
			},
			want: Credentials{
				AccessToken:  "access",
				RefreshToken: "refresh",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dep, clo := tc.fake(t)
			defer clo()
			a := New(dep)

			got, err := a.Authorize(context.Background(), tc.args.username, tc.args.password)
			t.Log(err)

			assert.Equal(t, tc.wantErr, err != nil)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestAuthenticator_UpdateUser(t *testing.T) {
	type args struct {
		accessToken string
		username    string
		password    string
	}
	testCases := []struct {
		name    string
		fake    func(t *testing.T) (Dependencies, func())
		args    args
		wantErr bool
	}{
		{
			name: "success",
			fake: func(t *testing.T) (Dependencies, func()) {
				ctrl := gomock.NewController(t)
				mockConfig := NewMockConfig(ctrl)
				mockConfig.EXPECT().GetUserToken().Return("token", nil).AnyTimes()
				mockConfig.EXPECT().SetUsername(gomock.Any()).AnyTimes()
				mockConfig.EXPECT().SetPassword(gomock.Any()).AnyTimes()
				mockCipher := NewMockCipher(ctrl)
				mockCipher.EXPECT().Encrypt(gomock.Any(), gomock.Any(), gomock.Any()).Return("cipher", nil).AnyTimes()
				return Dependencies{
					Config: mockConfig,
					Cipher: mockCipher,
				}, ctrl.Finish
			},
			args: args{
				accessToken: "access",
				password:    "password",
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dep, clo := tc.fake(t)
			defer clo()
			a := New(dep)

			err := a.UpdateUser(context.Background(), tc.args.username, tc.args.password)
			t.Log(err)

			assert.Equal(t, tc.wantErr, err != nil)
		})
	}
}

func TestAuthenticator_CheckAccessToken(t *testing.T) {
	type args struct {
		accessToken string
	}
	testCases := []struct {
		name    string
		fake    func(t *testing.T) (Dependencies, func())
		args    args
		wantErr bool
	}{
		{
			name: "success",
			fake: func(t *testing.T) (Dependencies, func()) {
				ctrl := gomock.NewController(t)
				mockConfig := NewMockConfig(ctrl)
				mockConfig.EXPECT().GetUserToken().Return("token", nil).AnyTimes()
				mockTokenOperator := NewMockTokenOperator(ctrl)
				mockTokenOperator.EXPECT().Check(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
				return Dependencies{TokenOperator: mockTokenOperator, Config: mockConfig}, ctrl.Finish
			},
			args: args{
				accessToken: "access",
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dep, clo := tc.fake(t)
			defer clo()
			a := New(dep)

			err := a.CheckAccessToken(context.Background(), tc.args.accessToken)
			t.Log(err)

			assert.Equal(t, tc.wantErr, err != nil)
		})
	}
}

func TestAuthenticator_RefreshCredentials(t *testing.T) {
	type args struct {
		refreshToken string
	}
	testCases := []struct {
		name    string
		fake    func(t *testing.T) (Dependencies, func())
		args    args
		want    Credentials
		wantErr bool
	}{
		{
			name: "success",
			fake: func(t *testing.T) (Dependencies, func()) {
				ctrl := gomock.NewController(t)
				mockTokenOperator := NewMockTokenOperator(ctrl)
				mockTokenOperator.EXPECT().Check(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
				mockConfig := NewMockConfig(ctrl)
				mockConfig.EXPECT().GetUserToken().Return("token", nil).AnyTimes()
				mockCipher := NewMockCipher(ctrl)
				mockCipher.EXPECT().GenerateKey(gomock.Any()).Return("token", nil).AnyTimes()
				gomock.InOrder(
					mockTokenOperator.EXPECT().Generate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("access", nil),
					mockTokenOperator.EXPECT().Generate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("refresh", nil),
				)
				return Dependencies{
					TokenOperator: mockTokenOperator,
					Config:        mockConfig,
					Cipher:        mockCipher,
				}, ctrl.Finish
			},
			args: args{
				refreshToken: "refresh",
			},
			want: Credentials{
				AccessToken:  "access",
				RefreshToken: "refresh",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dep, clo := tc.fake(t)
			defer clo()
			a := New(dep)

			got, err := a.RefreshCredentials(context.Background(), tc.args.refreshToken)
			t.Log(err)

			assert.Equal(t, tc.wantErr, err != nil)
			assert.Equal(t, tc.want, got)
		})
	}
}
