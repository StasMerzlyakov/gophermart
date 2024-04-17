package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/StasMerzlyakov/gophermart/internal/config"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
)

func NewOrder(ctx context.Context, conf *config.GophermartConfig,
	storage OrderStorage, acrualSystem AcrualSystem) *order {
	ord := &order{
		storage:               storage,
		processingWorkerCount: conf.AcrualSystemPoolCount,
		acrualSystem:          acrualSystem,
	}

	return ord
}

//go:generate mockgen -destination "./mocks/$GOFILE" -package mocks . OrderStorage
type OrderStorage interface {
	Upload(ctx context.Context, data *domain.OrderData) error
	Orders(ctx context.Context, userID int) ([]domain.OrderData, error)
	Update(ctx context.Context, number domain.OrderNumber, status domain.OrderStatus, accrual *float64) error
	UpdateBatch(ctx context.Context, orders []domain.OrderData) error
	ForProcessing(ctx context.Context, statuses []domain.OrderStatus) ([]domain.OrderData, error)
}

//go:generate
type AcrualSystem interface {
	Update(ctx context.Context, orderNum domain.OrderNumber) (*domain.AccrualData, error)
}

type order struct {
	storage               OrderStorage
	acrualSystem          AcrualSystem
	processingWorkerCount int
}

// Загрузка пользователем номера заказа для расчёта
// Возвращает:
//  - nil -  новый номер заказа принят в обработку
//  - domain.ErrOrderNumberAlreadyProcessed -  номер заказа уже был загружен этим пользователем;
//  - domain.ErrWrongOrderNumber - неверный формат номера заказа
//  - domain.ErrUserIsNotAuthorized - пользователь не авторизован
//  - domain.ErrDublicateOrderNumber - номер заказа уже был загружен другим пользователем
//  - domain.ErrServerInternal - внутренняя ошибка

func (ord *order) New(ctx context.Context, number domain.OrderNumber) error {
	logger, err := domain.GetLogger(ctx)
	if err != nil {
		log.Printf("%v: can't upload - logger not found in context", domain.ErrServerInternal)
		return fmt.Errorf("%w: logger not found in context", domain.ErrServerInternal)
	}

	userID, err := domain.GetUserID(ctx)
	if err != nil {
		logger.Errorw("order.New", "err", err.Error())
		return domain.ErrUserIsNotAuthorized
	}

	if !domain.CheckLuhn(number) {
		logger.Errorw("order.New", "err", fmt.Sprintf("%v: %v wrong value", domain.ErrWrongOrderNumber.Error(), number))
		return fmt.Errorf("%w: %v wrong value", domain.ErrWrongOrderNumber, number)
	}

	orderData := &domain.OrderData{
		UserID:     userID,
		Number:     number,
		Status:     domain.OrderStratusNew,
		UploadedAt: domain.RFC3339Time(time.Now()),
	}

	err = ord.storage.Upload(ctx, orderData)
	if err != nil {
		logger.Infow("order.New", "err", err.Error())

		if errors.Is(err, domain.ErrOrderNumberAlreadyUploaded) || errors.Is(err, domain.ErrDublicateOrderNumber) {
			return err
		}

		return fmt.Errorf("%w: %v", domain.ErrServerInternal, err.Error())
	}

	return nil
}

// Получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях;
// Возвращает ошибки:
//   - domain.ErrNotFound
//   - domain.ErrUserIsNotAuthorized
//   - domain.ErrServerInternal
func (ord *order) All(ctx context.Context) ([]domain.OrderData, error) {
	logger, err := domain.GetLogger(ctx)
	if err != nil {
		log.Printf("%v: can't upload - logger not found in context", domain.ErrServerInternal)
		return nil, fmt.Errorf("%w: logger not found in context", domain.ErrServerInternal)
	}

	userID, err := domain.GetUserID(ctx)
	if err != nil {
		logger.Errorw("order.All", "err", err.Error())
		return nil, domain.ErrUserIsNotAuthorized
	}

	data, err := ord.storage.Orders(ctx, userID)
	if err != nil {
		logger.Errorw("order.All", "err", err.Error())
		return nil, fmt.Errorf("%w: %v", domain.ErrServerInternal, err.Error())
	}

	if data == nil {
		logger.Errorw("order.All", "err", domain.ErrNotFound.Error())
		return nil, domain.ErrNotFound
	}

	return data, nil
}

