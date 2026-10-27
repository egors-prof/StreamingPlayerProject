// Package limiter ограничивает число одновременно обслуживаемых стримов.
//
// Это семафор на буферизованном канале: ёмкость канала — число слотов.
// Занять слот = положить значение в канал, освободить = забрать. Когда все
// слоты заняты, новый запрос ждёт освобождения не дольше wait, а затем
// получает отказ (HTTP 503). Так сервер не открывает тысячи файлов
// и соединений сразу, а нагрузка остаётся предсказуемой.
//
// Чем отличается от worker pool: пул держит N постоянных горутин и очередь
// задач. Здесь каждый HTTP-запрос и так обрабатывается своей горутиной
// (так устроен net/http), нам нужно лишь ограничить, сколько из них
// одновременно отдают аудио. Для этого семафор проще и честнее.
package limiter

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

// ErrBusy — все слоты заняты дольше допустимого ожидания.
var ErrBusy = errors.New("stream limit reached")

type Limiter struct {
	slots    chan struct{}
	wait     time.Duration
	rejected atomic.Int64
}

func New(maxConcurrent int, wait time.Duration) *Limiter {
	if maxConcurrent < 1 {
		maxConcurrent = 1
	}
	return &Limiter{slots: make(chan struct{}, maxConcurrent), wait: wait}
}

// Acquire занимает слот. Вызывающий обязан вызвать release ровно один раз
// (обычно через defer), иначе слот «утечёт» и лимит фактически уменьшится.
func (l *Limiter) Acquire(ctx context.Context) (release func(), err error) {
	// Быстрый путь: свободный слот есть — без таймера.
	select {
	case l.slots <- struct{}{}:
		return l.releaseFunc(), nil
	default:
	}

	timer := time.NewTimer(l.wait)
	defer timer.Stop()
	select {
	case l.slots <- struct{}{}:
		return l.releaseFunc(), nil
	case <-timer.C:
		l.rejected.Add(1)
		return nil, ErrBusy
	case <-ctx.Done(): // клиент ушёл, не дождавшись
		return nil, ctx.Err()
	}
}

func (l *Limiter) releaseFunc() func() {
	var once atomic.Bool
	return func() {
		if once.CompareAndSwap(false, true) { // защита от двойного освобождения
			<-l.slots
		}
	}
}

// Stats — текущее состояние для эндпоинта /stream/stats.
type Stats struct {
	Active   int   `json:"active_streams"`
	Capacity int   `json:"capacity"`
	Rejected int64 `json:"rejected_total"`
}

func (l *Limiter) Stats() Stats {
	return Stats{Active: len(l.slots), Capacity: cap(l.slots), Rejected: l.rejected.Load()}
}
