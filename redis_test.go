package dlock

import (
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/fengjx/dlock/mocks"
)

func TestLock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	testCases := []struct {
		name             string
		ttl              time.Duration
		timeout          time.Duration
		tries            int
		genValueFn       GenValueFn
		getMockCli       func() redis.Cmdable
		wantLockErr      error
		wantUnlockResult bool
		wantUnlockErr    error
		tryLock          bool
	}{
		{
			name:    "test-lock",
			ttl:     time.Second * 10,
			timeout: time.Second * 3,
			tries:   3,
			getMockCli: func() redis.Cmdable {
				cli := mocks.NewMockCmdable(ctrl)
				sha1 := "mock-unlock-sha1"
				cli.EXPECT().ScriptLoad(gomock.Any(), luaUnlock).
					Return(redis.NewStringResult(sha1, nil))

				cli.EXPECT().EvalSha(gomock.Any(), sha1, []string{"test-lock"}, gomock.Any()).
					Return(redis.NewCmdResult(int64(1), nil))

				cli.EXPECT().SetNX(gomock.Any(), "test-lock", gomock.Any(), time.Second*10).
					Return(redis.NewBoolResult(true, nil)).
					AnyTimes()
				return cli
			},
			wantUnlockResult: true,
		},
		{
			name:    "test-lock-failed",
			ttl:     time.Second * 10,
			timeout: time.Second * 3,
			tries:   3,
			getMockCli: func() redis.Cmdable {
				cli := mocks.NewMockCmdable(ctrl)

				sha1 := "mock-unlock-sha1"
				cli.EXPECT().ScriptLoad(gomock.Any(), luaUnlock).
					Return(redis.NewStringResult(sha1, nil))

				cli.EXPECT().SetNX(gomock.Any(), "test-lock-failed", gomock.Any(), time.Second*10).
					Return(redis.NewBoolResult(false, nil)).
					AnyTimes()
				return cli
			},
			wantLockErr: ErrTimeout,
		},
		{
			name:    "test-unlock-failed",
			ttl:     time.Second * 10,
			timeout: time.Second * 3,
			tries:   3,
			getMockCli: func() redis.Cmdable {
				cli := mocks.NewMockCmdable(ctrl)
				sha1 := "test-unlock-failed-sha1"
				cli.EXPECT().ScriptLoad(gomock.Any(), luaUnlock).
					Return(redis.NewStringResult(sha1, nil))

				cli.EXPECT().EvalSha(gomock.Any(), sha1, []string{"test-unlock-failed"}, gomock.Any()).
					Return(redis.NewCmdResult(int64(0), nil))

				cli.EXPECT().SetNX(gomock.Any(), "test-unlock-failed", gomock.Any(), time.Second*10).
					Return(redis.NewBoolResult(true, nil)).
					AnyTimes()
				return cli
			},
			wantUnlockResult: false,
		},
		{
			name:    "test-unlock-expired",
			ttl:     time.Second * 10,
			timeout: time.Second * 3,
			tries:   3,
			getMockCli: func() redis.Cmdable {
				cli := mocks.NewMockCmdable(ctrl)
				sha1 := "test-unlock-expired-sha1"
				cli.EXPECT().ScriptLoad(gomock.Any(), luaUnlock).
					Return(redis.NewStringResult(sha1, nil))

				cli.EXPECT().EvalSha(gomock.Any(), sha1, []string{"test-unlock-expired"}, gomock.Any()).
					Return(redis.NewCmdResult(int64(-1), nil))

				cli.EXPECT().SetNX(gomock.Any(), "test-unlock-expired", gomock.Any(), time.Second*10).
					Return(redis.NewBoolResult(true, nil)).
					AnyTimes()
				return cli
			},
			wantUnlockResult: false,
			wantUnlockErr:    ErrLockNotExistOrExpired,
		},
		{
			name:    "test-try-lock",
			ttl:     time.Second * 10,
			timeout: time.Second * 3,
			tries:   3,
			tryLock: true,
			getMockCli: func() redis.Cmdable {
				cli := mocks.NewMockCmdable(ctrl)
				sha1 := "mock-unlock-sha1"
				cli.EXPECT().ScriptLoad(gomock.Any(), luaUnlock).
					Return(redis.NewStringResult(sha1, nil))

				cli.EXPECT().EvalSha(gomock.Any(), sha1, []string{"test-try-lock"}, gomock.Any()).
					Return(redis.NewCmdResult(int64(1), nil))

				cli.EXPECT().SetNX(gomock.Any(), "test-try-lock", gomock.Any(), time.Second*10).
					Return(redis.NewBoolResult(true, nil)).
					AnyTimes()
				return cli
			},
			wantUnlockResult: true,
		},
		{
			name:    "test-try-lock-failed",
			ttl:     time.Second * 10,
			timeout: time.Second * 3,
			tries:   3,
			tryLock: true,
			getMockCli: func() redis.Cmdable {
				cli := mocks.NewMockCmdable(ctrl)
				sha1 := "test-try-lock-failed-sha1"
				cli.EXPECT().ScriptLoad(gomock.Any(), luaUnlock).
					Return(redis.NewStringResult(sha1, nil))

				cli.EXPECT().SetNX(gomock.Any(), "test-try-lock-failed", gomock.Any(), time.Second*10).
					Return(redis.NewBoolResult(false, nil)).
					AnyTimes()
				return cli
			},
			wantLockErr: ErrFailed,
		},
		{
			name:    "test-try-unlock-failed",
			ttl:     time.Second * 10,
			timeout: time.Second * 3,
			tries:   3,
			tryLock: true,
			getMockCli: func() redis.Cmdable {
				cli := mocks.NewMockCmdable(ctrl)
				sha1 := "test-try-unlock-failed-sha1"
				cli.EXPECT().ScriptLoad(gomock.Any(), luaUnlock).
					Return(redis.NewStringResult(sha1, nil))

				cli.EXPECT().EvalSha(gomock.Any(), sha1, []string{"test-try-unlock-failed"}, gomock.Any()).
					Return(redis.NewCmdResult(int64(0), nil))

				cli.EXPECT().SetNX(gomock.Any(), "test-try-unlock-failed", gomock.Any(), time.Second*10).
					Return(redis.NewBoolResult(true, nil)).
					AnyTimes()
				return cli
			},
			wantUnlockResult: false,
		},
		{
			name:    "test-try-unlock-expired",
			ttl:     time.Second * 10,
			timeout: time.Second * 3,
			tries:   3,
			tryLock: true,
			getMockCli: func() redis.Cmdable {
				cli := mocks.NewMockCmdable(ctrl)
				sha1 := "test-try-unlock-expired-sha1"
				cli.EXPECT().ScriptLoad(gomock.Any(), luaUnlock).
					Return(redis.NewStringResult(sha1, nil))

				cli.EXPECT().EvalSha(gomock.Any(), sha1, []string{"test-try-unlock-expired"}, gomock.Any()).
					Return(redis.NewCmdResult(int64(-1), nil))

				cli.EXPECT().SetNX(gomock.Any(), "test-try-unlock-expired", gomock.Any(), time.Second*10).
					Return(redis.NewBoolResult(true, nil)).
					AnyTimes()
				return cli
			},
			wantUnlockResult: false,
			wantUnlockErr:    ErrLockNotExistOrExpired,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cli := tc.getMockCli()
			lockClient := NewRedisLock(cli)
			var opts []Option
			if tc.ttl != 0 {
				opts = append(opts, WithTTL(tc.ttl))
			}
			if tc.tries != 0 {
				opts = append(opts, WithTries(tc.tries))
			}
			if tc.timeout != 0 {
				opts = append(opts, WithTimeout(tc.timeout))
			}
			if tc.genValueFn != nil {
				opts = append(opts, WithGenValueFunc(tc.genValueFn))
			}
			mutex := lockClient.NewMutex(tc.name, opts...)
			var err error
			if tc.tryLock {
				err = mutex.TryLock()
			} else {
				err = mutex.Lock()
			}
			if tc.wantLockErr != nil {
				assert.Equal(t, tc.wantLockErr, err)
				return
			}
			require.NoError(t, err)
			ok, err := mutex.Unlock()
			assert.Equal(t, tc.wantUnlockResult, ok)
			if tc.wantUnlockErr != nil {
				assert.Equal(t, tc.wantUnlockErr, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
