// Code generated by mockery v2.30.1. DO NOT EDIT.

package mock

import (
	context "context"
	time "time"

	v3ioc "github.com/v3io/v3io-go/pkg/controlplane"

	mock "github.com/stretchr/testify/mock"
)

// Session is an autogenerated mock type for the Session type
type Session struct {
	mock.Mock
}

// CreateAccessKeySync provides a mock function with given fields: _a0
func (_m *Session) CreateAccessKeySync(_a0 *v3ioc.CreateAccessKeyInput) (*v3ioc.CreateAccessKeyOutput, error) {
	ret := _m.Called(_a0)

	var r0 *v3ioc.CreateAccessKeyOutput
	var r1 error
	if rf, ok := ret.Get(0).(func(*v3ioc.CreateAccessKeyInput) (*v3ioc.CreateAccessKeyOutput, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(*v3ioc.CreateAccessKeyInput) *v3ioc.CreateAccessKeyOutput); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v3ioc.CreateAccessKeyOutput)
		}
	}

	if rf, ok := ret.Get(1).(func(*v3ioc.CreateAccessKeyInput) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateContainerSync provides a mock function with given fields: _a0
func (_m *Session) CreateContainerSync(_a0 *v3ioc.CreateContainerInput) (*v3ioc.CreateContainerOutput, error) {
	ret := _m.Called(_a0)

	var r0 *v3ioc.CreateContainerOutput
	var r1 error
	if rf, ok := ret.Get(0).(func(*v3ioc.CreateContainerInput) (*v3ioc.CreateContainerOutput, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(*v3ioc.CreateContainerInput) *v3ioc.CreateContainerOutput); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v3ioc.CreateContainerOutput)
		}
	}

	if rf, ok := ret.Get(1).(func(*v3ioc.CreateContainerInput) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateEventSync provides a mock function with given fields: _a0
func (_m *Session) CreateEventSync(_a0 *v3ioc.CreateEventInput) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*v3ioc.CreateEventInput) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateUserSync provides a mock function with given fields: _a0
func (_m *Session) CreateUserSync(_a0 *v3ioc.CreateUserInput) (*v3ioc.CreateUserOutput, error) {
	ret := _m.Called(_a0)

	var r0 *v3ioc.CreateUserOutput
	var r1 error
	if rf, ok := ret.Get(0).(func(*v3ioc.CreateUserInput) (*v3ioc.CreateUserOutput, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(*v3ioc.CreateUserInput) *v3ioc.CreateUserOutput); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v3ioc.CreateUserOutput)
		}
	}

	if rf, ok := ret.Get(1).(func(*v3ioc.CreateUserInput) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteAccessKeySync provides a mock function with given fields: _a0
func (_m *Session) DeleteAccessKeySync(_a0 *v3ioc.DeleteAccessKeyInput) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*v3ioc.DeleteAccessKeyInput) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteContainerSync provides a mock function with given fields: _a0
func (_m *Session) DeleteContainerSync(_a0 *v3ioc.DeleteContainerInput) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*v3ioc.DeleteContainerInput) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteUserSync provides a mock function with given fields: _a0
func (_m *Session) DeleteUserSync(_a0 *v3ioc.DeleteUserInput) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*v3ioc.DeleteUserInput) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetAppServicesManifests provides a mock function with given fields: _a0
func (_m *Session) GetAppServicesManifests(_a0 *v3ioc.GetAppServicesManifestsInput) (*v3ioc.GetAppServicesManifestsOutput, error) {
	ret := _m.Called(_a0)

	var r0 *v3ioc.GetAppServicesManifestsOutput
	var r1 error
	if rf, ok := ret.Get(0).(func(*v3ioc.GetAppServicesManifestsInput) (*v3ioc.GetAppServicesManifestsOutput, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(*v3ioc.GetAppServicesManifestsInput) *v3ioc.GetAppServicesManifestsOutput); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v3ioc.GetAppServicesManifestsOutput)
		}
	}

	if rf, ok := ret.Get(1).(func(*v3ioc.GetAppServicesManifestsInput) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetJob provides a mock function with given fields: getJobsInput
func (_m *Session) GetJob(getJobsInput *v3ioc.GetJobInput) (*v3ioc.GetJobOutput, error) {
	ret := _m.Called(getJobsInput)

	var r0 *v3ioc.GetJobOutput
	var r1 error
	if rf, ok := ret.Get(0).(func(*v3ioc.GetJobInput) (*v3ioc.GetJobOutput, error)); ok {
		return rf(getJobsInput)
	}
	if rf, ok := ret.Get(0).(func(*v3ioc.GetJobInput) *v3ioc.GetJobOutput); ok {
		r0 = rf(getJobsInput)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v3ioc.GetJobOutput)
		}
	}

	if rf, ok := ret.Get(1).(func(*v3ioc.GetJobInput) error); ok {
		r1 = rf(getJobsInput)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetRunningUserAttributesSync provides a mock function with given fields: _a0
func (_m *Session) GetRunningUserAttributesSync(_a0 *v3ioc.GetRunningUserAttributesInput) (*v3ioc.GetRunningUserAttributesOutput, error) {
	ret := _m.Called(_a0)

	var r0 *v3ioc.GetRunningUserAttributesOutput
	var r1 error
	if rf, ok := ret.Get(0).(func(*v3ioc.GetRunningUserAttributesInput) (*v3ioc.GetRunningUserAttributesOutput, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(*v3ioc.GetRunningUserAttributesInput) *v3ioc.GetRunningUserAttributesOutput); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v3ioc.GetRunningUserAttributesOutput)
		}
	}

	if rf, ok := ret.Get(1).(func(*v3ioc.GetRunningUserAttributesInput) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ReloadAppServicesConfig provides a mock function with given fields: ctx
func (_m *Session) ReloadAppServicesConfig(ctx context.Context) (string, error) {
	ret := _m.Called(ctx)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (string, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) string); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ReloadAppServicesConfigAndWaitForCompletion provides a mock function with given fields: ctx, retryInterval, timeout
func (_m *Session) ReloadAppServicesConfigAndWaitForCompletion(ctx context.Context, retryInterval time.Duration, timeout time.Duration) error {
	ret := _m.Called(ctx, retryInterval, timeout)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, time.Duration, time.Duration) error); ok {
		r0 = rf(ctx, retryInterval, timeout)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ReloadArtifactVersionManifest provides a mock function with given fields: ctx
func (_m *Session) ReloadArtifactVersionManifest(ctx context.Context) (string, error) {
	ret := _m.Called(ctx)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (string, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) string); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ReloadArtifactVersionManifestAndWaitForCompletion provides a mock function with given fields: ctx, retryInterval, timeout
func (_m *Session) ReloadArtifactVersionManifestAndWaitForCompletion(ctx context.Context, retryInterval time.Duration, timeout time.Duration) error {
	ret := _m.Called(ctx, retryInterval, timeout)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, time.Duration, time.Duration) error); ok {
		r0 = rf(ctx, retryInterval, timeout)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ReloadClusterConfig provides a mock function with given fields: ctx
func (_m *Session) ReloadClusterConfig(ctx context.Context) (string, error) {
	ret := _m.Called(ctx)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (string, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) string); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ReloadClusterConfigAndWaitForCompletion provides a mock function with given fields: ctx, retryInterval, timeout
func (_m *Session) ReloadClusterConfigAndWaitForCompletion(ctx context.Context, retryInterval time.Duration, timeout time.Duration) error {
	ret := _m.Called(ctx, retryInterval, timeout)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, time.Duration, time.Duration) error); ok {
		r0 = rf(ctx, retryInterval, timeout)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ReloadEventsConfig provides a mock function with given fields: ctx
func (_m *Session) ReloadEventsConfig(ctx context.Context) (string, error) {
	ret := _m.Called(ctx)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (string, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) string); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ReloadEventsConfigAndWaitForCompletion provides a mock function with given fields: ctx, retryInterval, timeout
func (_m *Session) ReloadEventsConfigAndWaitForCompletion(ctx context.Context, retryInterval time.Duration, timeout time.Duration) error {
	ret := _m.Called(ctx, retryInterval, timeout)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, time.Duration, time.Duration) error); ok {
		r0 = rf(ctx, retryInterval, timeout)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateAppServicesManifest provides a mock function with given fields: _a0
func (_m *Session) UpdateAppServicesManifest(_a0 *v3ioc.UpdateAppServicesManifestInput) (*v3ioc.GetJobOutput, error) {
	ret := _m.Called(_a0)

	var r0 *v3ioc.GetJobOutput
	var r1 error
	if rf, ok := ret.Get(0).(func(*v3ioc.UpdateAppServicesManifestInput) (*v3ioc.GetJobOutput, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(*v3ioc.UpdateAppServicesManifestInput) *v3ioc.GetJobOutput); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v3ioc.GetJobOutput)
		}
	}

	if rf, ok := ret.Get(1).(func(*v3ioc.UpdateAppServicesManifestInput) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateClusterInfoSync provides a mock function with given fields: _a0
func (_m *Session) UpdateClusterInfoSync(_a0 *v3ioc.UpdateClusterInfoInput) (*v3ioc.UpdateClusterInfoOutput, error) {
	ret := _m.Called(_a0)

	var r0 *v3ioc.UpdateClusterInfoOutput
	var r1 error
	if rf, ok := ret.Get(0).(func(*v3ioc.UpdateClusterInfoInput) (*v3ioc.UpdateClusterInfoOutput, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(*v3ioc.UpdateClusterInfoInput) *v3ioc.UpdateClusterInfoOutput); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v3ioc.UpdateClusterInfoOutput)
		}
	}

	if rf, ok := ret.Get(1).(func(*v3ioc.UpdateClusterInfoInput) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// WaitForJobCompletion provides a mock function with given fields: ctx, jobID, retryInterval, timeout
func (_m *Session) WaitForJobCompletion(ctx context.Context, jobID string, retryInterval time.Duration, timeout time.Duration) error {
	ret := _m.Called(ctx, jobID, retryInterval, timeout)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, time.Duration, time.Duration) error); ok {
		r0 = rf(ctx, jobID, retryInterval, timeout)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewSession creates a new instance of Session. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSession(t interface {
	mock.TestingT
	Cleanup(func())
}) *Session {
	mock := &Session{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
