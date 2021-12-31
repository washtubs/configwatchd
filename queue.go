package configwatchd

import "sync"

type queue struct {
	contents   []string
	mainConfig MainConfig
	mutex      sync.Mutex
}

func (q *queue) list() []string {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	output := make([]string, len(q.contents))
	copy(output, q.contents)
	return output
}

func (q *queue) executeAll() {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	for _, current := range q.contents {
		execute(current, q.mainConfig)
	}
}

func (q *queue) execute(configKeys []string) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.process(configKeys, false)
}

func (q *queue) clearAll() {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.contents = make([]string, 0, 10)
}

func (q *queue) clear(configKeys []string) {
	q.process(configKeys, true)
}

func (q *queue) process(configKeys []string, clear bool) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	newContents := make([]string, 0, len(q.contents))
	for _, current := range q.contents {
		found := false
		for _, configKey := range configKeys {
			if current == configKey {
				found = true
				if !clear {
					execute(current, q.mainConfig)
				}
				break
			}
		}
		if !found {
			newContents = append(newContents, current)
		}
	}
}

func (q *queue) enqueue(configKey string) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	for _, v := range q.contents {
		if configKey == v {
			return
		}
	}
	q.contents = append(q.contents, configKey)
}

func (q *queue) isEmpty() bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return len(q.contents) == 0
}
