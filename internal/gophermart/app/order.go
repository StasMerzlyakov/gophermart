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

func NewOrder(conf *config.GophermartConfig,
	storage OrderStorage, acrualSystem AcrualSystem) *order {
	ord := &order{
		storage:               storage,
		processingWorkerCount: conf.AcrualSystemPoolCount,
		acrualSystem:          acrualSystem,
		batchSize:             conf.BatchSize,
	}

	return ord
}

//go:generate mockgen -destination "./mocks/$GOFILE" -package mocks . OrderStorage,AcrualSystem
type OrderStorage interface {
	Upload(ctx context.Context, data *domain.OrderData) error
	Orders(ctx context.Context, userID int) ([]domain.OrderData, error)
	UpdateOrder(ctx context.Context, number domain.OrderNumber, status domain.OrderStatus, accrual *float64) error
	UpdateBatch(ctx context.Context, orders []domain.OrderData) error
	GetByStatus(ctx context.Context, statuse domain.OrderStatus) ([]domain.OrderData, error)
}

type AcrualSystem interface {
	GetStatus(ctx context.Context, orderNum domain.OrderNumber) (*domain.AccrualData, error)
}

type order struct {
	storage               OrderStorage
	acrualSystem          AcrualSystem
	processingWorkerCount int
	batchSize             int
	once                  sync.Once
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
	logger, err := domain.GetCtxLogger(ctx)
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
	logger, err := domain.GetCtxLogger(ctx)
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
	ordNumChan <-chan *domain.OrderNumber,
	acrualDataChan chan<- *domain.AccrualData) {

	logger := domain.GetMainLogger()

	// Пытаюсь осмыслить https://go.dev/blog/io2013-talk-concurrency
	var ordNumInternalChan <-chan *domain.OrderNumber = ordNumChan
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
			acrData, err := ord.acrualSystem.GetStatus(ctx, *orderNum)
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

func (ord *order) orderStatusUpdater(ctx context.Context, acrualDataChan <-chan *domain.AccrualData) {
	// Будем обновлять данные пачками либо по 10 записей либо через каждые 2 секунды
	var orders []domain.OrderData

	// Пытаюсь осмыслить https://go.dev/blog/io2013-talk-concurrency
	var waitAcrualsTimeoutChan <-chan time.Time = time.After(2 * time.Second)
	var sleepAfterErrChan <-chan time.Time
	var acrualDataInternalChan <-chan *domain.AccrualData = acrualDataChan

	logger := domain.GetMainLogger()

	for {
		select {
		case <-ctx.Done():
			logger.Infow("app.orderStatusUpdater", "status", "complete")
			return
		case <-sleepAfterErrChan:
			waitAcrualsTimeoutChan = time.After(2 * time.Second)
			sleepAfterErrChan = nil
			acrualDataInternalChan = acrualDataChan
		case <-waitAcrualsTimeoutChan:
			if len(orders) > 0 {
				err := ord.storage.UpdateBatch(ctx, orders)
				if err != nil {
					// Произошла ошибка - запускаем sleepChan канал в надежде на восстановление
					logger.Infow("order.UpdateBatch", "err", err.Error())
					sleepAfterErrChan = time.After(5 * time.Second)
					acrualDataInternalChan = nil
				}
				orders = nil
			}
			waitAcrualsTimeoutChan = time.After(2 * time.Second)
		case acrualData, ok := <-acrualDataInternalChan:
			if !ok {
				logger.Infow("app.orderStatusUpdater", "status", "complete")
				return
			}
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

			if len(orders) == ord.batchSize {
				err := ord.storage.UpdateBatch(ctx, orders)
				if err != nil {
					// Произошла ошибка - запускаем sleepChan канал в надежде на восстановление
					logger.Infow("order.UpdateBatch", "err", err.Error())
					sleepAfterErrChan = time.After(5 * time.Second)
					acrualDataInternalChan = nil
				}
				orders = nil
			}
		}
	}
}

// Ищет в хранилище запросы на начисление в статусе OrderStratusNew
func (ord *order) poolOrders(ctx context.Context, ordNumChan chan<- *domain.OrderNumber) {
	var sleepChan <-chan time.Time
	var ordNumChanInternal chan<- *domain.OrderNumber = ordNumChan
	var nexNum *domain.OrderNumber

	logger := domain.GetMainLogger()

Loop:
	for {
		orders, err := ord.storage.GetByStatus(ctx, domain.OrderStratusNew)
		if err != nil {
			logger.Infow("order.poolOrders", "err", err.Error())
			sleepChan = time.After(5 * time.Second)
			ordNumChanInternal = nil
			clear(orders) // для защиты от мусора
		} else {
			if len(orders) == 0 {
				logger.Infow("order.poolOrders", "status", "no record for pool")
				sleepChan = time.After(2 * time.Second)
				ordNumChanInternal = nil
			} else {
				sleepChan = nil
				var num = orders[0].Number
				nexNum = &num
			}
		}

	OrderLoop:
		for {
			select {
			case <-ctx.Done():
				logger.Infow("order.poolOrders", "status", "complete")
				return
			case <-sleepChan:
				sleepChan = nil
				ordNumChanInternal = ordNumChan
				continue Loop // спим либо при ошибке, либо при отстутсвии данных
			case ordNumChanInternal <- nexNum:
				orders = orders[1:]
				if len(orders) == 0 {
					continue Loop
				} else {
					var num = orders[0].Number
					nexNum = &num
					continue OrderLoop
				}

			}
		}
	}
}

func (ord *order) PoolAcrualSystem(ctx context.Context) {
	ord.once.Do(func() {
		go func() {

			ctx, err := domain.EnrichWithMainLogger(ctx)
			if err != nil {
				panic(err)
			}
			var wg sync.WaitGroup
			defer wg.Wait()

			ordNumChan := make(chan *domain.OrderNumber)
			acrualDataChan := make(chan *domain.AccrualData)

			wg.Add(1)
			go func() {
				defer wg.Done()
				ord.poolOrders(ctx, ordNumChan)
			}()

			wg.Add(1)
			go func() {
				defer wg.Done()
				ord.orderStatusUpdater(ctx, acrualDataChan)
			}()

			for i := 0; i < ord.processingWorkerCount; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					ord.refreshOrderStatus(ctx, ordNumChan, acrualDataChan)
				}()
			}

		}()
	})
}
