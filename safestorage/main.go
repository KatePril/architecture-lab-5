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

func Init(storage Storage) *SafeStorage {
	safeStorage := SafeStorage{ storage, make(chan command) }
	go func ()  {
		for cmd := range safeStorage.commands {
			if cmd.action == "get" {
				value, err := storage.Get(cmd.key)
				cmd.result <- result{ value, err }
			} else if cmd.action == "put" {
				err := storage.Put(cmd.key, cmd.value)
				cmd.result <- result{ "", err }
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
