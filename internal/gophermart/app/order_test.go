package app_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/StasMerzlyakov/gophermart/internal/gophermart/app"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/app/mocks"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestNewNoErr(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	userID := 1

	var err error
	ctx, err = domain.EnrichWithAuthData(ctx, &domain.AuthData{
		UserID: userID,
	})

	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockOrderStorage(ctrl)

	number := domain.OrderNumber("5062821234567892")

	mockStorage.EXPECT().GetOrder(gomock.Any()).DoAndReturn(func(nm domain.OrderNumber) (*domain.OrderData, error) {
		require.Equal(t, number, nm)
		return nil, nil
	})

	mockStorage.EXPECT().Upload(gomock.Any()).DoAndReturn(func(oData *domain.OrderData) error {
		require.NotNil(t, oData)
		require.Equal(t, userID, oData.UserID)
		require.Equal(t, number, oData.Number)
		require.Equal(t, domain.OrderStratusNew, oData.Status)
		require.Nil(t, oData.Accrual)
		return nil
	})

	order := app.NewOrder(mockStorage)

	err = order.New(ctx, number)

	require.NoError(t, err)
}

func TestNewErrOrderNumberAlreadyProcessed(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	userID := 1

	var err error
	ctx, err = domain.EnrichWithAuthData(ctx, &domain.AuthData{
		UserID: userID,
	})

	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockOrderStorage(ctrl)

	number := domain.OrderNumber("5062821234567892")

	mockStorage.EXPECT().GetOrder(gomock.Any()).DoAndReturn(func(nm domain.OrderNumber) (*domain.OrderData, error) {
		require.Equal(t, number, nm)
		return &domain.OrderData{
			UserID:     userID,
			Number:     number,
			Status:     domain.OrderStratusProcessed,
			Accrual:    domain.Float64Ptr(60.),
			UploadedAt: domain.RFC3339Time(time.Now()),
		}, nil
	})

	order := app.NewOrder(mockStorage)

	err = order.New(ctx, number)

	require.ErrorIs(t, err, domain.ErrOrderNumberAlreadyUploaded)
}

func TestNewErrDublicateOrderNumber(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	userID := 1

	var err error
	ctx, err = domain.EnrichWithAuthData(ctx, &domain.AuthData{
		UserID: userID,
	})

	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockOrderStorage(ctrl)

	number := domain.OrderNumber("5062821234567892")

	mockStorage.EXPECT().GetOrder(gomock.Any()).DoAndReturn(func(nm domain.OrderNumber) (*domain.OrderData, error) {
		require.Equal(t, number, nm)
		return &domain.OrderData{
			UserID:     userID + 1,
			Number:     number,
			Status:     domain.OrderStratusProcessed,
			Accrual:    domain.Float64Ptr(60.),
			UploadedAt: domain.RFC3339Time(time.Now()),
		}, nil
	})

	order := app.NewOrder(mockStorage)

	err = order.New(ctx, number)

	require.ErrorIs(t, err, domain.ErrDublicateOrderNumber)
}

func TestNewErrUserIsNotAuthorized(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockOrderStorage(ctrl)

	number := domain.OrderNumber("5062821234567892")

	order := app.NewOrder(mockStorage)

	err := order.New(ctx, number)

	require.ErrorIs(t, err, domain.ErrUserIsNotAuthorized)
}

func TestNewErrServerInternal(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	userID := 1

	var err error
	ctx, err = domain.EnrichWithAuthData(ctx, &domain.AuthData{
		UserID: userID,
	})

	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockOrderStorage(ctrl)

	number := domain.OrderNumber("5062821234567892")

	mockStorage.EXPECT().GetOrder(gomock.Any()).DoAndReturn(func(nm domain.OrderNumber) (*domain.OrderData, error) {
		require.Equal(t, number, nm)
		return nil, errors.New("any err")
	})

	order := app.NewOrder(mockStorage)

	err = order.New(ctx, number)

	require.ErrorIs(t, err, domain.ErrServerInternal)
}

func TestNewErrWrongOrderNumber(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	userID := 1

	var err error
	ctx, err = domain.EnrichWithAuthData(ctx, &domain.AuthData{
		UserID: userID,
	})

	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockOrderStorage(ctrl)

	number := domain.OrderNumber("50628212345678921")

	order := app.NewOrder(mockStorage)

	err = order.New(ctx, number)

	require.ErrorIs(t, err, domain.ErrWrongOrderNumber)
}

func TestAllErrNoErr(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	userID := 1

	var err error
	ctx, err = domain.EnrichWithAuthData(ctx, &domain.AuthData{
		UserID: userID,
	})

	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockOrderStorage(ctrl)

	ordrs := []domain.OrderData{
		{
			UserID:     userID,
			Number:     "5062821234567891",
			Status:     domain.OrderStratusNew,
			UploadedAt: domain.RFC3339Time(time.Now()),
		},
		{
			UserID:     userID,
			Number:     "5062821234567892",
			Status:     domain.OrderStratusProcessed,
			Accrual:    domain.Float64Ptr(10.),
			UploadedAt: domain.RFC3339Time(time.Now()),
		},
	}

	mockStorage.EXPECT().Orders(gomock.Any()).DoAndReturn(func(uID int) ([]domain.OrderData, error) {
		require.Equal(t, userID, uID)
		return ordrs, nil
	}).Times(1)

	order := app.NewOrder(mockStorage)

	res, err := order.All(ctx)

	require.Nil(t, err)

	require.True(t, reflect.DeepEqual(ordrs, res))
}

func TestAllErrUserIsNotAuthorized(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockOrderStorage(ctrl)

	order := app.NewOrder(mockStorage)

	res, err := order.All(ctx)

	require.Nil(t, res)
	require.ErrorIs(t, err, domain.ErrUserIsNotAuthorized)
}

func TestAllErrNotFound(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	userID := 1

	var err error
	ctx, err = domain.EnrichWithAuthData(ctx, &domain.AuthData{
		UserID: userID,
	})

	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockOrderStorage(ctrl)

	mockStorage.EXPECT().Orders(gomock.Any()).DoAndReturn(func(uID int) ([]domain.OrderData, error) {
		require.Equal(t, userID, uID)
		return nil, nil
	})

	order := app.NewOrder(mockStorage)

	res, err := order.All(ctx)

	require.Nil(t, res)

	require.ErrorIs(t, err, domain.ErrNotFound)
}

func TestAllErrServerInternal(t *testing.T) {
	ctx := EnrichTestContext(context.Background())

	userID := 1

	var err error
	ctx, err = domain.EnrichWithAuthData(ctx, &domain.AuthData{
		UserID: userID,
	})

	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockOrderStorage(ctrl)

	mockStorage.EXPECT().Orders(gomock.Any()).DoAndReturn(func(uID int) ([]domain.OrderData, error) {
		require.Equal(t, userID, uID)
		return nil, errors.New("any err")
	})

	order := app.NewOrder(mockStorage)

	res, err := order.All(ctx)

	require.Nil(t, res)

	require.ErrorIs(t, err, domain.ErrServerInternal)
}