func (ord *order) refreshOrderStatus(ctx context.Context,
	logger domain.Logger,
	ordNumChan <-chan domain.OrderNumber,
	acrualDataChan chan<- *domain.AccrualData) {

	// Пытаюсь осмыслить https://go.dev/blog/io2013-talk-concurrency
	var ordNumInternalChan <-chan domain.OrderNumber
	var sleepChan <-chan time.Time

	for {
		select {
		case <-ctx.Done():
			logger.Infow("app.refreshOrder", "status", "complete")
			return
		case <-sleepChan:
			// Подождали 5 секунд, можно запускать обработку
			ordNumInternalChan = ordNumChan
			sleepChan = nil
		case orderNum := <-ordNumInternalChan:
			acrData, err := ord.acrualSystem.Update(ctx, orderNum)
			if err != nil {
				// Произошла ошибка - запускаем sleepChan канал в надежде на восстановление
				logger.Infow("order.refreshOrder", "num", orderNum, "err", err.Error())
				sleepChan = time.After(5 * time.Second)
				ordNumInternalChan = nil
			} else {
				acrualDataChan <- acrData
			}
		}
	}
}

func (ord *order) orderStatusUpdater(ctx context.Context, logger domain.Logger, acrualDataChan <-chan *domain.AccrualData) {
	// Будем обновлять данные пачками либо по 10 записей либо через каждые 2 секунды
	var orders []domain.OrderData

	// Пытаюсь осмыслить https://go.dev/blog/io2013-talk-concurrency
	var waitAcrualsTimeoutChan <-chan time.Time
	var sleepAfterErrChan <-chan time.Time
	var acrualDataInternalChan <-chan *domain.AccrualData = acrualDataChan
	for {
		select {
		case <-ctx.Done():
			logger.Infow("app.orderStatusUpdater", "status", "complete")
			return
		case <-sleepAfterErrChan:
			waitAcrualsTimeoutChan = time.After(2 * time.Second)
			sleepAfterErrChan = nil
		case <-waitAcrualsTimeoutChan:
			if len(orders) > 0 {
				err := ord.storage.UpdateBatch(ctx, orders)
				if err != nil {
					// Произошла ошибка - запускаем sleepChan канал в надежде на восстановление
					logger.Infow("order.UpdateBatch", "err", err.Error())
					sleepAfterErrChan = time.After(5 * time.Second)
					acrualDataInternalChan = nil
				}
				clear(orders)
			}
		case acrualData := <-acrualDataInternalChan:
			switch acrualData.Status {
			case domain.AccrualStatusInvalid:
				orders = append(orders, domain.OrderData{
					Number: acrualData.Number,
					Status: domain.OrderStratusInvalid,
				})
			case domain.AccrualStatusProcessed:
				orders = append(orders, domain.OrderData{
					Number:  acrualData.Number,
					Status:  domain.OrderStratusProcessing,
					Accrual: acrualData.Accrual,
				})
			}
		}
	}
}

func (ord *order) PoolAcrualSystem(ctx context.Context) {
	logger, err := domain.GetLogger(ctx)
	if err != nil {
		fmt.Printf("app.orderPool error - logger not found")
		return
	}
	var wg sync.WaitGroup
	defer wg.Done()

	ordNumChan := make(chan domain.OrderNumber)
	acrualDataChan := make(chan *domain.AccrualData)

	wg.Add(1)
	go func() {
		defer wg.Done()
		ord.orderStatusUpdater(ctx, logger, acrualDataChan)
	}()

	for i := 0; i < ord.processingWorkerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ord.refreshOrderStatus(ctx, logger, ordNumChan, acrualDataChan)
		}()
	}

Loop:
	for {
		orders, err := ord.storage.ForProcessing(ctx, []domain.OrderStatus{domain.OrderStratusNew})
		if err != nil {
			logger.Errorw("app.orderPool", "err", err.Error())
			select {
			case <-ctx.Done():
				logger.Infow("app.orderPool", "status", "complete")
				close(ordNumChan)
				break Loop
			case <-time.After(5 * time.Second): // Возможно все восстановится за это время
				continue
			}
		}

		for _, ord := range orders {
			select {
			case <-ctx.Done():
				logger.Infow("app.orderPool", "status", "complete")
				close(ordNumChan)
				return
			case ordNumChan <- ord.Number:
			}
		}
	}
}