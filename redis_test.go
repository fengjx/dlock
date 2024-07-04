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
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	testCases := []struct {
		name       string
		key        string
		ttl        time.Duration
		timeout    time.Duration
		tries      int
		getMockCli func() redis.Cmdable
		wantErr    error
	}{
		{
			name:    "test-lock",
			key:     "test",
			ttl:     time.Second * 10,
			timeout: time.Second * 3,
			tries:   3,
			getMockCli: func() redis.Cmdable {
				cli := mocks.NewMockCmdable(ctrl)
				cli.EXPECT().ScriptLoad(gomock.Any(), luaUnlock).
					Return(redis.NewStringResult("mock-unlock-sha1", nil))

				cli.EXPECT().SetNX(gomock.Any(), "test-lock", gomock.Any(), time.Second*10).
					Return(redis.NewBoolResult(true, nil)).
					AnyTimes()
				return cli
			},
			wantErr: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cli := tc.getMockCli()
			lockClient := NewRedisLock(cli)
			mutex := lockClient.NewMutex(
				tc.name,
				WithTTL(tc.ttl),
				WithTimeout(tc.timeout),
				WithTries(tc.tries),
			)
			err := mutex.Lock()
			if tc.wantErr != nil {
				assert.ErrorAs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
