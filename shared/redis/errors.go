package redis

import "errors"

var (
	// ErrKeyNotFound возвращается, когда ключ не найден в Redis
	ErrKeyNotFound = errors.New("key not found")

	// ErrConnectionFailed возвращается при неудачном подключении к Redis
	ErrConnectionFailed = errors.New("connection to Redis failed")

	// ErrInvalidValue возвращается при попытке сериализации/десериализации невалидного значения
	ErrInvalidValue = errors.New("invalid value")

	// ErrLockAcquisitionFailed возвращается при неудачной попытке получить блокировку
	ErrLockAcquisitionFailed = errors.New("failed to acquire lock")

	// ErrLockNotHeld возвращается при попытке освободить блокировку, которая не принадлежит текущему клиенту
	ErrLockNotHeld = errors.New("lock not held")

	// ErrTimeout возвращается при превышении времени ожидания операции
	ErrTimeout = errors.New("operation timeout")

	// ErrTransactionFailed возвращается при неудачной транзакции
	ErrTransactionFailed = errors.New("transaction failed")

	// ErrInvalidConfiguration возвращается при неправильной конфигурации Redis
	ErrInvalidConfiguration = errors.New("invalid Redis configuration")

	// ErrClientClosed возвращается при попытке использовать закрытый клиент
	ErrClientClosed = errors.New("Redis client is closed")

	// ErrInvalidKey возвращается при попытке использовать невалидный ключ
	ErrInvalidKey = errors.New("invalid key")

	// ErrInvalidTTL возвращается при попытке установить невалидное время жизни
	ErrInvalidTTL = errors.New("invalid TTL")

	// ErrPipelineEmpty возвращается при попытке выполнить пустой пайплайн
	ErrPipelineEmpty = errors.New("pipeline is empty")

	// ErrMaxRetriesExceeded возвращается при превышении максимального количества попыток
	ErrMaxRetriesExceeded = errors.New("maximum number of retries exceeded")

	// ErrInvalidPattern возвращается при использовании невалидного шаблона для поиска ключей
	ErrInvalidPattern = errors.New("invalid key pattern")

	// ErrClusterNotSupported возвращается при попытке использовать неподдерживаемую операцию в кластере
	ErrClusterNotSupported = errors.New("operation not supported in cluster mode")

	// ErrScriptNotFound возвращается при попытке выполнить несуществующий Lua скрипт
	ErrScriptNotFound = errors.New("script not found")

	// ErrCircuitBreakerOpen возвращается когда circuit breaker находится в открытом состоянии
	ErrCircuitBreakerOpen = errors.New("circuit breaker is open")
)

// IsNotFound проверяет, является ли ошибка "ключ не найден"
func IsNotFound(err error) bool {
	return errors.Is(err, ErrKeyNotFound)
}

// IsConnectionError проверяет, является ли ошибка ошибкой соединения
func IsConnectionError(err error) bool {
	return errors.Is(err, ErrConnectionFailed)
}

// IsTimeout проверяет, является ли ошибка таймаутом
func IsTimeout(err error) bool {
	return errors.Is(err, ErrTimeout)
}

// IsTransactionError проверяет, является ли ошибка ошибкой транзакции
func IsTransactionError(err error) bool {
	return errors.Is(err, ErrTransactionFailed)
}

// IsLockError проверяет, является ли ошибка ошибкой блокировки
func IsLockError(err error) bool {
	return errors.Is(err, ErrLockAcquisitionFailed) || errors.Is(err, ErrLockNotHeld)
}

// IsConfigurationError проверяет, является ли ошибка ошибкой конфигурации
func IsConfigurationError(err error) bool {
	return errors.Is(err, ErrInvalidConfiguration)
}

// IsClientError проверяет, является ли ошибка ошибкой клиента
func IsClientError(err error) bool {
	return errors.Is(err, ErrClientClosed)
}

// IsValidationError проверяет, является ли ошибка ошибкой валидации
func IsValidationError(err error) bool {
	return errors.Is(err, ErrInvalidValue) ||
		errors.Is(err, ErrInvalidKey) ||
		errors.Is(err, ErrInvalidTTL) ||
		errors.Is(err, ErrInvalidPattern)
}

// IsRetryableError проверяет, можно ли повторить операцию при данной ошибке
func IsRetryableError(err error) bool {
	return IsConnectionError(err) || IsTimeout(err) || errors.Is(err, ErrCircuitBreakerOpen)
}
