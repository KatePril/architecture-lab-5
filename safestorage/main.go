package safestorage

type Storage interface {
	Put(key, value string) error
	Get(key string) (string, error)
}

type result struct {
	value string
	err error
}

type command struct {
	action, key, value string
	result chan result
}

type SafeStorage struct {
	Storage Storage
	commands chan command	
}

var cases = map[string] func(Storage, command) result {
	"get": func(storage Storage, cmd command) result {
		value, err := storage.Get(cmd.key)
		return result{ value, err }
	},
	"put": func(storage Storage, cmd command) result {
		err := storage.Put(cmd.key, cmd.value)
		return result{ "", err }
	},
}

func Init(storage Storage) *SafeStorage {
	safeStorage := SafeStorage{ storage, make(chan command) }
	go func ()  {
		for cmd := range safeStorage.commands {
			produce, exists := cases[cmd.action]
			if exists {
				cmd.result <- produce(storage, cmd)
			}
		}
	}()
	return &safeStorage
}

func (safeStorage *SafeStorage) Put(key, value string) error {
	resultChannel := make(chan result)
	cmd := command{ "put", key, value, resultChannel }
	safeStorage.commands <- cmd
	answer := <-resultChannel
	return answer.err
}

func (safeStorage *SafeStorage) Get(key string) (string, error) {
	resultChannel := make(chan result)
	cmd := command{ "get", key, "", resultChannel }
	safeStorage.commands <- cmd
	answer := <-resultChannel
	return answer.value, answer.err
}
